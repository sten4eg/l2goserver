package serverpackets

import (
	"l2goserver/packets"
)

func NewPlayFailPacket(reason uint32) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteByte(0x06) // Packet type: PlayFail
	buffer.WriteD(reason)

	return buffer.Bytes()
}
