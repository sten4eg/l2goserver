package loginserver

import (
	"context"
	"fmt"
	"github.com/puzpuzpuz/xsync"
	"l2goserver/db"
	"l2goserver/loginserver/IpManager"
	"l2goserver/loginserver/gameserver"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/network/c2ls"
	"l2goserver/loginserver/network/ls2c"
	"l2goserver/loginserver/types/gameServerStatuses"
	"l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/loginserver/types/state/clientState"
	"log"
	"math/rand"
	"net"
	"net/netip"
	"sync/atomic"
)

const AccountLastServerUpdate = "UPDATE accounts SET last_server = $1 WHERE login = $2"

type LoginServer struct {
	clientsListener *net.TCPListener
	accounts        *xsync.MapOf[*models.ClientCtx]
}

var Atom atomic.Int64
var AtomKick atomic.Int64

func New() *LoginServer {
	login := &LoginServer{}
	login.accounts = xsync.NewMapOf[*models.ClientCtx]()
	gs := gameserver.GetGameServerInstance()
	gs.AttachLS(login)
	return login
}

func (ls *LoginServer) StartListen() error {
	var err error
	addr := new(net.TCPAddr)
	addr.Port = 2106
	addr.IP = net.IP{127, 0, 0, 1}

	ls.clientsListener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		return err
	}

	return nil
}

func (ls *LoginServer) Run() {
	defer ls.clientsListener.Close()

	for {
		var err error

		client, err := models.NewClient()
		if err != nil {
			log.Println("Не создан клиент", err)
			continue
		}

		tcpConn, err := ls.AcceptTCPWithFloodProtection()
		if err != nil {
			log.Println("Accept() error", err)
			continue
		}

		client.SetConn(tcpConn)

		clientAddrPort := netip.MustParseAddrPort(client.GetRemoteAddr().String())

		if !clientAddrPort.IsValid() {
			continue
		}

		if IpManager.IsBannedIp(clientAddrPort.Addr()) {
			_ = client.SendBuf(ls2c.AccountKicked(clientReasons.PermanentlyBanned))
			client.CloseConnection()
			continue
		}

		client.SetState(clientState.Connected)

		go ls.handleClientPackets(client)

	}

}

func (ls *LoginServer) handleClientPackets(client *models.ClientCtx) {
	defer client.CloseConnection()

	err := ls2c.NewInitPacket(client)
	if err != nil {
		return
	}

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
			if opcode == 7 {
				err := c2ls.NewAuthGameGuard(data, client)
				if err != nil {
					//	log.Println(err)
					return
				}
			} else {
				//	log.Println(opcode, clientState.State)
				return
			}
		case clientState.AuthedGameGuard:
			if opcode == 0 {
				err = c2ls.NewRequestAuthLogin(data, client, ls)
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
				err = ls2c.NewServerListPacket(client)
				if err != nil {
					//		log.Println(err)
					return
				}
			}
		}
	}
}

func (ls *LoginServer) GetSessionKey(account string) *models.SessionKey {
	ctx, ok := ls.accounts.Load(account)
	if !ok {
		return nil
	}
	return ctx.SessionKey
}

func (ls *LoginServer) IsAccountInLoginAndAddIfNot(client *models.ClientCtx) bool {
	_, inLogin := ls.accounts.LoadOrStore(client.Account.Login, client)
	return inLogin
}

func (ls *LoginServer) AssignSessionKeyToClient(client *models.ClientCtx) *models.SessionKey {
	sessionKey := new(models.SessionKey)

	sessionKey.PlayOk1 = rand.Uint32()
	sessionKey.PlayOk2 = rand.Uint32()
	sessionKey.LoginOk1 = rand.Uint32()
	sessionKey.LoginOk2 = rand.Uint32()

	ls.accounts.Store(client.Account.Login, client)
	return sessionKey
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

func (_ *LoginServer) GetGameServerInfoList() []*gameserver.Info {
	return gameserver.GetGameServerInstance().GetGameServerInfoList()
}

func (ls *LoginServer) GetClientCtx(account string) *models.ClientCtx {
	ctx, ok := ls.accounts.Load(account)
	if ok {
		return ctx
	}
	return nil
}

func (_ *LoginServer) IsLoginPossible(client *models.ClientCtx, serverId byte) (bool, error) {
	gsi := gameserver.GetGameServerInstance().GetGameServerById(serverId)
	access := client.Account.AccessLevel
	if gsi != nil && gsi.IsAuthed() {
		loginOk := (gsi.GetCurrentPlayerCount() < gsi.GetMaxPlayer()) && (gsi.GetStatus() != gameServerStatuses.StatusGmOnly || access > 0)
		if loginOk && (client.Account.LastServer != int8(serverId)) {
			dbConn, err := db.GetConn()
			if err != nil {
				return loginOk, err
			}
			defer dbConn.Release()
			_, err = dbConn.Exec(context.Background(), AccountLastServerUpdate, serverId, client.Account.Login)
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
