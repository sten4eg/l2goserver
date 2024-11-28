package ls2c

import (
	"l2goserver/packets"
)

func Newggauth(sessionID uint32) []byte {
	buffer := packets.GetBuffer()
	defer packets.Put(buffer)
	buffer.WriteSingleByte(0x0b)
	buffer.WriteDU(sessionID)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	return buffer.CopyBytes()
}
