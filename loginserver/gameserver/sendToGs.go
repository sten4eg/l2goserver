package gameserver

import (
	"l2goserver/loginserver/network/loginserverpackets"
)

func SendRequestCharacters(account string) {
	gameServer := GetGameServerInstance()
	gameServer.Send(loginserverpackets.RequestCharacter(account))
}

func IsAccountInGameServer(account string) bool {
	gameServer := GetGameServerInstance()
	return gameServer.HasAccountOnGameServer(account)
}
