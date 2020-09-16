package serverpackets

import (
	"l2goserver/packets"
)

func Newggauth(sessionID []byte) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x0b)
	buffer.WriteSlice(sessionID)

	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	return buffer.Bytes()
}
