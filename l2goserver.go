package main

import (
	"l2goserver/config"
	"l2goserver/loginserver"
)

func main() {
	//x := crypt.Kek
	//log.Fatal(x, len(x))

	globalConfig := config.Read()
	server := loginserver.New(globalConfig)

	server.Init()
	server.Start()
	//log.Fatal(server)
}
