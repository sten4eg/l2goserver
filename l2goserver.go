package main

import (
	"l2goserver/config"
	"l2goserver/loginserver"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	globalConfig := config.Read()
	loginServer := loginserver.New(globalConfig)

	loginServer.Init()
	loginServer.Start()

}
