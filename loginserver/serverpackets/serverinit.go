package serverpackets

import (
	"l2goserver/packets"
)

func NewInitPacket(publicKey []byte, blowfish []byte) []byte {

	buffer := new(packets.Buffer)
	buffer.WriteB([]byte{0x00, 0x00, 0x00}) // Packet type: Init    1

	buffer.WriteD(1966545641) // Session id?    4  6
	buffer.WriteD(0xc621)     // PROTOCOL_REV	4  8
	buffer.WriteB(publicKey)  // Размер Паблик ключа 	128  138

	buffer.WriteD(0x29DD954E) // 4  142
	buffer.WriteD(0x77C39CFC) // 4  146
	buffer.WriteD(0x97ADB620) // 4  150
	buffer.WriteD(0x07BDE0F7) // 4  154

	buffer.WriteB(blowfish)     // 16  170
	buffer.WriteB([]byte{0x00}) // 1 index170???

	return buffer.Bytes()
}
