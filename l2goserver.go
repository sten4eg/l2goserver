package main

import (
	"fmt"
	"l2goserver/config"
	"l2goserver/database"
	"l2goserver/ipManager"
	"l2goserver/loginserver"
	"l2goserver/loginserver/gameserver"
	"log"
	"os"
	"runtime/trace"
	"time"
)

func main() {
	go Trace()

	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Ошибка чтения конфига", err)
	}
	db, err := database.NewDatabase(cfg.LoginServer.Database)
	if err != nil {
		log.Fatal(err)
	}

	err = gameserver.HandlerInit(db)
	if err != nil {
		log.Fatal(err)
	}

	loginServer, err := loginserver.New(db)
	if err != nil {
		log.Fatal(err)
	}

	err = ipManager.LoadBannedIp(db)
	if err != nil {
		log.Fatal(err)
	}

	loginServer.Run()

}
func Trace() {
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("failed to create trace output file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close trace file: %v", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		log.Fatalf("failed to start trace: %v", err)
	}

	time.Sleep(time.Second * 20)
	trace.Stop()
	fmt.Println("END TRACE")
}
func F() {
	for {
		time.Sleep(time.Second * 1)
		log.Println("a:", loginserver.Atom.Load())
		log.Println("k:", loginserver.AtomKick.Load())
	}
}
