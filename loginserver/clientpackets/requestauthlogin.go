package clientpackets

import (
	"crypto/rsa"

	_ "github.com/andreburgaud/crypt2go/ecb"
	_ "github.com/andreburgaud/crypt2go/padding"
	"l2goserver/loginserver/models"
	"log"
	"math/big"
)

type RequestAuthLogin struct {
	Username string
	Password string
}

func NewRequestAuthLogin(request []byte, client models.Client) RequestAuthLogin {
	var result RequestAuthLogin

	//rng := rand.Reader
	data := request[:128]

	// rsa.DecryptOAEP(rand.Reader, priv, ciphertext)

	//xxx ,err := PublicDecrypt(&client.PrivateKey.PublicKey, data)
	//if err != nil {
	//log.Fatal(err)
	//}
	c := new(big.Int).SetBytes(data)
	plainText := c.Exp(c, client.PrivateKey.D, client.PrivateKey.N).Bytes()
	x := plainText
	_ = x
	//xxx, err := rsa.DecryptOAEP(sha256.New(), rand.Reader,client.PrivateKey,plainText,[]byte(""))

	//	xxx , err := client.PrivateKey.Decrypt(rand.Reader,data,nil)

	log.Fatal(123321)
	//xxx, err := rsa.DecryptPKCS1v15(rng, client.PrivateKey ,request)
	//if err != nil {
	//	log.Fatal(err)
	//}

	result.Username = string(data[:14])
	result.Password = string(data[14:28])

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
