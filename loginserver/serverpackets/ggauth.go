package serverpackets

import (
	"l2goserver/packets"
)

func Newggauth(sessionID []byte) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteD(0x0b)      // Packet type: LoginOk
	buffer.WriteB(sessionID) // Session id 1/2

	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	return buffer.Bytes()
}
