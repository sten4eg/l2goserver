package ls2c

import (
	"l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/packets"
)

func NewLoginFailPacket(reason clientReasons.ClientLoginFailed) []byte {
	buffer := packets.GetBuffer()
	defer packets.Put(buffer)
	buffer.WriteSingleByte(0x01) // Packet type: LoginFail
	buffer.WriteSingleByte(byte(reason))

	return buffer.CopyBytes()
}
