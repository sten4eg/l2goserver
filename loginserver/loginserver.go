package loginserver

import (
	"fmt"
	"github.com/jackc/pgx"
	"l2goserver/config"
	"l2goserver/data/accounts"
	"l2goserver/loginserver/clientpackets"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/serverpackets"
	"log"
	"net"
)

type charactersAccount struct {
	account string
	count   int
}

type LoginServer struct {
	clients     []*models.Client
	gameservers []*models.GameServer
	database    *pgx.Conn
	config      config.Config

	clientsListener net.Listener
}

func New(cfg config.Config) *LoginServer {
	return &LoginServer{config: cfg}
}

func (l *LoginServer) Init() {
	var err error
	dbConfig := pgx.ConnConfig{
		Host:              l.config.LoginServer.Database.Host,
		Port:              l.config.LoginServer.Database.Port,
		Database:          l.config.LoginServer.Database.Name,
		User:              l.config.LoginServer.Database.User,
		Password:          l.config.LoginServer.Database.Password,
		TLSConfig:         nil,
		FallbackTLSConfig: nil,
	}
	l.database, err = pgx.Connect(dbConfig)
	//toDO need ping to table
	if err != nil {
		log.Fatal("Failed to connect to database: ", err.Error())
	} else {
		log.Println("Successful login database connection")
	}

	var ConnGS []*pgx.Conn
	for i, gameserver := range l.config.GameServers {
		dbConfig = pgx.ConnConfig{
			Host:              gameserver.Database.Host,
			Port:              gameserver.Database.Port,
			Database:          gameserver.Database.Name,
			User:              gameserver.Database.User,
			Password:          gameserver.Database.Password,
			TLSConfig:         nil,
			FallbackTLSConfig: nil,
		}
		s, err := pgx.Connect(dbConfig)
		if err != nil {
			log.Fatal("Failed to connect to database: ", err.Error())
		} else {
			log.Printf("Successful gameserver database connection #%d", i)
		}
		ConnGS = append(ConnGS, s)
	}

	accounts.Get(l.database, ConnGS)

	// Select the appropriate database

	// Listen for client connections
	l.clientsListener, err = net.Listen("tcp", ":2106")
	if err != nil {
		log.Fatal("Failed to connect to port 2106:", err.Error())
	} else {
		log.Println("Login server is listening on port 2106")
	}

	// Listen for game servers connections
}

func (l *LoginServer) Start() {
	defer l.database.Close()
	defer l.clientsListener.Close()

	for {
		var err error
		client := models.NewClient()
		client.Socket, err = l.clientsListener.Accept()
		l.clients = append(l.clients, client)
		if err != nil {
			log.Println("Couldn't accept the incoming connection.")
			continue
		} else {
			go l.handleClientPackets(client)
		}
	}

}
func (l *LoginServer) kickClient(client *models.Client) {
	err := client.Socket.Close()
	if err != nil {
		log.Fatal(err)
	}
	for i, item := range l.clients {
		if item.SessionID == client.SessionID {
			copy(l.clients[i:], l.clients[i+1:])
			l.clients[len(l.clients)-1] = nil
			l.clients = l.clients[:len(l.clients)-1]
			break
		}
	}
	log.Println("The client has been successfully kicked from the server.")
}

func (l *LoginServer) handleGameServerPackets(gameserver *models.GameServer) {
	defer gameserver.Socket.Close()

	for {
		opcode, _, err := gameserver.Receive()

		if err != nil {
			log.Println(err)
			log.Println("Closing the connection...")
			break
		}

		switch opcode {
		case 00:
			log.Println("A game server sent a request to register")
		default:
			log.Println("Can't recognize the packet sent by the gameserver")
		}
	}
}

func (l *LoginServer) handleClientPackets(client *models.Client) {

	log.Println("Client tried to connect")
	defer l.kickClient(client)

	crypt.IsStatic = true // todo костыль?
	initPacket := serverpackets.NewInitPacket(*client)

	err := client.Send(initPacket)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println("Init packet send")
	}

	for {
		opcode, data, err := client.Receive()
		if err != nil {
			log.Println(err)
			log.Println("Closing a connection")
			break
		}
		log.Println("Опкод", opcode)
		switch opcode {
		case 07:
			authGameGuard := clientpackets.NewAuthGameGuard(data, client.SessionID)
			buffer := serverpackets.Newggauth(authGameGuard)
			err = client.Send(buffer)
			if err != nil {
				log.Println(err)
			}
		case 00:
			requestAuthLogin, err := clientpackets.NewRequestAuthLogin(data, client, l.database, l.clients)
			var loginResult []byte
			if err != nil {
				if l.config.LoginServer.AutoCreate {
					log.Println("Авторегистрация нового аккаунта")
					err = clientpackets.CreateAccount(data, client, l.database)
					if err != nil {
						loginResult = serverpackets.NewLoginFailPacket(requestAuthLogin)
					} else {
						loginResult = serverpackets.NewLoginOkPacket(client)
					}
				} else {
					loginResult = serverpackets.NewLoginFailPacket(requestAuthLogin)
				}
			} else {
				loginResult = serverpackets.NewLoginOkPacket(client)
			}

			err = client.Send(loginResult)
			if err != nil {
				log.Println(err)
				return
			}
		case 02:
			requestPlay := clientpackets.NewRequestPlay(data)
			_ = requestPlay
			x := serverpackets.NewPlayOkPacket(client)
			err = client.Send(x)
			if err != nil {
				log.Println(err)
				return
			}
		case 05:
			requestServerList := serverpackets.NewServerListPacket(client, l.config.GameServers, client.Socket.RemoteAddr().String())
			err := client.Send(requestServerList)
			if err != nil {
				log.Println(err)
				return
			}
		default:
			log.Println("Unable to determine package type")
			fmt.Printf("opcode: %X", opcode)
		}
	}
}
