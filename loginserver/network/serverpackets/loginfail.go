package serverpackets

import (
	"l2goserver/loginserver/types/reason"
	"l2goserver/packets"
)

func NewLoginFailPacket(reason reason.Reason) *packets.Buffer {
	buffer := packets.Get()
	buffer.WriteSingleByte(0x01) // Packet type: LoginFail
	buffer.WriteSingleByte(byte(reason))

	return buffer
}
