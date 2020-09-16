package serverpackets

import (
	"l2goserver/loginserver/models"
	"l2goserver/packets"
)

func NewPlayOkPacket(client *models.Client) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x07)
	buffer.WriteD(client.SessionKey.PlayOk1) // Session Key
	buffer.WriteD(client.SessionKey.PlayOk2) // Session Key

	return buffer.Bytes()
}
