package ls2c

import (
	"l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/packets"
)

func AccountKicked(reason clientReasons.ClientLoginFailed) []byte {
	buffer := packets.GetBuffer()
	defer packets.Put(buffer)
	buffer.WriteSingleByte(0x02)
	buffer.WriteD(int32(reason))

	return buffer.CopyBytes()
}
