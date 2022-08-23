package gameserver

import "l2goserver/loginserver/models"

type LoginServInterface interface {
	GetSessionKey(string) *models.SessionKey
	RemoveAuthedLoginClient(string)
	GetAccount(string) *models.Account
}

func (gs *GS) AttachLS(i LoginServInterface) {
	gs.loginServerInfo = i
}

func (gs *GS) LoginServerGetSessionKey(account string) *models.SessionKey {
	return gs.loginServerInfo.GetSessionKey(account)
}
func (gsi *GameServerInfo) LoginServerGetSessionKey(account string) *models.SessionKey {
	return gsi.gs.loginServerInfo.GetSessionKey(account)
}

func (gs *GS) LoginServerRemoveAuthedLoginClient(account string) {
	gs.loginServerInfo.RemoveAuthedLoginClient(account)
}

func (gsi *GameServerInfo) LoginServerRemoveAuthedLoginClient(account string) {
	gsi.gs.loginServerInfo.RemoveAuthedLoginClient(account)
}
