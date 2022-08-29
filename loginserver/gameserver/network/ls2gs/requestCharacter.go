package ls2gs

import "l2goserver/packets"

func RequestCharacter(account string) *packets.Buffer {
	buf := packets.GetBuffer()
	buf.WriteSingleByte(0x05)
	buf.WriteS(account)
	return buf
}
