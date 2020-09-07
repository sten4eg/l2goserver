package crypt

import (
	"fmt"
	"l2goserver/loginserver/crypt/blowfish"
)

var StaticBlowfish = []byte{
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

func encXORPass(raww []byte, offset, size, key int) []byte {
	raw := make([]byte, 200)
	copy(raw[:], raww[:])

	var stop = size - 8
	var pos = 4 + offset
	var edx int
	var ecx = key // Initial xor key
	//pos-6 stop-176 raww - 172len(171index)

	for pos < stop {
		edx = int(raw[pos])
		edx |= int(raw[pos+1]) << 8
		edx |= int(raw[pos+2]) << 16
		edx |= int(raw[pos+3]) << 24

		ecx += edx

		edx ^= ecx

		pos++
		raw[pos] = byte(edx)
		pos++
		raw[pos] = byte(edx >> 8)
		pos++
		raw[pos] = byte(edx >> 16)
		pos++
		raw[pos] = byte(edx >> 24)

	}

	pos++
	raw[pos] = byte(ecx)
	pos++
	raw[pos] = byte(ecx >> 8)
	pos++
	raw[pos] = byte(ecx >> 16)
	raw[pos] = byte(ecx >> 24)
	return raw
}

func EncodeData(raw []byte) []byte {

	size := len(raw) + 15
	size = size - (size % 8) //184
	//
	kek := raw[0:171]
	data := encXORPass(kek, 2, size, 244820523) //выход инд181 с 000 вход 170индекс последнего значащего числа
	crypt(&data, 2, size)                       //  .. 185 выход с00
	return data[2:186]
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
		CipherDecryptBlock(&raw, i)
	}
}

func CipherDecryptBlock(raw **[]byte, i int) {
	cipher, _ := blowfish.NewCipher(StaticBlowfish)
	cipher.Decrypt(**raw, **raw, i, i)
}
func CipherEncryptBlock(raw **[]byte, i int) {
	cipher, _ := blowfish.NewCipher(StaticBlowfish)
	cipher.Encrypt(**raw, **raw, i, i)
}
