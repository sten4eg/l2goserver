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
		log.Fatal("Ошибка чтения конфига", err)
	}
	fmt.Println("конфигурационый файл прочитан")

	dbConn, err := db.ConfigureDB()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("подключение к БД установлено")
	manager, err := ipManager.LoadBannedIp(dbConn)
	if err != nil {
		log.Fatal(err)
	}

	err = gameserver.HandlerInit(dbConn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ожидание подключения к геймсерверу")
	loginServer, err := loginserver.New(dbConn, manager)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ожидание подключения клиентов")
	loginServer.Run()
}
