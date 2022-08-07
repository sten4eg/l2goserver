package models

import (
	crand "crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"l2goserver/utils"
	"math/rand"
	"net"
	"strconv"
)

type ClientCtx struct {
	noCopy          utils.NoCopy //nolint:unused,structcheck
	Account         Account
	SessionID       uint32
	Socket          net.Conn
	ScrambleModulus []byte
	SessionKey      *SessionKey
	PrivateKey      *rsa.PrivateKey
	BlowFish        []byte
	State           state.GameState
	JoinedGS        bool
}

type SessionKey struct {
	PlayOk1  uint32
	PlayOk2  uint32
	LoginOk1 uint32
	LoginOk2 uint32
}

func NewClient() *ClientCtx {
	id := rand.Uint32()

	sk := SessionKey{
		PlayOk1:  rand.Uint32(),
		PlayOk2:  rand.Uint32(),
		LoginOk1: rand.Uint32(),
		LoginOk2: rand.Uint32(),
	}
	blowfish := make([]byte, 16)
	_, _ = rand.Read(blowfish)

	privateKey, err := rsa.GenerateKey(crand.Reader, 1024)
	if err != nil {
		fmt.Println(err)
	}
	scrambleModulus := crypt.ScrambleModulus(privateKey.PublicKey.N.Bytes())

	return &ClientCtx{
		SessionID:       id,
		SessionKey:      &sk,
		BlowFish:        blowfish,
		PrivateKey:      privateKey,
		ScrambleModulus: scrambleModulus,
		State:           state.NoState,
		JoinedGS:        false,
	}
}

func (c *ClientCtx) Receive() (uint8, []byte, error) {
	header := make([]byte, 2)
	n, err := c.Socket.Read(header)
	if err != nil {
		return 0, nil, err
	}
	if n != 2 {
		return 0, nil, errors.New("Ожидалось 2 байта длинны, получено: " + strconv.Itoa(n))
	}

	// длинна пакета
	dataSize := (int(header[0]) | int(header[1])<<8) - 2

	// аллокация требуемого массива байт для входящего пакета
	data := make([]byte, dataSize)

	n, err = c.Socket.Read(data)

	if n != dataSize || err != nil {
		return 0, nil, errors.New("длинна прочитанного пакета не соответствует требуемому размеру")
	}

	fullPackage := make([]byte, 0, len(header)+len(data))
	fullPackage = append(fullPackage, header...)
	fullPackage = append(fullPackage, data...)

	fullPackage = crypt.DecodeData(fullPackage, c.BlowFish)

	opcode := fullPackage[0]

	return opcode, fullPackage[1:], nil
}

func (c *ClientCtx) Send(data []byte) error {
	data = crypt.EncodeData(data, c.BlowFish)
	// Calculate the packet length
	length := uint16(len(data) + 2)
	// Put everything together
	buffer := packets.NewBuffer()
	buffer.WriteH(length)
	_, err := buffer.Write(data)
	if err != nil {
		return errors.New("The packet couldn't be sent.(write in buffer)")
	}
	_, err = c.Socket.Write(buffer.Bytes())

	if err != nil {
		return errors.New("The packet couldn't be sent.")
	}

	return nil
}
