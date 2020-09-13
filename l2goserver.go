package main

import (
	"l2goserver/config"
	"l2goserver/loginserver"
)

func main() {

	globalConfig := config.Read()
	server := loginserver.New(globalConfig)

	server.Init()
	server.Start()
}
