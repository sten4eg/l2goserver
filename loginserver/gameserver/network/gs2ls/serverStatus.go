package gs2ls

import (
	"l2goserver/loginserver/types/gameServerStatuses"
	"l2goserver/packets"
)

type ServerStatusInterface interface {
	SetStatus(gameServerStatuses.ServerStatusValues)
	SetShowBracket(bool)
	SetMaxPlayer(int32)
	SetServerType(int32)
	SetAgeLimit(int32)
}

func ServerStatus(data []byte, server ServerStatusInterface) {
	packet := packets.NewReader(data)

	size := packet.ReadInt32()

	for i := 0; i < int(size); i++ {
		code := packet.ReadInt32()
		value := packet.ReadInt32()
		switch code {
		case gameServerStatuses.ServerListStatus:
			server.SetStatus(value)
		case gameServerStatuses.ServerType:
			server.SetServerType(value)
		case gameServerStatuses.ServerListSquareBracket:
			server.SetShowBracket(value == 1)
		case gameServerStatuses.MaxPlayers:
			server.SetMaxPlayer(value)
		case gameServerStatuses.ServerAge:
			server.SetAgeLimit(value)
		}
	}
}
