package models

import (
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"l2goserver/utils"
	"log"
	"net"
	"os"
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

	sRSA, err := x509.ParsePKCS1PrivateKey(readtf())
	if err != nil {
		log.Fatalln(err)
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
		State:           state.NoState,
		JoinedGS:        false,
	}
}
func savtf(bb []byte) {
	os.Create("bts")
	os.WriteFile("bts", bb, 0666)
}
func readtf() []byte {
	b, _ := os.ReadFile("bts")
	return b
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
	// Вычисление длинны пакета
	length := uint16(len(data) + 2)

	buffer := packets.Get()
	buffer.WriteHU(length)
	buffer.WriteSlice(data)

	_, err := c.Socket.Write(buffer.Bytes())
	packets.Put(buffer)
	if err != nil {
		log.Fatalln(err)
	}
	if err != nil {
		return errors.New("пакет не может быть отправлен")
	}

	return nil
}
