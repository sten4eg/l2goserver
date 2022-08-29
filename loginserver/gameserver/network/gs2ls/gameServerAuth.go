package gs2ls

import (
	"errors"
	"l2goserver/config"
	"l2goserver/loginserver/gameserver/network/ls2gs"
	"l2goserver/loginserver/types/reason/loginServer"
	"l2goserver/loginserver/types/state/gameServer"
	"l2goserver/packets"
	"l2goserver/utils"
	"log"
	"net/netip"
)

type gsInterfaceForGameServerAuth interface {
	ForceClose(reason loginServer.FailReason)
	Send(*packets.Buffer) error
	SetInfoGameServerInfo([]netip.Prefix, []byte, byte, int16, int32, bool)
	GetId() byte
	SetState(serverState gameServer.GameServerState)
	GetGsiById(byte) GsiIsAuthInterface
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
	if !utils.Contains(config.GetAllowedServerVersion(), data.serverVersion) {
		server.ForceClose(loginServer.InvalidGameServerVersion)
		return false
	}

	if utils.CompareHexId(data.hexId, config.GetGameServerHexId()) {
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
