package loginserver

import (
	"context"
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
	floodProtecor "github.com/sten4eg/floodProtector"
	"l2goserver/database"
	"l2goserver/ipManager"
	"l2goserver/loginserver/gameserver"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/network/c2ls"
	"l2goserver/loginserver/network/ls2c"
	"l2goserver/loginserver/types/gameServerStatuses"
	"l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/loginserver/types/state/clientState"
	"l2goserver/utils"
	"log"
	"math/rand"
	"net"
	"net/netip"
	"sync/atomic"
)

type account = string

type LoginServer struct {
	clientsListener *net.TCPListener
	accounts        *xsync.MapOf[account, *models.ClientCtx]
	gameServerTable *gameserver.Table
	database        database.Database
}

var Atom atomic.Int64
var AtomKick atomic.Int64

func New(db database.Database) (*LoginServer, error) {
	addr := new(net.TCPAddr)
	addr.Port = 2106
	addr.IP = net.IP{127, 0, 0, 1}
	clientsListener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		return nil, err
	}

	gs := gameserver.GetGameServerInstance()

	login := &LoginServer{
		accounts:        xsync.NewMapOf[account, *models.ClientCtx](),
		clientsListener: clientsListener,
		gameServerTable: gs,
		database:        db,
	}
	gs.AttachLS(login)

	return login, nil
}

func (ls *LoginServer) Run() {
	defer ls.clientsListener.Close()

	floodMap := xsync.NewMapOf[string, floodProtecor.ConnectionInfo]()
	for {
		var err error

		client, err := models.NewClient()
		if err != nil {
			log.Println("Не создан клиент", err)
			continue
		}

		conn, err := ls.clientsListener.AcceptTCP()
		if err != nil {
			log.Println(err)
		}

		tcpConn, err := floodProtecor.AcceptTCP(conn, floodMap)
		if err != nil {
			log.Println("Accept() error", err)
			continue
		}

		client.SetConn(tcpConn)
		client.SetState(clientState.Connected)

		go ls.handleClientPackets(client)
	}

}

func (ls *LoginServer) handleClientPackets(client *models.ClientCtx) {
	defer client.CloseConnection()

	client.SendInit(ls2c.NewInitPacket(client))

	for {
		opcode, data, err := client.Receive()
		fmt.Println(opcode)
		Atom.Add(1)
		if err != nil {
			ls.ClientDisconnect(client)

			//	log.Println(err)
			//	log.Println("Closing a connection")
			AtomKick.Add(1)
			return
		}
		//		log.Println("Опкод", opcode)
		state := client.GetState()
		switch state {
		default:
			//			log.Println("Неопознаный опкод")
			//			fmt.Printf("opcode: %X, state %X", opcode, clientState.State)
			return

		case clientState.Connected:
			clientAddrPort := netip.MustParseAddrPort(client.GetRemoteAddr().String())
			if ipManager.IsBannedIp(clientAddrPort.String()) {
				err := client.Send(ls2c.AccountKicked(clientReasons.PermanentlyBanned))
				if err != nil {
					log.Println(err)
				}
				return
			}

			if opcode == 7 {
				err := c2ls.NewAuthGameGuard(data, client)
				if err != nil {
					//	log.Println(err)
					return
				}
				client.SetState(clientState.AuthedGameGuard)
			} else {
				//	log.Println(opcode, clientState.State)
				return
			}

		case clientState.AuthedGameGuard:
			if opcode == 0 {
				err = c2ls.NewRequestAuthLogin(data, client, ls, gameserver.GetGameServerInstance(), ls.database)
				if err != nil {
					//	log.Println(err)
					return
				}
			} else {
				//	log.Println(opcode, clientState.State)
				return
			}
		case clientState.AuthedLogin:
			switch opcode {
			default:
				//	log.Println("Неопознаный опкод")
				//	fmt.Printf("opcode: %X, state %X", opcode, clientState.State)
				return
			case 02:
				err = c2ls.RequestServerLogin(data, client, ls)
				if err != nil {
					//	log.Println(err)
					return
				}
			case 05:
				client.Send(ls2c.NewServerListPacket(client, gameserver.GetGameServerInstance()))
				return
			}
		}
	}
}

func (ls *LoginServer) GetSessionKey(account string) (uint32, uint32, uint32, uint32) {
	ctx, ok := ls.accounts.Load(account)
	if !ok {
		return 0, 0, 0, 0
	}
	return ctx.SessionKey.LoginOk1, ctx.SessionKey.LoginOk2, ctx.SessionKey.PlayOk1, ctx.SessionKey.PlayOk2
}

func (ls *LoginServer) IsAccountInLoginAndAddIfNot(client c2ls.ClientRequestInterface) bool {
	_, inLogin := ls.accounts.LoadOrStore(client.GetAccountLogin(), nil) //todo nil
	return inLogin
}

func (ls *LoginServer) AssignSessionKeyToClient(client c2ls.GAL) (uint32, uint32, uint32, uint32) {
	ls.accounts.Store(client.GetAccountLogin(), nil) //todo nil
	return rand.Uint32(), rand.Uint32(), rand.Uint32(), rand.Uint32()
}

func (ls *LoginServer) RemoveAuthedLoginClient(account string) {
	ctx, loaded := ls.accounts.LoadAndDelete(account)
	if loaded && ctx != nil {
		ctx.CloseConnection()
	}
}

func (ls *LoginServer) GetAccount(account string) *models.Account {
	ctx, ok := ls.accounts.Load(account)
	if ok && ctx != nil {
		return &ctx.Account
	}
	return nil
}

func (_ *LoginServer) GetGameServerInfoList() []c2ls.GameServerInfoInterface {
	infoList := gameserver.GetGameServerInstance().GetGameServerInfoList()
	return utils.ConvertSlice(infoList, func(info *gameserver.Info) c2ls.GameServerInfoInterface {
		return info
	})
}

func (ls *LoginServer) GetClientCtx(account string) c2ls.ClientRequestInterface {
	ctx, ok := ls.accounts.Load(account)
	if ok {
		return ctx
	}
	return nil
}

func (ls *LoginServer) IsLoginPossible(client c2ls.ClientServerLogin, serverId byte) (bool, error) {
	const AccountLastServerUpdate = `UPDATE accounts SET last_server = $1 WHERE login = $2`

	gsi := gameserver.GetGameServerInstance().GetGameServerById(serverId)
	access := client.GetAccountAccessLevel()
	if gsi != nil && gsi.IsAuthed() {
		loginOk := (gsi.GetCurrentPlayerCount() < gsi.GetMaxPlayer()) && (gsi.GetStatus() != gameServerStatuses.StatusGmOnly || access > 0)
		if loginOk && (client.GetLastServer() != int8(serverId)) {
			_, err := ls.database.Exec(context.Background(), AccountLastServerUpdate, serverId, client.GetAccountLogin())
			if err != nil {
				log.Println(err.Error())
			}
		}
		return loginOk, nil
	}
	return false, nil
}

func (ls *LoginServer) ClientDisconnect(client *models.ClientCtx) {
	if !client.IsJoinedGS() {
		ls.RemoveAuthedLoginClient(client.Account.Login)
	}
}
