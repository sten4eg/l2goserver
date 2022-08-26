package gs2ls

import (
	"l2goserver/loginserver/types/state/gameServer"
	"l2goserver/packets"
)

type ServerStatusInterface interface {
	SetStatus(gameServer.ServerStatusValues)
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
		case gameServer.ServerListStatus:
			server.SetStatus(value)
		case gameServer.ServerType:
			server.SetServerType(value)
		case gameServer.ServerListSquareBracket:
			server.SetShowBracket(value == 1)
		case gameServer.MaxPlayers:
			server.SetMaxPlayer(value)
		case gameServer.ServerAge:
			server.SetAgeLimit(value)
		}
	}
}
