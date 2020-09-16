package main

import (
	"l2goserver/config"
	"l2goserver/loginserver"
)

func main() {

	globalConfig := config.Read()
	loginServer := loginserver.New(globalConfig)

	loginServer.Init()
	loginServer.Start()
}
