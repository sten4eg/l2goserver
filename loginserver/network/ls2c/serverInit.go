package ls2c

import (
	"l2goserver/packets"
)

type initPacketInterface interface {
	GetSessionId() uint32
	GetScrambleModulus() []byte
	GetBlowFish() []byte
}

func NewInitPacket(c initPacketInterface) []byte {
	buffer := packets.GetBuffer()
	buffer.WriteSingleByte(0x00)

	buffer.WriteDU(c.GetSessionId())          // SessionId
	buffer.WriteDU(0xc621)                    // PROTOCOL_REV
	buffer.WriteSlice(c.GetScrambleModulus()) // pubKey

	// unk GG related?
	buffer.WriteDU(0x29DD954E)
	buffer.WriteDU(0x77C39CFC)
	buffer.WriteDU(0x97ADB620)
	buffer.WriteDU(0x07BDE0F7)

	buffer.WriteSlice(c.GetBlowFish())
	buffer.WriteSingleByte(0x00)
	return buffer.CopyBytes()

}
