package serverpackets

import (
	"l2goserver/packets"
)

func NewLoginFailPacket(reason byte) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x01) // Packet type: LoginFail
	buffer.WriteSingleByte(reason)

	return buffer.Bytes()
}
