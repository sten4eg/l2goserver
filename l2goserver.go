package main

import (
	"fmt"
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/loginserver"
	"l2goserver/loginserver/gameserver"
	"log"
	"os"
	"runtime/debug"
	"runtime/trace"
	"time"
)

func main() {
	debug.SetGCPercent(200000)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//defer profile.Start(profile.MemProfile, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileAllocs, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileHeap, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.BlockProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.TraceProfile, profile.ProfilePath(".")).Stop()

	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.GoroutineProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MutexProfile, profile.ProfilePath(".")).Stop()

	go Trace()
	//go F()
	err := config.Read()
	if err != nil {
		log.Fatal("Ошибка чтения конфига", err)
	}
	gameserver.GameServerHandlerInit()

	loginServer := loginserver.New(config.GetConfig())

	err = db.ConfigureDB()
	if err != nil {
		log.Fatal(err)
	}

	err = loginserver.LoadBannedIp()
	if err != nil {
		log.Fatal(err)
	}

	err = loginServer.StartListen()
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
