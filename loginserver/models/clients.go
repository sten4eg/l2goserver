package models

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"errors"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"l2goserver/utils"
	"math/rand"
	"net"
	"runtime/trace"
	"strconv"
	"sync"
)

type Clients struct {
	C  []*ClientCtx
	Mu sync.Mutex
}
type ClientCtx struct {
	noCopy          utils.NoCopy //nolint:unused,structcheck
	Account         Account
	SessionID       uint32
	conn            *net.TCPConn
	ScrambleModulus []byte
	SessionKey      *SessionKey
	PrivateKey      *rsa.PrivateKey
	BlowFish        []byte
	state           state.GameState
	JoinedGS        bool
	Uid             uint64
	isStatic        bool
}

type SessionKey struct {
	PlayOk1  uint32
	PlayOk2  uint32
	LoginOk1 uint32
	LoginOk2 uint32
}

//go:embed bts
var b []byte

func NewClient() (*ClientCtx, error) {
	//id := rand.Uint32()

	var id uint32 = 2596996162
	//sk := SessionKey{
	//	PlayOk1:  rand.Uint32(),
	//	PlayOk2:  rand.Uint32(),
	//	LoginOk1: rand.Uint32(),
	//	LoginOk2: rand.Uint32(),
	//}

	sk := SessionKey{
		PlayOk1:  4039455774,
		PlayOk2:  2854263694,
		LoginOk1: 1879968118,
		LoginOk2: 1823804162,
	}

	//blowfish := make([]byte, 16)
	blowfish := []byte{216, 104, 29, 13, 134, 209, 233, 30, 0, 22, 121, 57, 203, 102, 148, 210}
	//_, _ = rand.Read(blowfish)

	//privateKey, err := rsa.GenerateKey(crand.Reader, 1024)
	//if err != nil {
	//	fmt.Println(err)
	//}

	sRSA, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		return nil, err
	}
	_ = sRSA
	scrambleModulus := crypt.ScrambleModulus(sRSA.PublicKey.N.Bytes())

	//scrambleModulus := []byte{134, 142, 95, 160, 18, 252, 106, 59, 228, 254, 60, 14, 60, 2, 90, 106, 224, 241, 174, 178, 47, 66, 122, 21, 110, 215, 76, 146, 27, 182, 122, 150, 1, 134, 164, 255, 126, 28, 105, 76, 133, 192, 162, 208, 233, 9, 184, 101, 194, 45, 164, 247, 101, 2, 210, 212, 118, 99, 115, 43, 231, 32, 183, 49, 136, 115, 208, 243, 39, 171, 54, 233, 219, 240, 167, 155, 202, 241, 240, 210, 1, 247, 75, 86, 226, 199, 41, 87, 111, 247, 168, 33, 182, 40, 202, 11, 189, 174, 210, 199, 242, 41, 127, 49, 208, 44, 221, 72, 240, 95, 21, 2, 195, 222, 83, 6, 225, 251, 182, 0, 179, 43, 149, 226, 56, 43, 3, 2}
	return &ClientCtx{
		SessionID:       id,
		SessionKey:      &sk,
		BlowFish:        blowfish,
		PrivateKey:      sRSA,
		ScrambleModulus: scrambleModulus,
		state:           state.NoState,
		JoinedGS:        false,
		Uid:             rand.Uint64(),
		isStatic:        true,
	}, nil
}

func (c *ClientCtx) SetConn(conn *net.TCPConn) {
	c.conn = conn
}
func (c *ClientCtx) Receive() (uint8, []byte, error) {
	header := make([]byte, 2)
	reg := trace.StartRegion(context.Background(), "readHeader")
	n, err := c.conn.Read(header)
	if err != nil {
		return 0, nil, err
	}
	if n != 2 {
		return 0, nil, errors.New("Ожидалось 2 байта длинны, получено: " + strconv.Itoa(n))
	}
	reg.End()
	// длинна пакета
	dataSize := (int(header[0]) | int(header[1])<<8) - 2

	// аллокация требуемого массива байт для входящего пакета
	data := make([]byte, dataSize)

	reg = trace.StartRegion(context.Background(), "readData")

	n, err = c.conn.Read(data)
	if n != dataSize || err != nil {
		return 0, nil, errors.New("длинна прочитанного пакета не соответствует требуемому размеру")
	}
	reg.End()

	fullPackage := make([]byte, 0, len(header)+len(data))
	fullPackage = append(fullPackage, header...)
	fullPackage = append(fullPackage, data...)

	fullPackage = crypt.DecodeData(fullPackage, c.BlowFish)

	opcode := fullPackage[0]

	return opcode, fullPackage[1:], nil
}

func (c *ClientCtx) SendBuf(buffer *packets.Buffer) error {
	data := buffer.Bytes()
	defer packets.Put(buffer)

	data = crypt.EncodeData(data, c.BlowFish, c.isStatic)
	// Вычисление длинны пакета
	length := uint16(len(data) + 2)

	toSend := packets.Get()
	toSend.WriteHU(length)
	toSend.WriteSlice(data) //TODO очень много выделяет
	defer packets.Put(toSend)

	_, err := c.conn.Write(toSend.Bytes())
	if err != nil {
		return errors.New("пакет не может быть отправлен")
	}
	if err != nil {
		return errors.New("пакет не может быть отправлен")
	}

	return nil
}

func (c *ClientCtx) SetState(state state.GameState) {
	c.state = state
}

func (c *ClientCtx) GetState() state.GameState {
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

func (c *ClientCtx) GetLocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *ClientCtx) SetStaticFalse() {
	c.isStatic = false
}
