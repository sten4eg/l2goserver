package ls2c

import (
	"l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/packets"
)

func NewLoginFailPacket(reason clientReasons.ClientLoginFailed) *packets.Buffer {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x01) // Packet type: LoginFail
	buffer.WriteSingleByte(byte(reason))

	return buffer
}
