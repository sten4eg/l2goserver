package state

type GameState int8

const (
	NoState         GameState = 0
	Connected       GameState = 1
	AuthedGameGuard GameState = 2
	AuthedLogin     GameState = 3
)

type GameServerState byte

const (
	CONNECTED    GameServerState = 0
	BF_CONNECTED GameServerState = 1
	AUTHED       GameServerState = 2
)

type LoginServerFail byte

const (
	REASON_INVALID_GAME_SERVER_VERSION LoginServerFail = 0
	REASON_WRONG_HEXID                 LoginServerFail = 3
	REASON_ALREADY_LOGGED_IN           LoginServerFail = 7
)
