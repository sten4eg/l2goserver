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

type ServerStatusValues = int32

const (
	StatusAuto   ServerStatusValues = 0x00
	StatusGood   ServerStatusValues = 0x01
	StatusNormal ServerStatusValues = 0x02
	StatusFull   ServerStatusValues = 0x03
	StatusDown   ServerStatusValues = 0x04
	StatusGmOnly ServerStatusValues = 0x05
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
