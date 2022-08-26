package loginServer

type FailReason byte

const (
	InvalidGameServerVersion FailReason = 0
	IpBanned                 FailReason = 1
	IpReserved               FailReason = 2
	WrongHexId               FailReason = 3
	IdReserved               FailReason = 4
	NoFreeId                 FailReason = 5
	NotAuthed                FailReason = 6
	AlreadyLoggedIn          FailReason = 7
)
