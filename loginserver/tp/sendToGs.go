package tp

import (
	"l2goserver/loginserver/gameserver"
	"l2goserver/loginserver/network/ls2gs"
	"strings"
)

func SendRequestCharacters(account string) {
	gameServer := gameserver.GetGameServerInstance()
	gameServer.Send(ls2gs.RequestCharacter(account))
}

func IsAccountInGameServer(account string) bool {
	gameServer := gameserver.GetGameServerInstance()
	return gameServer.HasAccountOnGameServer(account)
}

func GetGameServerIp() string {
	gameServer := gameserver.GetGameServerInstance()
	addr := gameServer.GetGameServerConn().RemoteAddr().String()
	b, _, _ := strings.Cut(addr, ":")
	return b

}

func GetGameServerPort() int16 {
	return gameserver.GetGameServerInstance().GetGameServerInfoPort()
}

func GetGameServerId() byte {
	return gameserver.GetGameServerInstance().GetGameServerInfoId()

}

func GetGameServerMaxPlayers() int32 {
	return gameserver.GetGameServerInstance().GetGameServerInfoMaxPlayer()
}

func GetGameServerAgeLimit() int32 {
	return gameserver.GetGameServerInstance().GetGameServerInfoAgeLimit()
}
func GetGameServerServerType() int32 {
	return gameserver.GetGameServerInstance().GetGameServerInfoType()

}
func GetGameServerStatus() byte {
	return byte(gameserver.GetGameServerInstance().GetGameServerInfoStatus())
}

func ShowBracketsInGameServer() byte {
	if gameserver.GetGameServerInstance().GetGameServerInfoShowBracket() {
		return 1
	}
	return 0
}
