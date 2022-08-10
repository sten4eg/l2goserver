package loginserver

import (
	"l2goserver/config"
	"l2goserver/loginserver/models"
	"net"
)

func GSInitialize() {
	var gs models.GS
	gs.InitRSAKeys()

	port := config.GetLoginPortForGameServer()

	blowfishKey := []byte{95, 59, 118, 46, 93, 48, 53, 45, 51, 49, 33, 124, 43, 45, 37, 120, 84, 33, 94, 91, 36, 0}

	gs.SetBlowFishKey(blowfishKey)

	listener, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		panic(err)
	}
	gs.Connection = listener

	go gs.Run()
}
