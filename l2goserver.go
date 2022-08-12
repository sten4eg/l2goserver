package main

import (
	"github.com/pkg/profile"
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/loginserver"
	"l2goserver/loginserver/gameserver"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	defer log.Println(loginserver.Atom.Load())

	defer profile.Start(profile.MemProfile, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileAllocs, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileHeap, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.BlockProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.GoroutineProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MutexProfile, profile.ProfilePath(".")).Stop()
	go F()
	config.Read()
	loginServer := loginserver.New(config.GetConfig())

	gameserver.GameServerHandlerInit()
	db.ConfigureDB()
	loginserver.LoadBannedIp()
	loginServer.StartListen()
	loginServer.Run()

}

func F() {
	//for {
	//	time.Sleep(time.Second * 1)
	//	log.Println(loginserver.Atom.Load())
	//}
}
