package clientpackets

import (
	"encoding/binary"
	"l2goserver/packets"
	"log"
)

func NewAuthGameGuard(request, clientSessionId []byte) []byte {
	var sessionId uint32
	var packet = packets.NewReader(request)

	sessionId = packet.ReadUInt32()

	data := binary.LittleEndian.Uint32(clientSessionId)
	if data != sessionId {
		log.Fatal("wrong sessionId") // Todo kick clienta
	}
	return clientSessionId

}
