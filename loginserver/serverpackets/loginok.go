package serverpackets

import (
	"l2goserver/loginserver/models"
	"l2goserver/packets"
)

func NewLoginOkPacket(client *models.Client) []byte {
	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x03)              // Packet type: LoginOk
	buffer.WriteD(client.SessionKey.LoginOk1) // SessionKey1_FistPart
	buffer.WriteD(client.SessionKey.LoginOk2) // SessionKey1_SecondPart
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

	return buffer.Bytes()
}
