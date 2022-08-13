package gameserver

import "l2goserver/loginserver/models"

type LoginServInterface interface {
	GetSessionKey(string) *models.SessionKey
	RemoveAuthedLoginClient(string)
}

func (gs *GS) AttachLS(i LoginServInterface) {
	gs.loginServerInfo = i
}
func (gs *GS) LoginServerGetSessionKey(account string) *models.SessionKey {
	return gs.loginServerInfo.GetSessionKey(account)
}
func (gs *GS) LoginServerRemoveAuthedLoginClient(account string) {
	gs.loginServerInfo.RemoveAuthedLoginClient(account)
}
