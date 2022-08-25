package gs2ls

import "l2goserver/packets"

type playerInGameInterface interface {
	AddAccountOnGameServer(string)
}

func PlayerInGame(data []byte, server playerInGameInterface) {
	packet := packets.NewReader(data)

	size := int(packet.ReadInt16())

	for i := 0; i < size; i++ {
		account := packet.ReadString()
		server.AddAccountOnGameServer(account)
	}
}
