package models

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/crypt/blowfish"
	"l2goserver/loginserver/gameserverpackets"
	"l2goserver/loginserver/loginserverpackets"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"log"
	"net"
	"sync"
)

type GS struct {
	Connection net.Listener
	privateKey *rsa.PrivateKey
	blowfish   *blowfish.Cipher
	mu         sync.Mutex
	conn       net.Conn
	state      state.GameServerState
}

func (gs *GS) InitRSAKeys() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		panic(err)
	}
	gs.privateKey = privateKey

}

func (gs *GS) Run() {
	for {

		var err error

		gs.conn, err = gs.Connection.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		gs.state = state.CONNECTED

		pubKey := make([]byte, 1, 65)
		pubKey = append(pubKey, gs.privateKey.PublicKey.N.Bytes()...)

		buf := loginserverpackets.InitLS(pubKey)

		gs.Send(buf)
		go gs.Listen()
	}
}

func (gs *GS) Send(buf *packets.Buffer) {
	size := buf.Len() + 4
	size = (size + 8) - (size % 8) // padding

	data := make([]byte, 200)
	copy(data, buf.Bytes())
	packets.Put(buf)

	rs := crypt.AppendCheckSum(data, size)

	for i := 0; i < size; i += 8 {
		gs.blowfish.Encrypt(rs, rs, i, i)
	}

	rs = rs[:size]
	leng := len(rs) + 2

	s, f := byte(leng>>8), byte(leng&0xff)
	res := append([]byte{f, s}, rs...)

	gs.mu.Lock()
	_, err := gs.conn.Write(res)
	gs.mu.Unlock()

	if err != nil {
		panic(err)
	}
}

func (gs *GS) Listen() {
	for {
		header := make([]byte, 2)

		n, err := gs.conn.Read(header)
		if err != nil {
			panic(err)
		}
		dataSize := (int(header[0]) | int(header[1])<<8) - 2

		data := make([]byte, dataSize)
		n, err = gs.conn.Read(data)
		if err != nil {
			panic(err)
		}
		if n != dataSize {
			panic("qweqwedsaasdcg")
		}

		for i := 0; i < dataSize; i += 8 {
			gs.blowfish.Decrypt(data, data, i, i)
		}

		ok := crypt.VerifyCheckSum(data, dataSize)
		if !ok {
			fmt.Println("Неверная контрольная сумма пакета, закрытие соединения.")
			gs.conn.Close()
			return
		}
		gs.HandlePackage(data)
	}
}
func (gs *GS) HandlePackage(data []byte) {

	switch gs.state {
	case state.CONNECTED:
		if data[0] == 0 {
			gameserverpackets.BlowFishKey(data, gs)
		}
	case state.BF_CONNECTED:
		if data[0] == 1 {
			gameserverpackets.GameServerAuth(data)
		}
	}
}

func (gs *GS) GetPrivKey() *rsa.PrivateKey {
	return gs.privateKey
}
func (gs *GS) SetBlowFishKey(key []byte) {
	cipher, err := blowfish.NewCipher(key)
	if err != nil {
		panic(err)
	}
	gs.blowfish = cipher
}
func (gs *GS) SetState(state state.GameServerState) {
	gs.state = state
}

func (gs *GS) ForceClose(reason state.LoginServerFail) {
	gs.Send(loginserverpackets.LoginServerFail(reason))
	_ = gs.conn.Close()
}
