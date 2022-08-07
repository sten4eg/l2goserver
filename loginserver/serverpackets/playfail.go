package serverpackets

import (
	"l2goserver/loginserver/types/reason"
	"l2goserver/packets"
)

func NewPlayFailPacket(reason reason.Reason) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteByte(0x06) // Packet type: PlayFail
	buffer.WriteD(uint32(reason))

	return buffer.Bytes()
}
