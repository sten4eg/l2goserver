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
	serversCount := gameserver.GetCountGameServer()
	buffer.WriteSingleByte(0x04)
	buffer.WriteSingleByte(serversCount)     // количество серверов
	buffer.WriteSingleByte(byte(lastServer)) // последний выбранный сервер
	//	network, _, _ := net.SplitHostPort(remoteAddr)

	for i := 0; i < int(serversCount); i++ {

		qq := gameserver.GetGameServerIp(i)
		ip := net.ParseIP(qq).To4()

		buffer.WriteSingleByte(gameserver.GetGameServerId(i)) // Server ID (Bartz)
		buffer.WriteSingleByte(ip[0])                         // Server IP address 1/4
		buffer.WriteSingleByte(ip[1])                         // Server IP address 2/4
		buffer.WriteSingleByte(ip[2])                         // Server IP address 3/4
		buffer.WriteSingleByte(ip[3])                         // Server IP address 4/4

		buffer.WriteDU(uint32(gameserver.GetGameServerPort(i)))           // GameServer port number
		buffer.WriteSingleByte(byte(gameserver.GetGameServerAgeLimit(i))) // Age Limit 0, 15, 18
		buffer.WriteSingleByte(0x01)                                      // Is pvp allowed?
		buffer.WriteH(100)                                                // How many players are online Unused In client
		buffer.WriteHU(uint16(gameserver.GetGameServerMaxPlayers(i)))     // Maximum allowed players

		var realStatus byte
		status := gameserver.GetGameServerStatus(i)
		if state.GameServerStatus(status) == state.StatusDown {
			realStatus = 0x00
		} else {
			realStatus = 0x01
		}

		buffer.WriteSingleByte(realStatus)
		buffer.WriteD(gameserver.GetGameServerServerType(i))           // Display a green clock (what is this for?)// Server Type  1: Normal, 2: Relax, 4: Public Test, 8: No Label, 16: Character Creation Restricted, 32: Event, 64: Free
		buffer.WriteSingleByte(gameserver.ShowBracketsInGameServer(i)) // bracket [NULL]Bartz

	}

	buffer.WriteH(0x00) // unknown

	buffer.WriteSingleByte(1) //
	for servId := 0; servId < int(serversCount); servId++ {
		realServerId := gameserver.ConvertIndexToServerId(servId)
		buffer.WriteSingleByte(gameserver.GetGameServerId(servId))
		buffer.WriteSingleByte(client.Account.CharacterCount[realServerId]) //todo тут не так
		buffer.WriteSingleByte(0)                                           // количесвто удаленных чаров
	}
	return client.SendBuf(buffer)

}
