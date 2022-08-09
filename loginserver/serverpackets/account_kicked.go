package serverpackets

import (
	"l2goserver/loginserver/types/reason"
	"l2goserver/packets"
)

func AccountKicked(reason reason.AccountKickedReason) *packets.Buffer {
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x02)
	buffer.WriteD(int32(reason))

	return buffer
}
