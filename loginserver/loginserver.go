package loginserver

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"l2goserver/config"
	"l2goserver/loginserver/clientpackets"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/serverpackets"
	"log"
	"net"
)

type LoginServer struct {
	clients             []*models.Client
	gameservers         []*models.GameServer
	database            *sql.DB
	config              config.Config
	internalServersList []byte
	externalServersList []byte
	status              loginServerStatus

	clientsListener     net.Listener
	gameServersListener net.Listener
}

type loginServerStatus struct {
	successfulAccountCreation uint32
	failedAccountCreation     uint32
	successfulLogins          uint32
	failedLogins              uint32
	hackAttempts              uint32
}

func New(cfg config.Config) *LoginServer {
	return &LoginServer{config: cfg}
}

func (l *LoginServer) Init() {
	var err error

	// Connect to our database
	//	l.database, err = sql.Open("mysql", "root:@/l2jmobiush5")
	if err != nil {

		log.Fatal("Failed to connect to database: ", err.Error())
		log.Fatal(l.config.LoginServer.Database.User + ":" + l.config.LoginServer.Database.Password + "@" + l.config.LoginServer.Database.Host + "/" + l.config.LoginServer.Database.Name)
	} else {
		fmt.Println("Successful database connection")
	}

	// Select the appropriate database

	// Listen for client connections
	l.clientsListener, err = net.Listen("tcp", ":2106")
	if err != nil {
		log.Fatal("Failed to connect to port 2106:", err.Error())
	} else {
		fmt.Println("Login server is listening on port 2106")
	}

	// Listen for game servers connections
	l.gameServersListener, err = net.Listen("tcp", ":9413")
	if err != nil {
		log.Fatal("Failed to connect to port 9413: ", err.Error())
	} else {
		fmt.Println("Login server is listening on port 9413")
	}
}

func (l *LoginServer) Start() {
	defer l.database.Close()
	defer l.clientsListener.Close()
	defer l.gameServersListener.Close()

	done := make(chan bool)

	go func() {
		for {
			var err error
			client := models.NewClient()
			client.Socket, err = l.clientsListener.Accept()
			l.clients = append(l.clients, client)
			if err != nil {
				fmt.Println("Couldn't accept the incoming connection.")
				continue
			} else {
				go l.handleClientPackets(client)
			}
		}
		done <- true
	}()

	go func() {
		for {
			var err error
			gameserver := models.NewGameServer()
			gameserver.Socket, err = l.gameServersListener.Accept()
			l.gameservers = append(l.gameservers, gameserver)
			if err != nil {
				fmt.Println("Couldn't accept the incoming connection.")
				continue
			} else {
				go l.handleGameServerPackets(gameserver)
			}
		}

		done <- true
	}()

	for i := 0; i < 2; i++ {
		<-done
	}

}
func (l *LoginServer) kickClient(client *models.Client) {
	client.Socket.Close()

	for i, item := range l.clients {
		if bytes.Equal(item.SessionID, client.SessionID) {
			copy(l.clients[i:], l.clients[i+1:])
			l.clients[len(l.clients)-1] = nil
			l.clients = l.clients[:len(l.clients)-1]
			break
		}
	}

	fmt.Println("The client has been successfully kicked from the server.")
}

func (l *LoginServer) handleGameServerPackets(gameserver *models.GameServer) {
	defer gameserver.Socket.Close()

	for {
		opcode, _, err := gameserver.Receive()

		if err != nil {
			fmt.Println(err)
			fmt.Println("Closing the connection...")
			break
		}

		switch opcode {
		case 00:
			fmt.Println("A game server sent a request to register")
		default:
			fmt.Println("Can't recognize the packet sent by the gameserver")
		}
	}
}

func (l *LoginServer) handleClientPackets(client *models.Client) {

	fmt.Println("Client tried to connect")
	defer l.kickClient(client)

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		fmt.Println(err)
		return
	}
	client.PrivateKey = privateKey
	client.ScrambleModulus = crypt.ScrambleModulus(privateKey.PublicKey.N.Bytes())

	initPacket := serverpackets.NewInitPacket(*client)

	err = client.Send(initPacket)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Init packet send")
	}

	for {
		opcode, data, err := client.Receive()

		if err != nil {
			fmt.Println(err)
			fmt.Println("Closing a connection")
			break
		}

		switch opcode {
		case 07:
			authGameGuard := clientpackets.NewAuthGameGuard(data, client.SessionID)
			buffer := serverpackets.Newggauth(authGameGuard)
			err = client.Send(buffer)
			if err != nil {
				fmt.Println(err)
			}
			break
		case 00:
			requestAuthLogin := clientpackets.NewRequestAuthLogin(data, *client)
			loginOk := serverpackets.NewLoginOkPacket(client)
			err = client.Send(loginOk)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("User %s is trying to login\n", requestAuthLogin.Username)
		case 02:
			requestPlay := clientpackets.NewRequestPlay(data)
			_ = requestPlay
			x := serverpackets.NewPlayOkPacket(client)
			err = client.Send(x)
			if err != nil {
				log.Fatal(err)
			}
		case 05:
			requestServerList := serverpackets.NewServerListPacket(l.config.GameServers, client.Socket.RemoteAddr().String())

			err := client.Send(requestServerList)
			if err != nil {
				log.Fatal(err)
			}

		default:

			fmt.Println("Unable to determine package type")
			fmt.Printf("opcode: %X", opcode)
		}
	}
}
