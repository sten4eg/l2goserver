package gs2ls

import (
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
)

type ServerStatusInterface interface {
	SetStatus(state.ServerStatusValues)
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
		case state.ServerListStatus:
			server.SetStatus(value)
		case state.ServerType:
			server.SetServerType(value)
		case state.ServerListSquareBracket:
			server.SetShowBracket(value == 1)
		case state.MaxPlayers:
			server.SetMaxPlayer(value)
		case state.ServerAge:
			server.SetAgeLimit(value)
		}
	}
}
