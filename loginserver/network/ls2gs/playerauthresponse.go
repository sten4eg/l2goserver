package ls2gs

import "l2goserver/packets"

func PlayerAuthResponse(account string, response bool) *packets.Buffer {
	buf := packets.Get()
	buf.WriteSingleByte(0x03)
	buf.WriteS(account)
	if response {
		buf.WriteSingleByte(1)
	} else {
		buf.WriteSingleByte(0)
	}
	return buf
}
