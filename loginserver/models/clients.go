package models

import (
	crand "crypto/rand"
	"crypto/rsa"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"l2goserver/crypt"
	"l2goserver/loginserver/types/state/clientState"
	"l2goserver/packets"
	"math/rand"
	"net"
)

type ClientCtx struct {
	_               noCopy //nolint:unused,structcheck
	joinedGS        bool
	state           clientState.ClientCtxState
	SessionID       uint32
	Uid             uint64
	conn            *net.TCPConn
	SessionKey      *SessionKey
	PrivateKey      *rsa.PrivateKey
	BlowFish        []byte
	ScrambleModulus []byte
	Account         Account
}

type SessionKey struct {
	PlayOk1  uint32
	PlayOk2  uint32
	LoginOk1 uint32
	LoginOk2 uint32
}

func NewClient() (*ClientCtx, error) {
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

	//scrambleModulus := []byte{134, 142, 95, 160, 18, 252, 106, 59, 228, 254, 60, 14, 60, 2, 90, 106, 224, 241, 174, 178, 47, 66, 122, 21, 110, 215, 76, 146, 27, 182, 122, 150, 1, 134, 164, 255, 126, 28, 105, 76, 133, 192, 162, 208, 233, 9, 184, 101, 194, 45, 164, 247, 101, 2, 210, 212, 118, 99, 115, 43, 231, 32, 183, 49, 136, 115, 208, 243, 39, 171, 54, 233, 219, 240, 167, 155, 202, 241, 240, 210, 1, 247, 75, 86, 226, 199, 41, 87, 111, 247, 168, 33, 182, 40, 202, 11, 189, 174, 210, 199, 242, 41, 127, 49, 208, 44, 221, 72, 240, 95, 21, 2, 195, 222, 83, 6, 225, 251, 182, 0, 179, 43, 149, 226, 56, 43, 3, 2}
	return &ClientCtx{
		SessionID:       id,
		SessionKey:      &sk,
		BlowFish:        blowfish,
		PrivateKey:      privateKey,
		ScrambleModulus: scrambleModulus,
		state:           clientState.NoState,
		joinedGS:        false,
		Uid:             rand.Uint64(),
	}, nil
}

func (c *ClientCtx) SetConn(conn *net.TCPConn) {
	c.conn = conn
}

func (c *ClientCtx) Receive() (uint8, []byte, error) {
	header := make([]byte, 2)
	_, err := io.ReadFull(c.conn, header)

	if err != nil {
		return 0, nil, err
	}

	dataSize := int(binary.LittleEndian.Uint16(header)) - 2
	if dataSize < 0 {
		return 0, nil, errors.New("negative data size")
	}

	data := make([]byte, dataSize)
	_, err = io.ReadFull(c.conn, data)
	if err != nil {
		return 0, nil, err
	}

	if ok := crypt.DecodeData(data, c.BlowFish); !ok {
		return 0, nil, errors.New("DecodeData fail")
	}

	if len(data) < 1 {
		return 0, nil, errors.New("data empty")
	}

	return data[0], data[1:], nil
}
func (c *ClientCtx) SendBuf(buffer *packets.Buffer) error {
	data := buffer.Bytes()
	defer packets.Put(buffer)

	data = crypt.EncodeData(data, c.BlowFish)
	// calculate packet length
	length := uint16(len(data) + 2)

	s, f := byte(length>>8), byte(length&0xff)

	data = append([]byte{f, s}, data...)

	_, err := c.conn.Write(data)
	if err != nil {
		return errors.New("packet can not be send")
	}

	return nil
}
func (c *ClientCtx) SendBufInit(buffer *packets.Buffer) error {
	data := buffer.Bytes()
	defer packets.Put(buffer)

	data = crypt.EncodeDataInit(data)
	// calculate packet length
	length := uint16(len(data) + 2)

	s, f := byte(length>>8), byte(length&0xff)

	data = append([]byte{f, s}, data...)

	_, err := c.conn.Write(data)
	if err != nil {
		return errors.New("packet can not be sent")
	}

	return nil
}

func (c *ClientCtx) SetState(state clientState.ClientCtxState) {
	c.state = state
}

func (c *ClientCtx) GetState() clientState.ClientCtxState {
	return c.state
}

func (c *ClientCtx) CloseConnection() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *ClientCtx) SetSessionKey(sessionKey *SessionKey) {
	c.SessionKey = sessionKey
}

func (c *ClientCtx) GetRemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *ClientCtx) SetJoinedGS(isJoinedGS bool) {
	c.joinedGS = isJoinedGS
}

func (c *ClientCtx) IsJoinedGS() bool {
	return c.joinedGS
}

type noCopy struct{} //nolint:unused

func (*noCopy) Lock()   {} //nolint:unused
func (*noCopy) Unlock() {} //nolint:unused
