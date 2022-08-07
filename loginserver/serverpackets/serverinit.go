package serverpackets

import (
	"l2goserver/loginserver/models"
	"l2goserver/packets"
)

func NewInitPacket(c *models.ClientCtx) []byte {

	buffer := new(packets.Buffer)
	buffer.WriteSingleByte(0x00)

	buffer.WriteD(c.SessionID)           // SessionId
	buffer.WriteD(0xc621)                // PROTOCOL_REV
	buffer.WriteSlice(c.ScrambleModulus) // pubKey

	// unk GG related?
	buffer.WriteD(0x29DD954E)
	buffer.WriteD(0x77C39CFC)
	buffer.WriteD(0x97ADB620)
	buffer.WriteD(0x07BDE0F7)

	buffer.WriteSlice(c.BlowFish)
	buffer.WriteSingleByte(0x00)
	return buffer.Bytes()
}
