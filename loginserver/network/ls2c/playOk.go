package ls2c

import (
	"l2goserver/packets"
)

type playOkPacket interface {
	GetSessionPlayOK1() uint32
	GetSessionPlayOK2() uint32
}

func NewPlayOkPacket(client playOkPacket) []byte {
	buffer := packets.GetBuffer()
	defer packets.Put(buffer)
	buffer.WriteSingleByte(0x07)
	buffer.WriteDU(client.GetSessionPlayOK1()) // Session Key
	buffer.WriteDU(client.GetSessionPlayOK2()) // Session Key

	return buffer.CopyBytes()
}
