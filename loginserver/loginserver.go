package loginserver

import (
	"l2goserver/config"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/gameserver"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/network/c2ls"
	"l2goserver/loginserver/network/ls2c"
	"l2goserver/loginserver/types/reason"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"math/rand"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
)

type LoginServer struct {
	clients         sync.Map
	config          config.Conf
	clientsListener net.Listener
	mu              sync.Mutex
	accounts        map[string]*models.ClientCtx //TODO Ну шо опять мютекс
}

var Atom atomic.Int64

func New(cfg config.Conf) *LoginServer {
	login := &LoginServer{config: cfg, accounts: make(map[string]*models.ClientCtx, 1000)}
	gs := gameserver.GetGameServerInstance()
	gs.AttachLS(login)
	return login
}

func (l *LoginServer) GetSessionKey(account string) *models.SessionKey {
	l.mu.Lock()
	q := l.accounts[account]
	l.mu.Unlock()
	if q == nil {
		return nil
	}
	return q.SessionKey
}
func (l *LoginServer) IsLoginServer() bool {
	return true
}
func (l *LoginServer) IsAccountInLoginAndAddIfNot(client *models.ClientCtx) bool {
	inLogin, ok := l.accounts[client.Account.Login]
	if !ok {
		l.accounts[client.Account.Login] = client
		return false
	}
	if nil == inLogin {
		l.accounts[client.Account.Login] = client
		return false
	}
	return true
}

func (l *LoginServer) AssignSessionKeyToClient(client *models.ClientCtx) *models.SessionKey {
	sessionKey := new(models.SessionKey)

	sessionKey.PlayOk1 = rand.Uint32()
	sessionKey.PlayOk2 = rand.Uint32()
	sessionKey.LoginOk1 = rand.Uint32()
	sessionKey.LoginOk2 = rand.Uint32()

	l.mu.Lock()
	l.accounts[client.Account.Login] = client
	l.mu.Unlock()
	return sessionKey
}

func (l *LoginServer) RemoveAuthedLoginClient(account string) {
	l.mu.Lock()
	delete(l.accounts, account)
	l.mu.Unlock()
}
func (l *LoginServer) StartListen() {
	var err error

	// Listen for client connections
	l.clientsListener, err = net.Listen("tcp4", ":2106")
	if err != nil {
		//log.Fatal("Failed to connect to port 2106:", err.Error())
	} else {
		//log.Println("Login server is listening on port 2106")
	}

	// Listen for game servers connections
}

func (l *LoginServer) Run() {
	defer l.clientsListener.Close()

	for {
		var err error
		crypt.IsStatic = true // todo костыль?

		client := models.NewClient()
		client.Conn, err = l.clientsListener.Accept()

		clientAddrPort := netip.MustParseAddrPort(client.Conn.LocalAddr().String())

		if !clientAddrPort.IsValid() {
			continue
		}

		if IsBannedIp(clientAddrPort.Addr()) {
			_ = client.SendBuf(ls2c.AccountKicked(reason.PermanentlyBanned))
			l.kickClient(client)
			continue
		}

		l.clients.Store(client.Uid, client)

		client.SetState(state.Connected)
		if err != nil {
			//	log.Println("Couldn't accept the incoming connection.")
			continue
		} else {
			go l.handleClientPackets(client)
		}
	}

}
func (l *LoginServer) kickClient(client *models.ClientCtx) {
	err := client.Conn.Close()
	if err != nil {
		//	log.Fatal(err)
	}
	l.clients.Delete(client.Uid)

	//	log.Println("The client has been successfully kicked from the server.")
}

func (l *LoginServer) handleClientPackets(client *models.ClientCtx) {
	defer l.kickClient(client)
	var err error

	bufToInit := packets.Get()
	initPacket := ls2c.NewInitPacket(client, bufToInit)

	err = client.SendBuf(initPacket)
	if err != nil {
		//		log.Println(err)
		return
	} else {
		//		log.Println("Init packet send")
	}

	for {
		Atom.Add(1)
		opcode, data, err := client.Receive()
		if err != nil {
			//			log.Println(err)
			//			log.Println("Closing a connection")
			break
		}
		//		log.Println("Опкод", opcode)
		switch client.GetState() {
		default:
			//			log.Println("Неопознаный опкод")
			//			fmt.Printf("opcode: %X, state %X", opcode, client.State)
			return

		case state.Connected:
			if opcode == 7 {
				err := c2ls.NewAuthGameGuard(data, client)
				if err != nil {
					//	log.Println(err)
					return
				}
			} else {
				//	log.Println(opcode, client.State)
				return
			}
		case state.AuthedGameGuard:
			if opcode == 0 {
				err = c2ls.NewRequestAuthLogin(data, client, l)
				if err != nil {
					//	log.Println(err)
					return
				}
			} else {
				//	log.Println(opcode, client.State)
				return
			}
		case state.AuthedLogin:
			switch opcode {
			default:
				//	log.Println("Неопознаный опкод")
				//	fmt.Printf("opcode: %X, state %X", opcode, client.State)
				return
			case 02:
				err = c2ls.NewRequestPlay(data, client)
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
