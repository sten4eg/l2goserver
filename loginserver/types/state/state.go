package state

type ClientCtxState int8

const (
	NoState         ClientCtxState = 0
	Connected       ClientCtxState = 1
	AuthedGameGuard ClientCtxState = 2
	AuthedLogin     ClientCtxState = 3
)

type GameServerState byte

const (
	CONNECTED   GameServerState = 0
	BfConnected GameServerState = 1
	AUTHED      GameServerState = 2
)

type LoginServerFail byte

const (
	ReasonInvalidGameServerVersion LoginServerFail = 0
	ReasonIpBanned                 LoginServerFail = 1
	ReasonIpReserved               LoginServerFail = 2
	ReasonWrongHexId               LoginServerFail = 3
	ReasonIdReserved               LoginServerFail = 4
	ReasonNoFreeId                 LoginServerFail = 5
	NotAuthed                      LoginServerFail = 6
	ReasonAlreadyLoggedIn          LoginServerFail = 7
)
