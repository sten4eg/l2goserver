package serverpackets

import (
	"l2goserver/packets"
)

func Newggauth(sessionID uint32) *packets.Buffer {
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x0b)
	buffer.WriteDU(sessionID)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	return buffer
}
