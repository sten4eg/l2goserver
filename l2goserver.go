package main

import (
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/loginserver"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	globalConfig := config.Read()
	loginServer := loginserver.New(globalConfig)

	db.ConfigureDB()
	loginServer.Init()
	loginServer.Start()

}
