package state

type ServerStatusValues = int32

const (
	StatusAuto   ServerStatusValues = 0x00
	StatusGood   ServerStatusValues = 0x01
	StatusNormal ServerStatusValues = 0x02
	StatusFull   ServerStatusValues = 0x03
	StatusDown   ServerStatusValues = 0x04
	StatusGmOnly ServerStatusValues = 0x05
)

type serverStatusCodes = int32

const (
	ServerListStatus serverStatusCodes = 1
	ServerType       serverStatusCodes = 2

	ServerListSquareBracket serverStatusCodes = 3
	MaxPlayers              serverStatusCodes = 4
	ServerAge               serverStatusCodes = 6
)
