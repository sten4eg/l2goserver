package ls2c

import (
	"l2goserver/loginserver/models"
	"l2goserver/packets"
)

func NewPlayOkPacket(client *models.ClientCtx) *packets.Buffer {
	buffer := packets.Get()
	buffer.WriteSingleByte(0x07)
	buffer.WriteDU(client.SessionKey.PlayOk1) // Session Key
	buffer.WriteDU(client.SessionKey.PlayOk2) // Session Key

	return buffer
}
