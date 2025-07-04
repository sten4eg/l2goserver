package gameserver

import (
	"strings"
)

func GetGameServerInstance() *Table {
	return gameServerInstance
}

func IsAccountInGameServer(account string) bool {
	gameServer := GetGameServerInstance()

	for _, v := range gameServer.gameServersInfo {
		if v.HasAccountOnGameServer(account) {
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
	gsi := gameServer.gameServersInfo[id] // todo need check
	addr := gsi.getGameServerConn().RemoteAddr().String()
	b, _, _ := strings.Cut(addr, ":")
	return b
}

func GetGameServerPort(id int) int16 {
	return GetGameServerInstance().gameServersInfo[id].getPort()
}

func GetGameServerId(index int) byte {
	return GetGameServerInstance().gameServersInfo[index].GetId() // возможна паника если в массиве нету id
}

func GetGameServerMaxPlayers(id int) int32 {
	return GetGameServerInstance().gameServersInfo[id].GetMaxPlayer()
}

func GetGameServerAgeLimit(id int) int32 {
	return GetGameServerInstance().gameServersInfo[id].getAgeLimit()
}

func GetGameServerServerType(id int) int32 {
	return GetGameServerInstance().gameServersInfo[id].GetType()
}

func GetGameServerStatus(id int) byte {
	return byte(GetGameServerInstance().gameServersInfo[id].GetStatus())
}

func ShowBracketsInGameServer(id int) byte {
	if GetGameServerInstance().gameServersInfo[id].getShowBracket() {
		return 1
	}
	return 0
}
