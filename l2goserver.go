package main

import (
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/loginserver"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config.Read()
	loginServer := loginserver.New(config.GetConfig())

	db.ConfigureDB()
	loginServer.Init()
	loginServer.Run()

}
