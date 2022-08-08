package serverpackets

import (
	"l2goserver/loginserver/models"
	"l2goserver/packets"
)

func NewPlayOkPacket(client *models.ClientCtx) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x07)
	buffer.WriteDU(uint32(client.SessionKey.PlayOk1)) // Session Key
	buffer.WriteDU(uint32(client.SessionKey.PlayOk2)) // Session Key

	return buffer.Bytes()
}
