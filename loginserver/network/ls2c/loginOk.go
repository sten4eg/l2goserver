package ls2c

import (
	"l2goserver/loginserver/network"
	"l2goserver/packets"
)

func NewLoginOkPacket(client network.Ls2c) []byte {
	buffer := packets.GetBuffer()
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
