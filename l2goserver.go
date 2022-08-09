package main

import (
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/loginserver"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//defer profile.Start(profile.MemProfile, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileAllocs, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileHeap, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.BlockProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.GoroutineProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MutexProfile, profile.ProfilePath(".")).Stop()

	config.Read()
	loginServer := loginserver.New(config.GetConfig())
	loginserver.GSInitialize()

	db.ConfigureDB()
	loginserver.LoadBannedIp()
	loginServer.Initialize()
	loginServer.Run()

}
