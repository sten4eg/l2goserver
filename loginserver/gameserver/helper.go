package gameserver

import (
	"net"
	"strings"
)

func GetGameServerInstance() *GS {
	return gameServerInstance
}
func (gsi *GameServerInfo) getGameServerConn() net.Conn {
	return gsi.conn
}

func IsAccountInGameServer(account string) bool {
	gameServer := GetGameServerInstance()

	for _, v := range gameServer.gameServersInfo {
		if v.hasAccountOnGameServer(account) {
			return true
		}
	}
	return false
}

func GetCountGameServer() byte {
	gameServer := GetGameServerInstance()
	return byte(len(gameServer.gameServersInfo))
}
func GetGameServerIp(id int) string {
	gameServer := GetGameServerInstance()
	gsi := gameServer.gameServersInfo[id] //todo надо проверка
	addr := gsi.getGameServerConn().RemoteAddr().String()
	b, _, _ := strings.Cut(addr, ":")
	return b
}

func GetGameServerPort(id int) int16 {
	return GetGameServerInstance().gameServersInfo[id].getGameServerInfoPort()
}
func GetGameServerId(id int) byte {
	return GetGameServerInstance().gameServersInfo[id].getGameServerInfoId() //возможна паника если в массиве нету id
}
func GetGameServerMaxPlayers(id int) int32 {
	return GetGameServerInstance().gameServersInfo[id].getGameServerInfoMaxPlayer()
}
func GetGameServerAgeLimit(id int) int32 {
	return GetGameServerInstance().gameServersInfo[id].getGameServerInfoAgeLimit()
}
func GetGameServerServerType(id int) int32 {
	return GetGameServerInstance().gameServersInfo[id].getGameServerInfoType()
}
func GetGameServerStatus(id int) byte {
	return byte(GetGameServerInstance().gameServersInfo[id].getGameServerInfoStatus())
}
func ShowBracketsInGameServer(id int) byte {
	if GetGameServerInstance().gameServersInfo[id].getGameServerInfoShowBracket() {
		return 1
	}
	return 0
}
