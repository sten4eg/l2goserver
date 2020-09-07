package models

import (
	"crypto/rand"
	"errors"
	"fmt"
	"l2goserver/loginserver/crypt"
	"l2goserver/packets"
	"log"
	"net"
)

type Client struct {
	Account   Account
	SessionID []byte
	Socket    net.Conn
}

func NewClient() *Client {
	id := make([]byte, 16)
	_, err := rand.Read(id)

	if err != nil {
		return nil
	}
	return &Client{SessionID: id}
}

func (c *Client) Receive() (opcode byte, data []byte, e error) {
	// Read the first two bytes to define the packet size
	header := make([]byte, 2)
	n, err := c.Socket.Read(header)
	if n != 2 || err != nil {
		return 0x00, nil, errors.New("An error occured while reading the packet header.")
	}

	// Calculate the packet size
	size := 0
	size = size + int(header[0])
	size = size + int(header[1])*256

	// Allocate the appropriate size for our data (size - 2 bytes used for the length
	data = make([]byte, size-2)

	// Read the encrypted part of the packet
	n, err = c.Socket.Read(data)

	if n != size-2 || err != nil {
		return 0x00, nil, errors.New("An error occured while reading the packet data.")
	}

	crypt.Decrypt(&data, 0, 40)
	log.Println(data)
	log.Println(len(data))
	log.Printf("%X\n", data)
	log.Println(data[0])
	// Print the raw packet
	fmt.Printf("package : Header: %X  Data: %X\n", header, data)
	// Decrypt the packet data using the blowfish key
	//data, err = crypt.BlowfishDecrypt(data, []byte("_;v.]05-31!|+-%xT!^[$\000"))

	if err != nil {
		return 0x00, nil, errors.New("An error occured while decrypting the packet data.")
	}

	// Verify our checksum...
	if check := crypt.Checksum(data); check {
		fmt.Printf("Расшифрованный контент пакета : %X\n", data)
		fmt.Println("Чексумма пакета ok")
	} else {
		return 0x00, nil, errors.New("Bad chechsum.")
	}

	// Extract the op code
	//log.Println(data)
	log.Printf("%X\n", data)
	log.Printf("%X\n", data[0])
	//log.Println(data[0])
	opcode = data[0]

	data = data[1:]
	e = nil
	return
}

func (c *Client) Send(data []byte, params ...bool) error {

	var doChecksum, doBlowfish bool = true, true

	// Should we skip the checksum?
	if len(params) >= 1 && params[0] == false {
		doChecksum = false
	}

	// Should we skip the blowfish encryUnable to determine package typeption?
	if len(params) >= 2 && params[1] == false {
		doBlowfish = false
	}

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

		// Finally do the checksum
		crypt.Checksum(data)
	}
	if doBlowfish == true {
		var err error
		//data, err = crypt.BlowfishEncrypt(data, []byte("_;v.]05-31!|+-%xT!^[$\000"))

		if err != nil {
			return err
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
