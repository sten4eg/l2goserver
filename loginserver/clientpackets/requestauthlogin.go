package clientpackets

import (
	"bytes"
	"l2goserver/loginserver/models"
	"math/big"
)

type RequestAuthLogin struct {
	Username string
	Password string
}

func NewRequestAuthLogin(request []byte, client models.Client) RequestAuthLogin {
	var result RequestAuthLogin

	payload := request[:128]

	c := new(big.Int).SetBytes(payload)
	decodeData := c.Exp(c, client.PrivateKey.D, client.PrivateKey.N).Bytes()

	trimLogin := bytes.Trim(decodeData[1:14], string(rune(0)))
	trimPassword := bytes.Trim(decodeData[14:28], string(rune(0)))

	result.Username = string(trimLogin)
	result.Password = string(trimPassword)

	return result
}
