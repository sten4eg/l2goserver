package c2ls

import (
	"errors"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/network/ls2c"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
)

var wrongSession = errors.New("sessionId не совпал")

func NewAuthGameGuard(request []byte, ctx *models.ClientCtx) error {
	var sessionId uint32
	var packet = packets.NewReader(request)

	sessionId = packet.ReadUInt32()

	if ctx.SessionID != sessionId {
		return wrongSession
	}
	ctx.SetState(state.AuthedGameGuard)
	return ctx.SendBuf(ls2c.Newggauth(ctx.SessionID))

}
