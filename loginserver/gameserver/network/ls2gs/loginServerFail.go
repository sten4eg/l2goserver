package ls2gs

import (
	"l2goserver/loginserver/types/reason/loginServer"
	"l2goserver/packets"
)

func LoginServerFail(reason loginServer.FailReason) *packets.Buffer {
	buf := packets.Get()
	buf.WriteSingleByte(0x01)
	buf.WriteSingleByte(byte(reason))
	return buf
}
