package ls2c

import (
	"l2goserver/loginserver/types/gameServerStatuses"
	"l2goserver/packets"
	"net"
	"time"
)

type serverListInterface interface {
	GetLastServer() int8
	GetAccountCharacterCountOnServerId(serverId uint8) uint8
	GetAccountCharacterToDelCountOnServerId(serverId uint8) ([]int64, bool)
}
type gameserverInter interface {
	GetCountGameServer() byte
	GetGameServerIp(int) string
	GetGameServerId(int) byte
	GetGameServerPort(int) int16
	GetGameServerAgeLimit(int) int32
	GetGameServerMaxPlayers(int) int32
	GetGameServerStatus(int) byte
	GetGameServerServerType(int) int32
	ShowBracketsInGameServer(int) byte
}

func NewServerListPacket(client serverListInterface, gs gameserverInter) []byte {
	buffer := packets.GetBuffer()
	lastServer := client.GetLastServer()

	serversCount := gs.GetCountGameServer()
	buffer.WriteSingleByte(0x04)
	buffer.WriteSingleByte(serversCount)     // количество серверов
	buffer.WriteSingleByte(byte(lastServer)) // последний выбранный сервер

	for i := 0; i < int(serversCount); i++ {

		gsIp := gs.GetGameServerIp(i)
		ip := net.ParseIP(gsIp).To4()

		buffer.WriteSingleByte(gs.GetGameServerId(i)) // Server ID (Bartz)
		buffer.WriteSingleByte(ip[0])                 // Server IP address 1/4
		buffer.WriteSingleByte(ip[1])                 // Server IP address 2/4
		buffer.WriteSingleByte(ip[2])                 // Server IP address 3/4
		buffer.WriteSingleByte(ip[3])                 // Server IP address 4/4

		buffer.WriteDU(uint32(gs.GetGameServerPort(i)))           // GameServer port number
		buffer.WriteSingleByte(byte(gs.GetGameServerAgeLimit(i))) // Age Limit 0, 15, 18
		buffer.WriteSingleByte(0x01)                              // Is pvp allowed? default True
		buffer.WriteH(100)                                        // How many players are online Unused In clientState
		buffer.WriteHU(uint16(gs.GetGameServerMaxPlayers(i)))     // Maximum allowed players

		var realStatus byte
		status := gs.GetGameServerStatus(i)
		if gameServerStatuses.ServerStatusValues(status) == gameServerStatuses.StatusDown {
			realStatus = 0x00
		} else {
			realStatus = 0x01
		}

		buffer.WriteSingleByte(realStatus)
		buffer.WriteD(gs.GetGameServerServerType(i))           // Display a green clock (what is this for?)// Server Type  1: Normal, 2: Relax, 4: Public Test, 8: No Label, 16: Character Creation Restricted, 32: Event, 64: Free
		buffer.WriteSingleByte(gs.ShowBracketsInGameServer(i)) // bracket [NULL]Bartz

	}

	buffer.WriteH(0x00) // unknown

	buffer.WriteSingleByte(serversCount)
	for i := 0; i < int(serversCount); i++ {
		serverId := gs.GetGameServerId(i)
		buffer.WriteSingleByte(serverId)
		buffer.WriteSingleByte(client.GetAccountCharacterCountOnServerId(serverId))
		charsToDel, ok := client.GetAccountCharacterToDelCountOnServerId(serverId)
		if ok && len(charsToDel) != 0 {
			buffer.WriteSingleByte(byte(len(charsToDel)))
			for j := range charsToDel {
				buffer.WriteD(int32((charsToDel[j] - time.Now().UnixMilli()) / 1000))
			}
		} else {
			buffer.WriteSingleByte(0)
		}
	}
	return buffer.CopyBytes()

}
