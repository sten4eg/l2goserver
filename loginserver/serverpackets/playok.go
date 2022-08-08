package serverpackets

import (
	"l2goserver/loginserver/models"
	"l2goserver/packets"
)

func NewPlayOkPacket(client *models.ClientCtx, buffer *packets.Buffer) *packets.Buffer {
	buffer.WriteSingleByte(0x07)
	buffer.WriteDU(client.SessionKey.PlayOk1) // Session Key
	buffer.WriteDU(client.SessionKey.PlayOk2) // Session Key

	return buffer
}
