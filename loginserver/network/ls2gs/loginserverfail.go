package ls2gs

import (
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
)

func LoginServerFail(reason state.LoginServerFail) *packets.Buffer {
	buf := packets.Get()
	buf.WriteSingleByte(0x01)
	buf.WriteSingleByte(byte(reason))
	return buf
}
