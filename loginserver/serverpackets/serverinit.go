package serverpackets

import (
	"l2goserver/loginserver/crypt"
	"l2goserver/packets"
	"log"
)

func NewInitPacket(publicKey []byte, blowfish []byte) []byte {

	buffer := new(packets.Buffer)
	_, err := buffer.Write([]byte{0x00, 0x00, 0x00}) // Packet type: Init    1
	if err != nil {
		log.Fatal(err)
	}
	buffer.WriteDDD(1966545641) // Session id?    4
	buffer.WriteDDD(0xc621)     // PROTOCOL_REV	4
	buffer.WriteB(publicKey)    // Размер Паблик ключа 	128

	buffer.WriteDD(0x29DD954E) // 4
	buffer.WriteDD(0x77C39CFC) // 4
	buffer.WriteDD(0x97ADB620) // 4
	buffer.WriteDD(0x07BDE0F7) // 4

	buffer.WriteB(crypt.Kek)   // 16
	buffer.Write([]byte{0x00}) // 1
	//	buffer.Write([]byte{0x9c, 0x77, 0xed, 0x03}) // Session id?
	//	buffer.Write([]byte{0x5a, 0x78, 0x00, 0x00}) // Protocol version : 785a

	return buffer.Bytes()
}
