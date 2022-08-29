package ls2gs

import "l2goserver/packets"

func InitLS(pubKey []byte) *packets.Buffer {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x00)
	buffer.WriteD(int32(len(pubKey)))
	buffer.WriteSlice(pubKey)
	return buffer
}
