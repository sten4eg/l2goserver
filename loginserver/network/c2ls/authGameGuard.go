package c2ls

import (
	"errors"
	"l2goserver/loginserver/network/ls2c"
	"l2goserver/packets"
)

var errWrongSession = errors.New("sessionId не совпал")

type authGameGuardInterface interface {
	GetSessionId() uint32
	Send([]byte) error
}

func NewAuthGameGuard(request []byte, ctx authGameGuardInterface) error {
	var sessionId uint32
	var packet = packets.NewReader(request)

	sessionId = packet.ReadUInt32()

	if ctx.GetSessionId() != sessionId {
		return errWrongSession
	}
	return ctx.Send(ls2c.Newggauth(ctx.GetSessionId()))
}
