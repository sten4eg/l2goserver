package ls2c

import (
	"l2goserver/loginserver/gameserver"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"net"
)

func NewServerListPacket(client *models.ClientCtx) error {
	buffer := packets.Get()
	lastServer := client.Account.LastServer
	buffer.WriteSingleByte(0x04)
	buffer.WriteSingleByte(1)                // количество серверов
	buffer.WriteSingleByte(byte(lastServer)) // последний выбранный сервер
	//	network, _, _ := net.SplitHostPort(remoteAddr)

	qq := gameserver.GetGameServerIp()
	ip := net.ParseIP(qq).To4()
	ip = []byte{127, 0, 0, 1}
	buffer.WriteSingleByte(gameserver.GetGameServerId()) // Server ID (Bartz)
	buffer.WriteSingleByte(ip[0])                        // Server IP address 1/4
	buffer.WriteSingleByte(ip[1])                        // Server IP address 2/4
	buffer.WriteSingleByte(ip[2])                        // Server IP address 3/4
	buffer.WriteSingleByte(ip[3])                        // Server IP address 4/4

	buffer.WriteDU(uint32(gameserver.GetGameServerPort()))           // GameServer port number
	buffer.WriteSingleByte(byte(gameserver.GetGameServerAgeLimit())) // Age Limit 0, 15, 18
	buffer.WriteSingleByte(0x01)                                     // Is pvp allowed?
	buffer.WriteH(100)                                               // How many players are online Unused In client
	buffer.WriteHU(uint16(gameserver.GetGameServerMaxPlayers()))     // Maximum allowed players

	var realStatus byte
	status := gameserver.GetGameServerStatus()
	if state.GameServerStatus(status) == state.StatusDown {
		realStatus = 0x00
	} else {
		realStatus = 0x01
	}

	buffer.WriteSingleByte(realStatus)
	buffer.WriteD(gameserver.GetGameServerServerType())                 // Display a green clock (what is this for?)// Server Type  1: Normal, 2: Relax, 4: Public Test, 8: No Label, 16: Character Creation Restricted, 32: Event, 64: Free
	buffer.WriteSingleByte(byte(gameserver.ShowBracketsInGameServer())) // bracket [NULL]Bartz

	buffer.WriteH(0x00) // unknown

	buffer.WriteSingleByte(1) //
	//	for servId, _ := range gameServers {
	buffer.WriteSingleByte(gameserver.GetGameServerId())
	buffer.WriteSingleByte(client.Account.CharacterCount)
	buffer.WriteSingleByte(0) // количесвто удаленных чаров
	//	}
	return client.SendBuf(buffer)

}
