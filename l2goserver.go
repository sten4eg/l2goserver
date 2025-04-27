package main

import (
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/ipManager"
	"l2goserver/loginserver"
	"l2goserver/loginserver/gameserver"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := config.Read()
	if err != nil {
		log.Fatal("Ошибка чтения конфига", err)
	}

	dbConn, err := db.ConfigureDB()
	if err != nil {
		log.Fatal(err)
	}

	manager, err := ipManager.LoadBannedIp(dbConn)
	if err != nil {
		log.Fatal(err)
	}

	err = gameserver.HandlerInit(dbConn)
	if err != nil {
		log.Fatal(err)
	}

	loginServer, err := loginserver.New(dbConn, manager)
	if err != nil {
		log.Fatal(err)
	}

	//loginserver.InitializeFloodProtection() //TODO
	loginServer.Run()

}
