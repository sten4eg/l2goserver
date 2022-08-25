package gs2ls

import "l2goserver/packets"

type ServerStatusInterface interface {
	SetStatus(status int32)
	SetShowBracket(showBracket bool)
	SetMaxPlayer(maxPlayer int32)
	SetServerType(serverType int32)
	SetAgeLimit(ageLimit int32)
}
type serverStatusCodes = int32

const (
	SERVER_LIST_STATUS serverStatusCodes = 1
	SERVER_TYPE        serverStatusCodes = 2

	SERVER_LIST_SQUARE_BRACKET serverStatusCodes = 3
	MAX_PLAYERS                serverStatusCodes = 4
	SERVER_AGE                 serverStatusCodes = 6
)

func ServerStatus(data []byte, server ServerStatusInterface) {
	packet := packets.NewReader(data)

	size := packet.ReadInt32()

	for i := 0; i < int(size); i++ {
		code := packet.ReadInt32()
		value := packet.ReadInt32()
		switch code {
		case SERVER_LIST_STATUS:
			server.SetStatus(value)
		case SERVER_TYPE:
			server.SetServerType(value)
		case SERVER_LIST_SQUARE_BRACKET:
			server.SetShowBracket(value == 1)
		case MAX_PLAYERS:
			server.SetMaxPlayer(value)
		case SERVER_AGE:
			server.SetAgeLimit(value)
		}
	}
}
