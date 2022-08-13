package gameserver

import (
	"net"
	"strings"
)

func GetGameServerInstance() *GS {
	return gameServerInstance
}
func (gs *GS) getGameServerConn() net.Conn {
	return gs.conn
}

func IsAccountInGameServer(account string) bool {
	gameServer := GetGameServerInstance()
	return gameServer.hasAccountOnGameServer(account)
}

func GetGameServerIp() string {
	gameServer := GetGameServerInstance()
	addr := gameServer.getGameServerConn().RemoteAddr().String()
	b, _, _ := strings.Cut(addr, ":")
	return b
}

func GetGameServerPort() int16 {
	return GetGameServerInstance().getGameServerInfoPort()
}
func GetGameServerId() byte {
	return GetGameServerInstance().getGameServerInfoId()
}
func GetGameServerMaxPlayers() int32 {
	return GetGameServerInstance().getGameServerInfoMaxPlayer()
}
func GetGameServerAgeLimit() int32 {
	return GetGameServerInstance().getGameServerInfoAgeLimit()
}
func GetGameServerServerType() int32 {
	return GetGameServerInstance().getGameServerInfoType()
}
func GetGameServerStatus() byte {
	return byte(GetGameServerInstance().getGameServerInfoStatus())
}
func ShowBracketsInGameServer() byte {
	if GetGameServerInstance().getGameServerInfoShowBracket() {
		return 1
	}
	return 0
}
