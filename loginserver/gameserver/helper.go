package gameserver

import (
	"strings"
)

func GetGameServerInstance() *Table {
	return gameServerInstance
}

func (t *Table) IsAccountInGameServer(account string) bool {
	for _, v := range t.gameServersInfo {
		if v.HasAccountOnGameServer(account) {
			return true
		}
	}
	return false
}

func (t *Table) GetCountGameServer() byte {
	return byte(len(t.gameServersInfo))
}

func (t *Table) GetGameServerIp(id int) string {
	gsi := t.gameServersInfo[id] // todo надо проверка
	addr := gsi.getGameServerConn().RemoteAddr().String()
	b, _, _ := strings.Cut(addr, ":")
	return b
}

func (t *Table) GetGameServerPort(id int) int16 {
	return t.gameServersInfo[id].getPort()
}

func (t *Table) GetGameServerId(index int) byte {
	return t.gameServersInfo[index].GetId() // возможна паника если в массиве нету id
}

func (t *Table) GetGameServerMaxPlayers(id int) int32 {
	return t.gameServersInfo[id].GetMaxPlayer()
}

func (t *Table) GetGameServerAgeLimit(id int) int32 {
	return t.gameServersInfo[id].getAgeLimit()
}

func (t *Table) GetGameServerServerType(id int) int32 {
	return t.gameServersInfo[id].GetType()
}

func (t *Table) GetGameServerStatus(id int) byte {
	return byte(t.gameServersInfo[id].GetStatus())
}

func (t *Table) ShowBracketsInGameServer(id int) byte {
	if t.gameServersInfo[id].getShowBracket() {
		return 1
	}
	return 0
}
