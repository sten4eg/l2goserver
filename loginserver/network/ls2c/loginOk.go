package ls2c

import (
	"l2goserver/packets"
)

type i interface {
	GetSessionLoginOK1() uint32
	GetSessionLoginOK2() uint32
}

func NewLoginOkPacket(client i) []byte {
	buffer := packets.GetBuffer()
	defer packets.Put(buffer)
	buffer.WriteSingleByte(0x03)
	buffer.WriteDU(client.GetSessionLoginOK1())
	buffer.WriteDU(client.GetSessionLoginOK2())
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x000003ea)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)

	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)
	buffer.WriteD(0x00)

	return buffer.CopyBytes()
}
