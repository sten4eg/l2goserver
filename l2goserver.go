package main

import (
	"fmt"
	"github.com/pkg/profile"
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/loginserver"
	"log"
	"unsafe"
)

func main() {
	type Cipher struct { // 4168
		s0, s1, s2, s3 [256]uint32
		p              [18]uint32
	}
	type Cipher2 struct { // 4168
		s0, s1, s2, s3 [256]uint32
		p              [18]uint32
		_              [503]int
	}

	var x Cipher2

	fmt.Println(unsafe.Sizeof(x))

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//	defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	defer profile.Start(profile.MemProfileAllocs, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfileHeap, profile.MemProfileRate(1), profile.ProfilePath(".")).Stop()

	//	defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.GoroutineProfile, profile.ProfilePath(".")).Stop()
	//	defer profile.Start(profile.MutexProfile, profile.ProfilePath(".")).Stop()

	config.Read()
	loginServer := loginserver.New(config.GetConfig())

	db.ConfigureDB()
	loginServer.Init()
	loginServer.Run()

}
