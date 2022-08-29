package gs2ls

import (
	"l2goserver/loginserver/gameserver/network/ls2gs"
	"l2goserver/loginserver/models"
	"l2goserver/packets"
)

type playerAuthRequestInterface interface {
	Send(*packets.Buffer) error
	LoginServerGetSessionKey(string) *models.SessionKey
	LoginServerRemoveAuthedLoginClient(string)
}

func PlayerAuthRequest(data []byte, gs playerAuthRequestInterface) {
	packet := packets.NewReader(data)
	account := packet.ReadString()
	playerKey1 := packet.ReadUInt32()
	playerKey2 := packet.ReadUInt32()
	loginKey1 := packet.ReadUInt32()
	loginKey2 := packet.ReadUInt32()

	key := gs.LoginServerGetSessionKey(account)
	var buffer *packets.Buffer

	if key != nil || playerKey1 != key.PlayOk1 || playerKey2 != key.PlayOk2 || loginKey1 != key.LoginOk1 || loginKey2 != key.LoginOk2 {
		gs.LoginServerRemoveAuthedLoginClient(account)
		buffer = ls2gs.PlayerAuthResponse(account, true)
	} else {
		buffer = ls2gs.PlayerAuthResponse(account, false)
	}

	_ = gs.Send(buffer)
}
