package c2ls

import (
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/network/ls2c"
)

func RequestServerList(client *models.ClientCtx) error {
	return ls2c.ServerList(client)
}
