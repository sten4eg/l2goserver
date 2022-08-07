package state

type GameState int8

const (
	NoState GameState = iota
	Connected
	AuthedGameGuard
	AuthedLogin
)
