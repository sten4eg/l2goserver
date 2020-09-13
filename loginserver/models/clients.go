package models

import (
	"crypto/rand"
	"errors"
	"l2goserver/loginserver/crypt"
	"l2goserver/packets"
	"net"
)

type Client struct {
	Account    Account
	SessionID  []byte
	Socket     net.Conn
	Rsa        []byte
	SessionKey []byte
}

func NewClient() *Client {
	id := make([]byte, 4)
	_, err := rand.Read(id)

	if err != nil {
		return nil
	}
	return &Client{SessionID: id}
}

func (c *Client) Receive() (uint8, []byte, error) {
	// Read the first two bytes to define the packet size
	header := make([]byte, 2)
	n, err := c.Socket.Read(header)
	if n != 2 || err != nil {
		return 0, nil, errors.New("An error occured while reading the packet header.")
	}

	// Calculate the packet size
	size := 0
	size = size + int(header[0])
	size = size + int(header[1])*256

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

	opcode := fullPackage[2]

	return opcode, fullPackage[3:], nil
}

func (c *Client) Send(data []byte, params ...bool) error {

	var doChecksum bool = true

	// Should we skip the checksum?
	if len(params) >= 1 && params[0] == false {
		doChecksum = false
	}

	// Should we skip the blowfish encryUnable to determine package typeption?

	if doChecksum == true {
		// Add 4 empty bytes for the checksum new( new(
		data = append(data, []byte{0x00, 0x00, 0x00, 0x00}...)

		// Add blowfish padding
		missing := len(data) % 8

		if missing != 0 {
			for i := missing; i < 8; i++ {
				data = append(data, byte(0x00))
			}
		}

	}

	// Calculate the packet length
	length := uint16(len(data) + 2)
	// Put everything together
	buffer := packets.NewBuffer()
	buffer.WriteUInt16(length)
	buffer.Write(data)

	_, errr := c.Socket.Write(buffer.Bytes())

	if errr != nil {
		return errors.New("The packet couldn't be sent.")
	}

	return nil
}
