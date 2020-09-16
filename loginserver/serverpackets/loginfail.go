package serverpackets

import (
	"l2goserver/packets"
)

func NewLoginFailPacket(reason uint32) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteByte(0x01) // Packet type: LoginFail
	buffer.WriteD(reason)

	return buffer.Bytes()
}
