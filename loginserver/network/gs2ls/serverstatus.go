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
	ServerListStatus serverStatusCodes = 1
	ServerType       serverStatusCodes = 2

	ServerListSquareBracket serverStatusCodes = 3
	MaxPlayers              serverStatusCodes = 4
	ServerAge               serverStatusCodes = 6
)

func ServerStatus(data []byte, server ServerStatusInterface) {
	packet := packets.NewReader(data)

	size := packet.ReadInt32()

	for i := 0; i < int(size); i++ {
		code := packet.ReadInt32()
		value := packet.ReadInt32()
		switch code {
		case ServerListStatus:
			server.SetStatus(value)
		case ServerType:
			server.SetServerType(value)
		case ServerListSquareBracket:
			server.SetShowBracket(value == 1)
		case MaxPlayers:
			server.SetMaxPlayer(value)
		case ServerAge:
			server.SetAgeLimit(value)
		}
	}
}
