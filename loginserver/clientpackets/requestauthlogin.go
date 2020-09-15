package clientpackets

import (
	"crypto/rsa"

	_ "github.com/andreburgaud/crypt2go/ecb"
	_ "github.com/andreburgaud/crypt2go/padding"
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
func PublicDecrypt(pubKey *rsa.PublicKey, data []byte) ([]byte, error) {
	c := new(big.Int)
	m := new(big.Int)
	m.SetBytes(data)
	e := big.NewInt(int64(pubKey.E))
	c.Exp(m, e, pubKey.N)
	out := c.Bytes()
	skip := 0
	for i := 2; i < len(out); i++ {
		if i+1 >= len(out) {
			break
		}
		if out[i] == 0xff && out[i+1] == 0 {
			skip = i + 2
			break
		}
	}

	return out[skip:], nil
}
