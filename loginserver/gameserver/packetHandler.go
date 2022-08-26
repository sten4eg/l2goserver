package gameserver

import (
	"fmt"
	"l2goserver/loginserver/gameserver/network/gs2ls"
	"l2goserver/loginserver/types/state/gameServer"
)

func (gsi *Info) HandlePacket(data []byte) error {
	var err error
	opcode := data[0]
	data = data[1:]
	fmt.Println(opcode)

	switch gsi.state {
	case gameServer.Connected:
		if opcode == 0 {
			gs2ls.BlowFishKey(data, gsi)
		}
	case gameServer.BfConnected:
		if opcode == 1 {
			err = gs2ls.GameServerAuth(data, gsi)
		}
	case gameServer.Authed:
		switch opcode {
		case 0x02:
			gs2ls.PlayerInGame(data, gsi)
		case 0x03:
			gs2ls.PlayerLogout(data, gsi)
		case 0x04:
			gs2ls.ChangeAccessLevel(data)
		case 0x05:
			gs2ls.PlayerAuthRequest(data, gsi)
		case 0x06:
			gs2ls.ServerStatus(data, gsi)
		case 0x07:
			gs2ls.PlayerTracert(data)
		case 0x08:
			gs2ls.ReplyCharacters(data, gsi)
		case 0x0A:
			gs2ls.RequestTempBan(data)
		}
	}
	return err
}
