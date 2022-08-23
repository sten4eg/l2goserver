package ls2c

import (
	"l2goserver/loginserver/gameserver"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"net"
	"time"
)

func NewServerListPacket(client *models.ClientCtx) error {
	buffer := packets.Get()
	lastServer := client.Account.LastServer
	serversCount := gameserver.GetCountGameServer()
	buffer.WriteSingleByte(0x04)
	buffer.WriteSingleByte(serversCount)     // количество серверов
	buffer.WriteSingleByte(byte(lastServer)) // последний выбранный сервер

	for i := 0; i < int(serversCount); i++ {

		gsIp := gameserver.GetGameServerIp(i)
		ip := net.ParseIP(gsIp).To4()

		buffer.WriteSingleByte(gameserver.GetGameServerId(i)) // Server ID (Bartz)
		buffer.WriteSingleByte(ip[0])                         // Server IP address 1/4
		buffer.WriteSingleByte(ip[1])                         // Server IP address 2/4
		buffer.WriteSingleByte(ip[2])                         // Server IP address 3/4
		buffer.WriteSingleByte(ip[3])                         // Server IP address 4/4

		buffer.WriteDU(uint32(gameserver.GetGameServerPort(i)))           // GameServer port number
		buffer.WriteSingleByte(byte(gameserver.GetGameServerAgeLimit(i))) // Age Limit 0, 15, 18
		buffer.WriteSingleByte(0x01)                                      // Is pvp allowed? default True
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

	buffer.WriteSingleByte(serversCount)
	for i := 0; i < int(serversCount); i++ {
		serverId := gameserver.GetGameServerId(i)
		buffer.WriteSingleByte(serverId)
		buffer.WriteSingleByte(client.Account.CharacterCount[serverId])
		charsToDel, ok := client.Account.CharactersToDel[serverId]
		if ok && len(charsToDel) != 0 {
			buffer.WriteSingleByte(byte(len(charsToDel)))
			for j := range charsToDel {
				buffer.WriteD(int32((charsToDel[j] - time.Now().UnixMilli()) / 1000))
			}
		} else {
			buffer.WriteSingleByte(0)
		}
	}
	return client.SendBuf(buffer)

}
