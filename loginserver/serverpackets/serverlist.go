package serverpackets

import (
	"l2goserver/config"
	"l2goserver/data/accounts"
	"l2goserver/loginserver/models"
	"l2goserver/packets"
	"net"
	"strconv"
	"time"
)

func NewServerListPacket(account *models.Client, gameServers []config.GameServerType, remoteAddr string) []byte {
	lastServer := account.Account.LastServer
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x04)
	buffer.WriteSingleByte(uint8(len(gameServers))) // Servers count
	buffer.WriteSingleByte(byte(lastServer))        // Last Picked Server
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
		buffer.WriteSingleByte(checkConnect(gameserver.InternalIP, gameserver.Port))
		buffer.WriteD(0x40)          // Display a green clock (what is this for?)
		buffer.WriteSingleByte(0x00) // bracket [NULL]Bartz
	}

	buffer.WriteH(0x00)
	buffer.WriteSingleByte(uint8(len(gameServers)))
	for servId, _ := range gameServers {
		buffer.WriteSingleByte(uint8(servId + 1))
		buffer.WriteSingleByte(byte(accounts.CountCharacterInAccount(servId, account.Account.Login)))
		buffer.WriteSingleByte(0)
	}
	return buffer.Bytes()
}

//Проверка соединения с гейм-сервером
//Возращает 0x00 - выключен серв, 0x01 включен
func checkConnect(host string, port int) byte {
	timeout := time.Second / 3
	strPort := strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strPort), timeout)
	if err != nil {
		return 0x00
	}
	//defer conn.Close()
	if conn != nil {

		return 0x01
	}
	return 0x00
}
