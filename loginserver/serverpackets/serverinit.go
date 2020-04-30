package serverpackets

import (
	"crypto/rand"
	"crypto/rsa"
	"l2goserver/packets"
)

func NewInitPacket() []byte {
	lenaPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := lenaPrivateKey.PublicKey

	buffer := new(packets.Buffer)
	buffer.WriteByte(0x00) // Packet type: Init
	buffer.WriteD(0x0106) // PROTOCOL_REV
	buffer.WriteInt(pub.Size()) // Размер Паблик ключа


	//buffer.WriteByte(pub)


	buffer.Write([]byte{0x9c, 0x77, 0xed, 0x03}) // Session id?
	buffer.Write([]byte{0x5a, 0x78, 0x00, 0x00}) // Protocol version : 785a

	return buffer.Bytes()
}
