package models

import (
	"crypto/rsa"
	"errors"
	"l2goserver/loginserver/crypt"
	"l2goserver/packets"
	"math/rand"
	"net"
)

type Client struct {
	Account         Account
	SessionID       uint32
	Socket          net.Conn
	ScrambleModulus []byte
	SessionKey      *SessionKey
	PrivateKey      *rsa.PrivateKey
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
	return &Client{SessionID: id, SessionKey: &sk}
}

func (c *Client) Receive() (uint8, []byte, error) {
	// Read the first two bytes to define the packet size
	header := make([]byte, 2)
	n, err := c.Socket.Read(header)
	if n != 2 || err != nil {
		return 0, nil, errors.New("An error occured while reading the packet header.3" + c.Socket.LocalAddr().String())
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
	fullPackage = crypt.DecodeData(fullPackage)

	opcode := fullPackage[0]

	return opcode, fullPackage[1:], nil
}

func (c *Client) Send(data []byte) error {
	data = crypt.EncodeData(data)
	// Calculate the packet length
	length := uint16(len(data) + 2)
	// Put everything together
	buffer := packets.NewBuffer()
	buffer.WriteH(length)
	buffer.Write(data)

	_, err := c.Socket.Write(buffer.Bytes())

	if err != nil {
		return errors.New("The packet couldn't be sent.")
	}

	return nil
}
