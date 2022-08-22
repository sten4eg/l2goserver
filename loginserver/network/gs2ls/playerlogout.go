package gs2ls

import "l2goserver/packets"

type RemoveAccountInterface interface {
	RemoveAccountOnGameServer(string)
}

func PlayerLogout(data []byte, server RemoveAccountInterface) {
	packet := packets.NewReader(data)
	_ = packet.ReadSingleByte()
	account := packet.ReadString()

	server.RemoveAccountOnGameServer(account)

}
