package loginserver

import (
	"l2goserver/config"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/models"
	clientpackets2 "l2goserver/loginserver/network/clientpackets"
	serverpackets2 "l2goserver/loginserver/network/serverpackets"
	"l2goserver/loginserver/types/reason"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
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
	accounts        map[string]bool //TODO Ну шо опять мютекс
}

var Atom atomic.Int64

func New(cfg config.Conf) *LoginServer {
	return &LoginServer{config: cfg, accounts: make(map[string]bool, 1000)}
}

func (l *LoginServer) IsAccountInLoginAndAddIfNot(account string) bool {
	inLogin, ok := l.accounts[account]
	if !ok {
		l.accounts[account] = true
		return false
	}
	if !inLogin {
		l.accounts[account] = true
		return false
	}
	return true
}

func (l *LoginServer) AssignSessionKeyToClient(account string, client *models.ClientCtx) *models.SessionKey {
	sessionKey := new(models.SessionKey)

	l.mu.Lock()
	l.accounts[account] = true
	l.mu.Unlock()
	return sessionKey
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
			_ = client.SendBuf(serverpackets2.AccountKicked(reason.PermanentlyBanned))
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
	initPacket := serverpackets2.NewInitPacket(client, bufToInit)

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
				err := clientpackets2.NewAuthGameGuard(data, client)
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
				err = clientpackets2.NewRequestAuthLogin(data, client, l)
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
				err = clientpackets2.NewRequestPlay(data, client)
				if err != nil {
					//	log.Println(err)
					return
				}
			case 05:
				client.CloseConnection()
				//requestServerList := serverpackets.NewServerListPacket(client, l.config.GameServers, client.conn.RemoteAddr().String())
				//err := client.SendBuf(requestServerList)
				if err != nil {
					//		log.Println(err)
					return
				}
			}
		}
	}
}
