package gameserver

import (
	"l2goserver/loginserver/network/ls2gs"
	"sync"
)

type account struct {
	accounts map[string]bool
	mu       sync.Mutex
}
type GameServerInfo struct {
	host        string
	hexId       []byte
	Id          byte
	port        int16
	maxPlayer   int32
	authed      bool
	status      int32
	serverType  int32
	ageLimit    int32
	showBracket bool
	accounts    account
}

func SendRequestCharacters(account string) {
	gameServer := GetGameServerInstance()
	gameServer.Send(ls2gs.RequestCharacter(account))
}

func (gs *GS) getGameServerInfoPort() int16 {
	return gs.gameServersInfo.port
}
func (gs *GS) getGameServerInfoId() byte {
	return gs.gameServersInfo.Id
}
func (gs *GS) getGameServerInfoMaxPlayer() int32 {
	return gs.gameServersInfo.maxPlayer
}
func (gs *GS) getGameServerInfoAgeLimit() int32 {
	return gs.gameServersInfo.ageLimit
}
func (gs *GS) getGameServerInfoType() int32 {
	return gs.gameServersInfo.serverType
}

func (gs *GS) getGameServerInfoStatus() int32 {
	return gs.gameServersInfo.status
}
func (gs *GS) getGameServerInfoShowBracket() bool {
	return gs.gameServersInfo.showBracket
}

func (gs *GS) hasAccountOnGameServer(account string) bool {
	gs.gameServersInfo.accounts.mu.Lock()
	inGame, ok := gs.gameServersInfo.accounts.accounts[account]
	if !ok {
		return false
	}
	gs.gameServersInfo.accounts.mu.Unlock()
	return inGame
}
