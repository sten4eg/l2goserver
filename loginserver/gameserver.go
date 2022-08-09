package loginserver

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"l2goserver/config"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/crypt/blowfish"
	"l2goserver/loginserver/loginserverpackets"
	"l2goserver/packets"
	"l2goserver/utils"
	"log"
	"net"
	"sync"
)

type GS struct {
	utils.NoCopy
	Connection net.Listener
	privateKet *rsa.PrivateKey
	blowfish   *blowfish.Cipher
}
type gsCtx struct {
	mu   sync.Mutex
	conn net.Conn
}

func (gs *GS) initRSAKeys() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		fmt.Println(err)
	}
	gs.privateKet = privateKey

}
func GSInitialize() {
	var gs GS
	gs.initRSAKeys()

	port := config.GetLoginPortForGameServer()

	blowfishKey := []byte{95, 59, 118, 46, 93, 48, 53, 45, 51, 49, 33, 124, 43, 45, 37, 120, 84, 33, 94, 91, 36, 0}
	//blowfishKey = []byte("_;v.]05-31!|+-%xT!^[$\\00")
	// "_;v.]05-31!|+-%xT!^[$\00"

	cipher, err := blowfish.NewCipher(blowfishKey)
	if err != nil {
		panic(err)
	}
	gs.blowfish = cipher
	listener, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		panic(err)
	}
	gs.Connection = listener

	go gs.Run()
}

func (gs *GS) Run() {
	for {
		client := new(gsCtx)
		var err error

		client.conn, err = gs.Connection.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		tmp := gs.privateKet.PublicKey.N.Bytes()
		tmp = append([]byte{0}, tmp...)

		buf := loginserverpackets.InitLS(tmp)

		gs.Send(client, buf)
		go gs.GsPackageHandler(client)
	}
}

func (c *GS) Send(client *gsCtx, buf *packets.Buffer) {
	//data := buf.Bytes()
	size := buf.Len() + 4
	size = (size + 8) - (size % 8) // padding

	data := make([]byte, 200)
	copy(data, buf.Bytes())
	packets.Put(buf)

	rs := crypt.AppendCheckSum(data, size)

	_ = rs
	for i := 0; i < size; i += 8 {
		c.blowfish.Encrypt(rs, rs, i, i)
	}

	rs = rs[:size]
	leng := len(rs) + 2

	s, f := byte(leng>>8), byte(leng&0xff)
	res := append([]byte{f, s}, rs...)

	client.mu.Lock()
	n, err := client.conn.Write(res)
	client.mu.Unlock()
	_ = n
	if err != nil {
		panic(err)
	}
}

func (gs *GS) GsPackageHandler(client *gsCtx) {

	l := make([]byte, 0, 2)
	for {

		//var lengthHi, lengthLo byte
		n, err := client.conn.Read(l)
		if err != nil {
			panic(err)
		}
		_ = n
	}

}
