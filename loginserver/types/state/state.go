package state

type GameState int8

const (
	NoState GameState = iota
	Connected
	AuthedGameGuard
	AuthedLogin
)

type GameServerState byte

const (
	CONNECTED    GameServerState = iota
	BF_CONNECTED GameServerState = iota
	AUTHED       GameServerState = iota
)

type LoginServerFail byte

const (
	REASON_INVALID_GAME_SERVER_VERSION LoginServerFail = 0
)
