package ls2gs

import (
	"l2goserver/loginserver/types/reason/loginServer"
	"l2goserver/packets"
)

func LoginServerFail(reason loginServer.FailReason) []byte {
	buffer := packets.GetBuffer()
	defer packets.Put(buffer)
	buffer.WriteSingleByte(0x01)
	buffer.WriteSingleByte(byte(reason))
	return buffer.CopyBytes()
}
