package main

import (
	"github.com/pkg/profile"
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/loginserver"
	"l2goserver/loginserver/gameserver"
	"log"
	"net"
	"runtime/debug"
	"time"
)

type cli struct {
	con net.Conn
}

func (c *cli) Clos() {
	if c.con != nil {
		_ = c.con.Close()
	}
}
func main() {
	var cl cli
	cl.Clos()
	debug.SetGCPercent(20000)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	defer profile.Start(profile.MemProfile, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileAllocs, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileHeap, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.BlockProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.GoroutineProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MutexProfile, profile.ProfilePath(".")).Stop()
	go F()
	config.Read()
	gameserver.GameServerHandlerInit()

	loginServer := loginserver.New(config.GetConfig())

	db.ConfigureDB()
	loginserver.LoadBannedIp()
	loginServer.StartListen()
	loginServer.Run()

}

func F() {
	for {
		time.Sleep(time.Second * 1)
		log.Println(loginserver.Atom.Load())
	}
}
