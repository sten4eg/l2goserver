package clientState

type ClientCtxState int8

const (
	NoState         ClientCtxState = 0
	Connected       ClientCtxState = 1
	AuthedGameGuard ClientCtxState = 2
	AuthedLogin     ClientCtxState = 3
)

type ClientAuthState byte

const (
	AccountBanned ClientAuthState = iota
	AlreadyOnLs
	AlreadyOnGs
	AuthSuccess
)
