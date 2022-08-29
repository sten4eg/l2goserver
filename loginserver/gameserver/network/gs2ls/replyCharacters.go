package gs2ls

import "l2goserver/packets"

type SetCharactersOnServerInterface interface {
	SetCharactersOnServer(account string, charsNum uint8, timeToDel []int64)
}

func ReplyCharacters(data []byte, server SetCharactersOnServerInterface) {
	packet := packets.NewReader(data)
	account := packet.ReadString()
	chars := packet.ReadUInt8()
	charsToDel := packet.ReadInt8()
	charsList := make([]int64, charsToDel)

	for i := 0; i < int(charsToDel); i++ {
		charsList[i] = packet.ReadInt64()
	}
	server.SetCharactersOnServer(account, chars, charsList)
}
