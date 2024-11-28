package gameserver

import (
	"l2goserver/config"
	"l2goserver/database"
	"l2goserver/loginserver/gameserver/network/ls2gs"
	"l2goserver/loginserver/network/c2ls"
	"l2goserver/loginserver/types/state/gameServer"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Table struct {
	Connection      *net.TCPListener
	gameServersInfo []*Info
	loginServerInfo LoginServInterface
	database        database.Database
}

var gameServerInstance *Table
var initBlowfishKey = []byte{95, 59, 118, 46, 93, 48, 53, 45, 51, 49, 33, 124, 43, 45, 37, 120, 84, 33, 94, 91, 36, 0}
var uniqId atomic.Uint32

func HandlerInit(db database.Database) error {
	gameServerInstance = new(Table)
	gameServerInstance.database = db

	uniqId.Add(1)
	port := config.GetLoginPortForGameServer()
	intPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	addr := new(net.TCPAddr)
	addr.Port = intPort
	addr.IP = net.IP{127, 0, 0, 1}

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		return err
	}
	gameServerInstance.Connection = listener

	go gameServerInstance.Run()
	return nil
}

func (t *Table) Run() {
	for {
		var err error
		gsi, err := InitGameServerInfo(t.database)
		if err != nil {
			log.Println("ошибка при создании Gsi:", err)
			continue
		}
		gsi.gameServerTable = t
		gsi.uniqId = uniqId.Load()
		uniqId.Add(1)

		gsi.SetBlowFishKey(initBlowfishKey)

		gsi.conn, err = t.Connection.AcceptTCP()
		if err != nil {
			log.Println("ошибка при Accept gameserver")
			continue
		}

		gsi.state = gameServer.Connected

		t.gameServersInfo = append(t.gameServersInfo, gsi)

		pubKey := make([]byte, 1, 65)
		pubKey = append(pubKey, gsi.privateKey.PublicKey.N.Bytes()...)

		buffer := ls2gs.InitLS(pubKey)

		err = gsi.Send(buffer)
		if err != nil {
			log.Println("ошибка при отправке в геймсервера")
			gameServerInstance.RemoveGsi(gsi.uniqId)
			continue
		}

		go gsi.Listen()
	}
}

func (t *Table) RemoveGsi(connId uint32) {
	for i, gsi := range t.gameServersInfo {
		if gsi.uniqId == connId {
			t.gameServersInfo = append(t.gameServersInfo[:i], t.gameServersInfo[i+1:]...)
		}
	}
}

func (t *Table) GetAccountOnGameServer(account string) c2ls.GameServerInfoInterface {
	for _, gsi := range t.GetGameServerInfoList() {
		if gsi.HasAccountOnGameServer(account) {
			return gsi
		}
	}
	return nil
}

func (t *Table) GetGameServerById(serverId byte) *Info {
	for _, gsi := range t.gameServersInfo {
		if gsi.GetId() == serverId {
			return gsi
		}
	}
	return nil
}

func (t *Table) GetGameServerInfoList() []*Info {
	return t.gameServersInfo
}
