package ls2gs

import "l2goserver/packets"

func RequestCharacter(account string) []byte {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x05)
	buffer.WriteS(account)
	return buffer.CopyBytes()
}
