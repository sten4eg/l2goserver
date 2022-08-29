package ls2gs

import "l2goserver/packets"

func PlayerAuthResponse(account string, response bool) *packets.Buffer {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x03)
	buffer.WriteS(account)
	if response {
		buffer.WriteSingleByte(1)
	} else {
		buffer.WriteSingleByte(0)
	}
	return buffer
}
