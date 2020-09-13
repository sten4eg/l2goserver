package crypt

import (
	"l2goserver/loginserver/crypt/blowfish"
	"log"
	"math/rand"
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

func verifyChecksum(raw []byte, offset, size int) bool {
	var checksum int64
	count := size - 4
	var i int

	for i = offset; i < count; i += 4 {
		var ecx = int64(raw[i])
		ecx |= (int64(raw[i+1]) << 8) & 0xff00
		ecx |= (int64(raw[i+2]) << 0x10) & 0xff0000
		ecx |= (int64(raw[i+3]) << 0x18) & 0xff000000
		checksum ^= ecx
	}

	var ecx = int64(raw[i])
	ecx |= (int64(raw[i+1]) << 8) & 0xff00
	ecx |= (int64(raw[i+2]) << 0x10) & 0xff0000
	ecx |= (int64(raw[i+3]) << 0x18) & 0xff000000

	return ecx == checksum
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

		raw[pos] = byte(edx)
		pos++
		raw[pos] = byte(edx >> 8)
		pos++
		raw[pos] = byte(edx >> 16)
		pos++
		raw[pos] = byte(edx >> 24)
		pos++
	}

	raw[pos] = byte(ecx)
	pos++
	raw[pos] = byte(ecx >> 8)
	pos++
	raw[pos] = byte(ecx >> 16)
	pos++
	raw[pos] = byte(ecx >> 24)
	return raw
}

func EncodeData(raw []byte) []byte {

	size := len(raw) + 15
	size = size - (size % 8) //184
	//
	data := encXORPass(raw, 2, size, rand.Int()) //выход инд181 с 000 вход 170индекс последнего значащего числа
	crypt(&data, 2, size)                        //  .. 185 выход с00
	return data[2:186]
}

func DecodeData(raw []byte) []byte {
	raww := make([]byte, 200)
	copy(raww[:], raw[:])
	decrypt(&raww, 2, 40) //size 40 , offset 2
	valid := verifyChecksum(raww, 2, 40)
	if !valid {
		log.Fatal("not verifiedCheckSum")
	}
	return raww
}

func crypt(raw *[]byte, offset int, size int) {
	stop := offset + size
	cipher, _ := blowfish.NewCipher(StaticBlowfish)
	for i := offset; i < stop; i += 8 {
		cipher.Encrypt(*raw, *raw, i, i)
	}
}

func decrypt(raw *[]byte, offset int, size int) {
	stop := offset + size
	cipher, _ := blowfish.NewCipher(StaticBlowfish)
	for i := offset; i < stop; i += 8 {
		cipher.Decrypt(*raw, *raw, i, i)

	}
}
