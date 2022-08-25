package ls2gs

import "l2goserver/packets"

func KickPlayer(account string) *packets.Buffer {
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x04)
	buffer.WriteS(account)

	return buffer
}
