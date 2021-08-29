package models

import (
	crand "crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"l2goserver/loginserver/crypt"
	"l2goserver/packets"
	"math/rand"
	"net"
	"strconv"
)

type Client struct {
	Account         Account
	SessionID       uint32
	Socket          net.Conn
	ScrambleModulus []byte
	SessionKey      *SessionKey
	PrivateKey      *rsa.PrivateKey
	BlowFish        []byte
}
type SessionKey struct {
	PlayOk1  uint32
	PlayOk2  uint32
	LoginOk1 uint32
	LoginOk2 uint32
}

func NewClient() *Client {
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

	return &Client{SessionID: id, SessionKey: &sk, BlowFish: blowfish, PrivateKey: privateKey, ScrambleModulus: scrambleModulus}
}

func (c *Client) Receive() (uint8, []byte, error) {
	// Read the first two bytes to define the packet size
	header := make([]byte, 2)
	n, err := c.Socket.Read(header)
	if n != 2 {
		return 0, nil, errors.New("Ожидалось 2 байта длинны, получено: " + strconv.Itoa(n))
	}
	if err != nil {
		return 0, nil, err
	}

	// Calculate the packet size
	size := 0
	size += int(header[0])
	size += int(header[1]) * 256

	// Allocate the appropriate size for our data (size - 2 bytes used for the length
	data := make([]byte, size-2)

	// Read the encrypted part of the packet
	n, err = c.Socket.Read(data)

	if n != size-2 || err != nil {
		return 0, nil, errors.New("An error occured while reading the packet data.")
	}

	fullPackage := header
	fullPackage = append(fullPackage, data...)
	fullPackage = crypt.DecodeData(fullPackage, c.BlowFish)

	opcode := fullPackage[0]

	return opcode, fullPackage[1:], nil
}

func (c *Client) Send(data []byte) error {

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
