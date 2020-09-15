package clientpackets

import (
	"l2goserver/loginserver/models"
	"math/big"
)

type RequestAuthLogin struct {
	Username string
	Password string
}

func NewRequestAuthLogin(request []byte, client models.Client) RequestAuthLogin {
	var result RequestAuthLogin

	data := request[:128]

	c := new(big.Int).SetBytes(data)
	decodeData := c.Exp(c, client.PrivateKey.D, client.PrivateKey.N).Bytes()

	result.Username = string(decodeData[1:14])
	result.Password = string(decodeData[14:28])

	return result
}
