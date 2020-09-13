package serverpackets

import (
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/models"
	"l2goserver/packets"
)

type InitLs struct {
	data   []byte
	offset int
	size   int
}

func NewInitPacket(c models.Client) []byte {

	buffer := new(packets.Buffer)
	buffer.WriteB([]byte{0x00}) // Packet type: Init    1

	KEK := []byte{
		-3 & 255,
		-100 & 255,
		-74 & 255,
		-48 & 255,
		-54 & 255,
		45,
		122,
		87,
		70,
		56,
		74,
		12,
		-85 & 255,
		-94 & 255,
		-24 & 255,
		123,
		93,
		-81 & 255,
		108,
		2,
		-28 & 255,
		57,
		101,
		-118 & 255,
		29,
		-59 & 255,
		95,
		-4 & 255,
		-90 & 255,
		-22 & 255,
		93,
		-66 & 255,
		-49 & 255,
		88,
		-128 & 255,
		-30 & 255,
		-91 & 255,
		-108 & 255,
		10,
		102,
		14,
		-84 & 255,
		21,
		-68 & 255,
		-106 & 255,
		-102 & 255,
		107,
		60,
		102,
		-114 & 255,
		9,
		-37 & 255,
		39,
		-5 & 255,
		7,
		49,
		-16 & 255,
		-104 & 255,
		-127 & 255,
		-26 & 255,
		-71 & 255,
		-22 & 255,
		-93 & 255,
		86,
		-84 & 255,
		92,
		65,
		28,
		-93 & 255,
		1,
		69,
		-96 & 255,
		12,
		-106 & 255,
		23,
		7,
		97,
		56,
		40,
		-91 & 255,
		-93 & 255,
		-4 & 255,
		-92 & 255,
		-123 & 255,
		55,
		-22 & 255,
		81,
		65,
		30,
		-46 & 255,
		83,
		115,
		50,
		-40 & 255,
		83,
		20,
		125,
		-88 & 255,
		10,
		62,
		42,
		-98 & 255,
		112,
		-104 & 255,
		-67 & 255,
		103,
		-31 & 255,
		69,
		-44 & 255,
		-82 & 255,
		-104 & 255,
		36,
		-44 & 255,
		-45 & 255,
		-63 & 255,
		103,
		2,
		-120 & 255,
		-48 & 255,
		-93 & 255,
		-18 & 255,
		89,
		-74 & 255,
		87,
		6,
		101,
		-22 & 255,
		103,
	}

	//	buffer.WriteB(KEK)
	buffer.WriteB(c.SessionID) // Session id?    4  6
	buffer.WriteD(0xc621)      // PROTOCOL_REV	4  8
	buffer.WriteB(KEK)         // pub key 	128  138

	buffer.WriteD(0x29DD954E) // 4  142
	buffer.WriteD(0x77C39CFC) // 4  146
	buffer.WriteD(0x97ADB620) // 4  150
	buffer.WriteD(0x07BDE0F7) // 4  154

	buffer.WriteB(crypt.StaticBlowfish) // 16  170
	buffer.WriteB([]byte{0x00})         // 1 index170???
	return buffer.Bytes()
}
