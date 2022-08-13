package c2ls

import (
	"l2goserver/packets"
)

type RequestServerList struct {
	SessionID []byte
}

func NewRequestServerList(request []byte) RequestServerList {
	var packet = packets.NewReader(request)
	var result RequestServerList

	result.SessionID = packet.ReadBytes(8)

	return result
}
