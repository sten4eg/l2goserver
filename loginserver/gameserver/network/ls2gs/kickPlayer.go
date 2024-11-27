package ls2gs

import "l2goserver/packets"

func KickPlayer(account string) []byte {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x04)
	buffer.WriteS(account)

	return buffer.CopyBytes()
}
