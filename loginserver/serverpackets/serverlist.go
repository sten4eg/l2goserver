package serverpackets

import (
	"l2goserver/config"
	"l2goserver/packets"
	"net"
)

func NewServerListPacket(gameServers []config.GameServerType, remoteAddr string) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x04)
	buffer.WriteSingleByte(uint8(len(gameServers))) // Servers count
	buffer.WriteSingleByte(0x00)                    // Last Picked Server

	network, _, _ := net.SplitHostPort(remoteAddr)

	// Server Data (Repeat for each server)
	for index, gameserver := range gameServers {
		var ip net.IP
		if network == "127.0.0.1" {
			ip = net.ParseIP(gameserver.InternalIP).To4()
		} else {
			ip = net.ParseIP(gameserver.ExternalIP).To4()
		}

		buffer.WriteSingleByte(uint8(index + 1))     // Server ID (Bartz)
		buffer.WriteSingleByte(ip[0])                // Server IP address 1/4
		buffer.WriteSingleByte(ip[1])                // Server IP address 2/4
		buffer.WriteSingleByte(ip[2])                // Server IP address 3/4
		buffer.WriteSingleByte(ip[3])                // Server IP address 4/4
		buffer.WriteD(uint32(gameserver.Port))       // GameServer port number
		buffer.WriteSingleByte(0x00)                 // Age Limit 0, 15, 18
		buffer.WriteSingleByte(0x01)                 // Is pvp allowed?
		buffer.WriteH(0)                             // How many players are online Unused In client
		buffer.WriteH(gameserver.Options.MaxPlayers) // Maximum allowed players
		if gameserver.Options.Testing == true {      // Is this a testing server? (Status Up or Down)
			buffer.WriteSingleByte(0x00)
		} else {
			buffer.WriteSingleByte(0x01)
		}
		buffer.WriteD(0x02)          // Display a green clock (what is this for?)
		buffer.WriteSingleByte(0x00) //bracket [NULL]Bartz
	}
	buffer.WriteH(0x00)
	//todo Count characters in servers
	return buffer.Bytes()
}
