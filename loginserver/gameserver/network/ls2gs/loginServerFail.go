package ls2gs

import (
	"l2goserver/loginserver/types/reason/loginServer"
	"l2goserver/packets"
)

func LoginServerFail(reason loginServer.FailReason) *packets.Buffer {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x01)
	buffer.WriteSingleByte(byte(reason))
	return buffer
}
