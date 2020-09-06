package crypt

import (
	"fmt"
	"l2goserver/loginserver/crypt/blowfish"
	"log"
)

var Kek = []byte{
	0x6b,
	0x60,
	0xcb,
	0x5b,
	0x82,
	0xce,
	0x90,
	0xb1,
	0xcc,
	0x2b,
	0x6c,
	0x55,
	0x6c,
	0x6c,
	0x6c,
	0x6c,
}

func Checksum(raw []byte) bool {
	var chksum uint64

	size := len(raw)
	count := size - 4
	i := 0

	for i = 0; i < count; i += 4 {
		var ecx = (uint64(raw[i])) & 0xff
		ecx |= (uint64(raw[i+1]) << 8) & 0xff00
		ecx |= (uint64(raw[i+2]) << 0x10) & 0xff0000
		ecx |= (uint64(raw[i+3]) << 0x18) & 0xff000000
		chksum ^= ecx
	}

	var ecx = (uint64(raw[i])) & 0xff
	ecx |= (uint64(raw[i+1]) << 8) & 0xff00
	ecx |= (uint64(raw[i+2]) << 0x10) & 0xff0000
	ecx |= (uint64(raw[i+3]) << 0x18) & 0xff000000

	//raw[i] = byte(chksum)
	//raw[i+1] = byte(chksum >> 0x08)
	//raw[i+2] = byte(chksum >> 0x10)
	//raw[i+3] = byte(chksum >> 0x18)
	fmt.Println(ecx, chksum)
	return ecx == chksum
}

//func BlowfishDecrypt(encrypted, key []byte) ([]byte, error) {
//cipher, err := blowfish.NewCipher(key)
//
//if err != nil {
//	return nil, errors.New("Couldn't initialize the blowfish cipher")
//}
//
//// Check if the encrypted data is a multiple of our block size
//if len(encrypted)%8 != 0 {
//	return nil, errors.New("The encrypted data is not a multiple of the block size")
//}
//
//count := len(encrypted) / 8
//
//decrypted := make([]byte, len(encrypted))
//
//for i := 0; i < count; i++ {
//	cipher.Decrypt(decrypted[i*8:], encrypted[i*8:])
//}
//
//return decrypted, nil
//}

func BlowfishEncrypt(decrypted, key []byte) ([]byte, error) {
	//cipher, err := blowfish.NewCipher(key)

	//	if err != nil {
	//		return nil, errors.New("Couldn't initialize the blowfish cipher")
	//	}

	// Check if the decrypted data is a multiple of our block size

	//	count := len(decrypted) / 8

	//	encrypted := make([]byte, 65536)

	for i := 2; i < 176; i += 8 {
		//	cipher.Encrypt(encrypted[i*8:], decrypted[i*8:])
	}

	return decrypted, nil
}

func encXORPass(raww []byte, offset, size, key int) []byte {
	raw := make([]byte, 200, 200)
	for i := range raww {
		raw[i] = raww[i]
	}

	var stop = size - 8
	var pos = 4 + offset
	var edx int
	var ecx = key // Initial xor key
	//pos-6 stop-176 raww - 172len(171index)

	for pos < stop {
		edx = int(raw[pos] & 255)
		edx |= int(raw[pos+1]&255) << 8
		edx |= int(raw[pos+2]&255) << 16
		edx |= int(raw[pos+3]&255) << 24

		ecx += edx

		edx ^= ecx

		pos++
		raw[pos] = byte(edx & 255)
		pos++
		raw[pos] = byte((edx >> 8) & 255)
		pos++
		raw[pos] = byte((edx >> 16) & 255)
		pos++
		raw[pos] = byte((edx >> 24) & 255)

	}

	pos++
	raw[pos] = byte(ecx & 255)
	pos++
	raw[pos] = byte((ecx >> 8) & 255)
	pos++
	raw[pos] = byte((ecx >> 16) & 255)
	pos++
	raw[pos] = byte((ecx >> 24) & 255)
	return raw
}

func Enc(raw []byte) []byte {

	size := len(raw) + 15
	size = size - (size % 8) //184
	//
	data := encXORPass(raw, 2, 184, 244820523) // kak na java 181 выход
	crypt(&data, 2, 184)                       // 184 вышло без 0   .. 185 выход с00
	kek := data[2:186]
	return kek
	dataq, err := BlowfishEncrypt(data, Kek) //тут должно быть с 0 по 181 сайз 184 оффсет 2
	if err != nil {
		log.Fatal("tut error", err)
	}
	//dataq.Encrypt(data,data)
	newArray := dataq[0:184]

	return newArray
}

func crypt(raw *[]byte, offset int, size int) {
	stop := offset + size
	for i := offset; i < stop; i += 8 {
		CipherEncryptBlock(&raw, i)
	}
}

func Decrypt(raw *[]byte, offset int, size int) {
	stop := offset + size
	for i := offset; i < stop; i += 8 {
		CipherDcryptBlock(&raw, i)
	}
}

func CipherDcryptBlock(raw **[]byte, i int) {
	cipher, _ := blowfish.NewCipher(Kek)
	cipher.Decrypt(**raw, **raw, i, i)
}
func CipherEncryptBlock(raw **[]byte, i int) {
	cipher, _ := blowfish.NewCipher(Kek)
	cipher.Encrypt(**raw, **raw, i, i)
}
