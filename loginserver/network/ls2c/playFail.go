package ls2c

import (
	"l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/packets"
)

func NewPlayFailPacket(reason clientReasons.ClientLoginFailed) []byte {
	buffer := packets.GetBuffer()
	defer packets.Put(buffer)
	buffer.WriteSingleByte(0x06) // Packet type: PlayFail
	buffer.WriteDU(uint32(reason))

	return buffer.CopyBytes()
}
