package main

import (
	"crypto/rand"
	"crypto/rsa"
	"l2goserver/config"
	"l2goserver/loginserver"
	"log"
)

func main()  {

	globalConfig := config.Read()
	server := loginserver.New(globalConfig)

	lenaPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := lenaPrivateKey.PublicKey

	log.Fatal([]byte{pub.N})
	server.Init()
	server.Start()
log.Fatal(server)
}
