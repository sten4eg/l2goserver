package main

import (
	"fmt"
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/ipManager"
	"l2goserver/loginserver"
	"l2goserver/loginserver/gameserver"
	"log"
)

func main() {
	err := config.Read()
	if err != nil {
		log.Fatal("error read config ", err)
	}
	fmt.Println("config file read")

	dbConn, err := db.ConfigureDB()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connection to the database established")
	manager, err := ipManager.LoadBannedIp(dbConn)
	if err != nil {
		log.Fatal(err)
	}

	err = gameserver.HandlerInit(dbConn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Waiting for connection to game server")
	loginServer, err := loginserver.New(dbConn, manager)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Waiting for clients to connect")
	loginServer.Run()
}
