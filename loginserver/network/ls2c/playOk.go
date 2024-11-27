package ls2c

import (
	"l2goserver/loginserver/network"
	"l2goserver/packets"
)

func NewPlayOkPacket(client network.Ls2c) []byte {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x07)
	buffer.WriteDU(client.GetSessionPlayOK1()) // Session Key
	buffer.WriteDU(client.GetSessionPlayOK2()) // Session Key

	return buffer.CopyBytes()
}
