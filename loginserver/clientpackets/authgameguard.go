package clientpackets

import (
	"l2goserver/packets"
	"log"
)

func NewAuthGameGuard(request []byte, clientSessionId uint32) uint32 {
	var sessionId uint32
	var packet = packets.NewReader(request)

	sessionId = packet.ReadUInt32()

	if clientSessionId != sessionId {
		log.Fatal("wrong sessionId") // Todo kick clienta
	}
	return clientSessionId

}
