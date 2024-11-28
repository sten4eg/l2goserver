package ls2gs

import "l2goserver/packets"

func InitLS(pubKey []byte) []byte {
	buffer := packets.GetBuffer()
	defer packets.Put(buffer)
	buffer.WriteSingleByte(0x00)
	buffer.WriteD(int32(len(pubKey)))
	buffer.WriteSlice(pubKey)
	return buffer.CopyBytes()
}
