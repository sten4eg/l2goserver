package clientpackets

import (
	"crypto/rand"
	"crypto/rsa"
	"l2goserver/loginserver/models"
	"log"
)

type RequestAuthLogin struct {
	Username string
	Password string
}

func NewRequestAuthLogin(request []byte, client models.Client) RequestAuthLogin {
	var result RequestAuthLogin

	rng := rand.Reader
	xxx, err := rsa.DecryptPKCS1v15(rng, client.PrivateKey, request)
	if err != nil {
		log.Fatal(err)
	}
	//xxx, err := rsa.DecryptPKCS1v15(rng, client.PrivateKey ,request)
	//if err != nil {
	//	log.Fatal(err)
	//}

	result.Username = string(xxx[:14])
	result.Password = string(xxx[14:28])

	return result
}
