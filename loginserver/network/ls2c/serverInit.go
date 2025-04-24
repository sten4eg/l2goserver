package ls2c

import (
	"l2goserver/loginserver/models"
	"l2goserver/packets"
)

func NewInitPacket(c *models.ClientCtx) error {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x00)

	buffer.WriteDU(c.SessionID)          // SessionId
	buffer.WriteDU(0xc621)               // PROTOCOL_REV
	buffer.WriteSlice(c.ScrambleModulus) // pubKey

	// unk GG related?
	buffer.WriteDU(0x29DD954E)
	buffer.WriteDU(0x77C39CFC)
	buffer.WriteDU(0x97ADB620)
	buffer.WriteDU(0x07BDE0F7)

	buffer.WriteSlice(c.BlowFish)
	buffer.WriteSingleByte(0x00)

	return c.SendBufInit(buffer)
}
