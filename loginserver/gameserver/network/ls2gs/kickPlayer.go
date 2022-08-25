package ls2gs

import "l2goserver/packets"

func KickPlayer(account string) *packets.Buffer {
	buf := packets.Get()
	buf.WriteSingleByte(0x04)
	buf.WriteS(account)

	return buf
}
