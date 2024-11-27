package gameserver

import "l2goserver/loginserver/models"

type LoginServInterface interface {
	GetSessionKey(string) (uint32, uint32, uint32, uint32)
	RemoveAuthedLoginClient(string)
	GetAccount(string) *models.Account
}

func (t *Table) AttachLS(i LoginServInterface) {
	t.loginServerInfo = i
}

func (t *Table) LoginServerGetSessionKey(account string) (uint32, uint32, uint32, uint32) {
	return t.loginServerInfo.GetSessionKey(account)
}

func (gsi *Info) LoginServerGetSessionKey(account string) (uint32, uint32, uint32, uint32) {
	return gsi.gameServerTable.loginServerInfo.GetSessionKey(account)
}

func (t *Table) LoginServerRemoveAuthedLoginClient(account string) {
	t.loginServerInfo.RemoveAuthedLoginClient(account)
}

func (gsi *Info) LoginServerRemoveAuthedLoginClient(account string) {
	gsi.gameServerTable.loginServerInfo.RemoveAuthedLoginClient(account)
}
