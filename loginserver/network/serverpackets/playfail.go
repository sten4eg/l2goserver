package serverpackets

import (
	"l2goserver/loginserver/types/reason"
	"l2goserver/packets"
)

func NewPlayFailPacket(reason reason.Reason) *packets.Buffer {
	buffer := packets.Get()
	buffer.WriteSingleByte(0x06) // Packet type: PlayFail
	buffer.WriteDU(uint32(reason))

	return buffer
}
