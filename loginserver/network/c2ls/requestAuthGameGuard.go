package c2ls

import (
	"fmt"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/network/ls2c"
	reasons "l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/loginserver/types/state/clientState"
	"l2goserver/packets"
)

func RequestAuthGameGuard(request []byte, ctx *models.ClientCtx) error {
	var packet = packets.NewReader(request)
	var errMsg = fmt.Errorf("AccessFailed")
	sessionId := packet.ReadUInt32()

	if ctx.SessionID != sessionId {
		_ = ctx.SendBuf(ls2c.NewLoginFailPacket(reasons.AccessFailed))
		return errMsg
	}
	ctx.SetState(clientState.AuthedGameGuard)
	return ctx.SendBuf(ls2c.Newggauth(ctx.SessionID))

}
