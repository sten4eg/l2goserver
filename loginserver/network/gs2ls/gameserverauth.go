package gs2ls

import (
	"l2goserver/config"
	"l2goserver/loginserver/network/ls2gs"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"l2goserver/utils"
	"log"
	"net/netip"
)

type gsInterfaceForGameServerAuth interface {
	ForceClose(state.LoginServerFail)
	Send(*packets.Buffer) error
	SetInfoGameServerInfo([]netip.Prefix, []byte, byte, int16, int32, bool)
	GetGameServerInfoId() byte
	SetState(state.GameServerState)
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

func GameServerAuth(data []byte, server gsInterfaceForGameServerAuth) {
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
	var hosts []netip.Addr

	for i := 0; i < int(sizeSubNetsAndHosts); i++ {
		subNetsAddr, err := netip.ParsePrefix(packet.ReadString())
		if err != nil {
			log.Println(err.Error())
		}
		hostsPrefix, err := netip.ParseAddr(packet.ReadString())
		if err != nil {
			log.Println(err.Error())
		}
		subNets = append(subNets, subNetsAddr)
		hosts = append(hosts, hostsPrefix)
	}
	gsa.hosts = subNets

	if handleRegProcess(server, gsa) {
		_ = server.Send(ls2gs.AuthedResponse(server.GetGameServerInfoId()))
		server.SetState(state.AUTHED)
	}

}

func handleRegProcess(server gsInterfaceForGameServerAuth, data gameServerAuthData) bool {
	if !utils.Contains(config.GetAllowedServerVersion(), data.serverVersion) {
		server.ForceClose(state.ReasonInvalidGameServerVersion)
		return false
	}

	if utils.CompareHexId(data.hexId, config.GetGameServerHexId()) {
		server.SetInfoGameServerInfo(data.hosts, data.hexId, data.desiredId, data.port, data.maxPlayers, true)
	} else {
		server.ForceClose(state.ReasonAlreadyLoggedIn)
		return false
	}
	return true
}
