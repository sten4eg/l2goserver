package loginserver

import (
	"l2goserver/config"
	"l2goserver/loginserver/clientpackets"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/serverpackets"
	"l2goserver/loginserver/types/state"
	"net"
)

type charactersAccount struct {
	account string
	count   int
}

type LoginServer struct {
	clients         []*models.ClientCtx
	gameservers     []*models.GameServer
	config          config.Conf
	clientsListener net.Listener
}

func New(cfg config.Conf) *LoginServer {
	return &LoginServer{config: cfg}
}

func (l *LoginServer) Init() {
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
		client := models.NewClient()
		client.Socket, err = l.clientsListener.Accept()
		l.clients = append(l.clients, client)
		client.State = state.Connected
		if err != nil {
			//	log.Println("Couldn't accept the incoming connection.")
			continue
		} else {
			go l.handleClientPackets(client)
		}
	}

}
func (l *LoginServer) kickClient(client *models.ClientCtx) {
	err := client.Socket.Close()
	if err != nil {
		//	log.Fatal(err)
	}
	for i, item := range l.clients {
		if item.SessionID == client.SessionID {
			copy(l.clients[i:], l.clients[i+1:])
			l.clients[len(l.clients)-1] = nil
			l.clients = l.clients[:len(l.clients)-1]
			break
		}
	}
	//	log.Println("The client has been successfully kicked from the server.")
}

//func (l *LoginServer) handleGameServerPackets(gameserver *models.GameServer) {
//	defer gameserver.Socket.Close()
//
//	for {
//		opcode, _, err := gameserver.Receive()
//
//		if err != nil {
//			log.Println(err)
//			log.Println("Closing the connection...")
//			break
//		}
//
//		switch opcode {
//		case 00:
//			log.Println("A game server sent a request to register")
//		default:
//			log.Println("Can't recognize the packet sent by the gameserver")
//		}
//	}
//}

func (l *LoginServer) handleClientPackets(client *models.ClientCtx) {
	defer l.kickClient(client)
	var err error

	crypt.IsStatic = true // todo костыль?
	initPacket := serverpackets.NewInitPacket(client)

	err = client.Send(initPacket)
	if err != nil {
		//		log.Println(err)
		return
	} else {
		//		log.Println("Init packet send")
	}

	for {
		opcode, data, err := client.Receive()
		if err != nil {
			//			log.Println(err)
			//			log.Println("Closing a connection")
			break
		}
		//		log.Println("Опкод", opcode)
		switch client.State {
		default:
			//			log.Println("Неопознаный опкод")
			//			fmt.Printf("opcode: %X, state %X", opcode, client.State)
			return

		case state.Connected:
			if opcode == 7 {
				err := clientpackets.NewAuthGameGuard(data, client)
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
				err = clientpackets.NewRequestAuthLogin(data, client, l.clients, l.config.LoginServer.AutoCreate)
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
				err = clientpackets.NewRequestPlay(data, client)
				if err != nil {
					//	log.Println(err)
					return
				}
			case 05:
				requestServerList := serverpackets.NewServerListPacket(client, l.config.GameServers, client.Socket.RemoteAddr().String())
				err := client.Send(requestServerList)
				if err != nil {
					//		log.Println(err)
					return
				}
			}
		}
	}
}
