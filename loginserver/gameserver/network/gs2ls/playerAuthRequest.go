package gs2ls

import (
	"l2goserver/loginserver/gameserver/network/ls2gs"
	"l2goserver/packets"
)

type playerAuthRequestInterface interface {
	Send(*packets.Buffer) error
	LoginServerGetSessionKey(string) (uint32, uint32, uint32, uint32)
	LoginServerRemoveAuthedLoginClient(string)
}

func PlayerAuthRequest(data []byte, gs playerAuthRequestInterface) {
	packet := packets.NewReader(data)
	account := packet.ReadString()
	playerKey1 := packet.ReadUInt32()
	playerKey2 := packet.ReadUInt32()
	loginKey1 := packet.ReadUInt32()
	loginKey2 := packet.ReadUInt32()

	LoginOk1, LoginOk2, PlayOk1, PlayOk2 := gs.LoginServerGetSessionKey(account)
	var buffer *packets.Buffer

	if playerKey1 != PlayOk1 || playerKey2 != PlayOk2 || loginKey1 != LoginOk1 || loginKey2 != LoginOk2 {
		gs.LoginServerRemoveAuthedLoginClient(account)
		buffer = ls2gs.PlayerAuthResponse(account, true)
	} else {
		buffer = ls2gs.PlayerAuthResponse(account, false)
	}

	_ = gs.Send(buffer)
}
