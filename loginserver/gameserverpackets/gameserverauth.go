package gameserverpackets

import (
	"l2goserver/config"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"l2goserver/utils"
	"strings"
)

type gsInterfaceGSA interface {
	ForceClose(state.LoginServerFail)
	Send(*packets.Buffer)
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

func GameServerAuth(data []byte, server gsInterfaceGSA) {
	packet := packets.NewReader(data)
	_ = packet.ReadSingleByte() // пропускаем опкод

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
		//server.Send()
	}
	var q = 32
	_ = q
}

func handleRegProcess(server gsInterfaceGSA, data gameServerAuthData) bool {
	if !utils.Contains(config.GetAllowedServerVersion(), data.serverVersion) {
		server.ForceClose(state.REASON_INVALID_GAME_SERVER_VERSION)
		return false
	}

	return true
}
