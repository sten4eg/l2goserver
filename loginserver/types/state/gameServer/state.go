package gameServer

type GameServerState byte

const (
	Connected   GameServerState = 0
	BfConnected GameServerState = 1
	Authed      GameServerState = 2
)
