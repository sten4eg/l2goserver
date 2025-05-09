package gs2ls

import (
	"bytes"
	"errors"
	"l2goserver/config"
	"l2goserver/loginserver/gameserver/network/ls2gs"
	"l2goserver/loginserver/types/reason/loginServer"
	"l2goserver/loginserver/types/state/gameServer"
	"l2goserver/packets"
	"log"
	"net/netip"
)

type gsInterfaceForGameServerAuth interface {
	ForceClose(reason loginServer.FailReason)
	Send(buffer *packets.Buffer) error
	SetInfoGameServerInfo(host []netip.Prefix, hexId []byte, id byte, port int16, maxPlayer int32, authed bool)
	GetId() byte
	SetState(serverState gameServer.GameServerState)
	GetGsiById(id byte) GsiIsAuthInterface
}

type GsiIsAuthInterface interface {
	IsAuthed() bool
}

type gameServerAuthData struct {
	hexId               []byte
	hosts               []netip.Prefix
	maxPlayers          int32
	port                int16
	serverVersion       byte
	desiredId           byte
	acceptAlternativeId bool
	hostReserved        bool
}

func GameServerAuth(data []byte, server gsInterfaceForGameServerAuth) error {
	packet := packets.NewReader(data)
	var gsa gameServerAuthData
	gsa.serverVersion = packet.ReadSingleByte()
	gsa.desiredId = packet.ReadSingleByte()
	gsa.acceptAlternativeId = packet.ReadSingleByte() != 0
	gsa.hostReserved = packet.ReadSingleByte() != 0
	gsa.port = packet.ReadInt16()
	gsa.maxPlayers = packet.ReadInt32()
	sizeHexId := packet.ReadInt32()
	gsa.hexId = packet.ReadBytes(int(sizeHexId))

	sizeSubNetsAndHosts := packet.ReadInt32()

	var subNets []netip.Prefix

	for i := 0; i < int(sizeSubNetsAndHosts); i++ {
		subNetsAddr, err := netip.ParsePrefix(packet.ReadString())
		if err != nil {
			log.Println(err.Error())
			continue
		}
		_ = packet.ReadString() // TODO Убрать получение дублирующегося пакета с ip без маски подсети
		subNets = append(subNets, subNetsAddr)
	}
	gsa.hosts = subNets
	if handleRegProcess(server, gsa) {
		_ = server.Send(ls2gs.AuthedResponse(server.GetId()))
		server.SetState(gameServer.Authed)
		return nil
	}

	handleReqProcessFailed := errors.New("функция handleReqProcess не выполнена")
	return handleReqProcessFailed
}

func handleRegProcess(server gsInterfaceForGameServerAuth, data gameServerAuthData) bool {
	if bytes.IndexByte(config.GetAllowedServerVersion(), data.serverVersion) == -1 {
		server.ForceClose(loginServer.InvalidGameServerVersion)
		return false
	}

	if compareHexId(data.hexId, config.GetGameServerHexId()) {
		gsi := server.GetGsiById(data.desiredId)
		if gsi != nil {
			if gsi.IsAuthed() {
				server.ForceClose(loginServer.AlreadyLoggedIn)
				return false
			}
		}
		server.SetInfoGameServerInfo(data.hosts, data.hexId, data.desiredId, data.port, data.maxPlayers, true)
	} else {
		server.ForceClose(loginServer.WrongHexId)
		return false
	}
	return true
}

func compareHexId(hexId []byte, hexIds [][]byte) bool {
	for i := range hexIds {
		if bytes.Equal(hexId, hexIds[i]) {
			return true
		}
	}
	return false
}
