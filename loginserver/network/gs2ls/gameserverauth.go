package gs2ls

import (
	"bytes"
	"l2goserver/config"
	"l2goserver/loginserver/network/ls2gs"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"l2goserver/utils"
	"strings"
)

type gsInterfaceForGameServerAuth interface {
	ForceClose(state.LoginServerFail)
	Send(*packets.Buffer)
	SetInfoGameServerInfo(string, []byte, byte, int16, int32, bool)
	GetServerInfoId() byte
	SetState(state.GameServerState)
}

type gameServerAuthData struct {
	hexId               []byte
	hosts               string
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
	size := packet.ReadInt32()
	gsa.hexId = packet.ReadBytes(int(size))
	size = int32(2 * packet.ReadInt16())

	var sb strings.Builder
	for i := 0; i < int(size); i++ {
		sb.Write(utils.S2b(packet.ReadString()))
	}
	gsa.hosts = sb.String()

	if handleRegProcess(server, gsa) {
		server.Send(ls2gs.AuthedResponse(server.GetServerInfoId()))
		server.SetState(state.AUTHED)
	}

}

func handleRegProcess(server gsInterfaceForGameServerAuth, data gameServerAuthData) bool {
	if !utils.Contains(config.GetAllowedServerVersion(), data.serverVersion) {
		server.ForceClose(state.ReasonInvalidGameServerVersion)
		return false
	}

	if bytes.Equal(data.hexId, config.GetGameServerHexId()) {
		server.SetInfoGameServerInfo(data.hosts, data.hexId, data.desiredId, data.port, data.maxPlayers, true)
	} else {
		server.ForceClose(state.ReasonAlreadyLoggedIn)
		return false
	}
	return true
}
