package loginserverpackets

import "l2goserver/packets"

func RequestCharacter(account string) *packets.Buffer {
	buf := packets.Get()
	buf.WriteSingleByte(0x05)
	buf.WriteS(account)
	return buf
}
