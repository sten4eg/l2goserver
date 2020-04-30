package loginserver

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"l2goserver/config"
	"l2goserver/loginserver/clientpackets"
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
	l.database, err = sql.Open("mysql", "root:@/l2jmobiush5")
	if err != nil {

		log.Fatal("Не удалось подключиться к базе данных: ", err.Error())
		log.Fatal(l.config.LoginServer.Database.User + ":" + l.config.LoginServer.Database.Password+ "@"+ l.config.LoginServer.Database.Host+"/" + l.config.LoginServer.Database.Name)
	} else {
		fmt.Println("Удачное подключение к базе данных")
	}

	// Select the appropriate database

	// Listen for client connections
	l.clientsListener, err = net.Listen("tcp", ":2106")
	if err != nil {
		log.Fatal("Не удалось подключиться к порту 2106: ", err.Error())
	} else {
		fmt.Println("Логин сервер слушает порт 2106")
	}

	// Listen for game servers connections
	l.gameServersListener, err = net.Listen("tcp", ":9413")
	if err != nil {
		log.Fatal("Не удалось подключиться к порту 9413: ", err.Error())
	} else {
		fmt.Println("Логин сервер слушает порт 9413")
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

	fmt.Println("Клиент попытался подключиться")
	defer l.kickClient(client)

	buffer := serverpackets.NewInitPacket()
	err := client.Send(buffer, false, false)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Пакет Init отправлен")
	}

	for {
		opcode, data, err := client.Receive()

		if err != nil {
			fmt.Println(err)
			fmt.Println("Закрытие соединения")
			break
		}
		switch opcode {
		case 00:
			// response buffer
			var buffer []byte

			requestAuthLogin := clientpackets.NewRequestAuthLogin(data)

			fmt.Printf("User %s is trying to login\n", requestAuthLogin.Username)
			//accounts := l.database.Exec("accounts")
		//	err := accounts.Find(bson.M{"username": requestAuthLogin.Username}).One(&client.Account)

			if err != nil {
				if l.config.LoginServer.AutoCreate == true {
					hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestAuthLogin.Password), 10)
					if err != nil {
						fmt.Println("An error occured while trying to generate the password")
						l.status.failedAccountCreation += 1

						buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_SYSTEM_ERROR)
					} else {
						client.Account = models.Account{
						//	Id:          bson.NewObjectId(),
							Username:    requestAuthLogin.Username,
							Password:    string(hashedPassword),
							AccessLevel: ACCESS_LEVEL_PLAYER}

				//		err = accounts.Insert(&client.Account)
						if err != nil {
							fmt.Printf("Couldn't create an account for the user %s\n", requestAuthLogin.Username)
							l.status.failedAccountCreation += 1

							buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_SYSTEM_ERROR)
						} else {
							fmt.Printf("Account successfully created for the user %s\n", requestAuthLogin.Username)
							l.status.successfulAccountCreation += 1

							buffer = serverpackets.NewLoginOkPacket(client.SessionID)
						}
					}
				} else {
					fmt.Println("Account not found !")
					l.status.failedLogins += 1

					buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_USER_OR_PASS_WRONG)
				}
			} else {
				// Account exists; Is the password ok?
				err = bcrypt.CompareHashAndPassword([]byte(client.Account.Password), []byte(requestAuthLogin.Password))

				if err != nil {
					fmt.Printf("Wrong password for the account %s\n", requestAuthLogin.Username)
					l.status.failedLogins += 1

					buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_USER_OR_PASS_WRONG)
				} else {

					if client.Account.AccessLevel >= ACCESS_LEVEL_PLAYER {
						l.status.successfulLogins += 1

						buffer = serverpackets.NewLoginOkPacket(client.SessionID)
					} else {
						l.status.failedLogins += 1

						buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_ACCESS_FAILED)
					}

				}
			}

			err = client.Send(buffer)

			if err != nil {
				fmt.Println(err)
			}

		case 02:
			requestPlay := clientpackets.NewRequestPlay(data)

			fmt.Printf("The client wants to connect to the server : %d\n", requestPlay.ServerID)

			var buffer []byte
			if len(l.config.GameServers) >= int(requestPlay.ServerID) && (l.config.GameServers[requestPlay.ServerID-1].Options.Testing == false || client.Account.AccessLevel > ACCESS_LEVEL_PLAYER) {
				if !bytes.Equal(client.SessionID[:8], requestPlay.SessionID) {
					l.status.hackAttempts += 1

					buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_ACCESS_FAILED)
				} else {
					buffer = serverpackets.NewPlayOkPacket()
				}
			} else {
				l.status.hackAttempts += 1

				buffer = serverpackets.NewPlayFailPacket(serverpackets.REASON_ACCESS_FAILED)
			}
			err := client.Send(buffer)

			if err != nil {
				fmt.Println(err)
			}

		case 05:
			requestServerList := clientpackets.NewRequestServerList(data)

			var buffer []byte
			if !bytes.Equal(client.SessionID[:8], requestServerList.SessionID) {
				l.status.hackAttempts += 1

				buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_ACCESS_FAILED)
			} else {
				buffer = serverpackets.NewServerListPacket(l.config.GameServers, client.Socket.RemoteAddr().String())
			}
			err := client.Send(buffer)

			if err != nil {
				fmt.Println(err)
			}

		default:
			fmt.Println("Не удалось определить тип пакета")
		}
	}
}