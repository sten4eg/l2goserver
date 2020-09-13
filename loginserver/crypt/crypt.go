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

var isStatic = true

func verifyChecksum(raw []byte, size int) bool {
	var checksum int64
	count := size - 4
	var i int

	for i = 0; i < count; i += 4 {
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

func appendchecksum(raw []byte, offset, size int) []byte {
	var chksum int64
	var count = size - 4
	var ecx int64
	var i int

	for i = offset; i < count; i += 4 {
		var ecx = int64(raw[i])
		ecx |= (int64(raw[i+1]) << 8) & 0xff00
		ecx |= (int64(raw[i+2]) << 0x10) & 0xff0000
		ecx |= (int64(raw[i+3]) << 0x18) & 0xff000000
		chksum ^= ecx
	}

	ecx = int64(raw[i] & 0xff)
	ecx |= int64(raw[i+1]<<8) & 0xff00
	ecx |= int64(raw[i+2]<<0x10) & 0xff0000
	ecx |= int64(raw[i+3]<<0x18) & 0xff000000

	raw[i] = (byte)(chksum & 0xff)
	raw[i+1] = (byte)((chksum >> 0x08) & 0xff)
	raw[i+2] = (byte)((chksum >> 0x10) & 0xff)
	raw[i+3] = (byte)((chksum >> 0x18) & 0xff)
	return raw
}
func encXORPass(raww []byte, size, key int) []byte {
	raw := make([]byte, 200)
	copy(raw[:], raww[:])

	var stop = size - 8
	var pos = 4
	var edx int
	var ecx = key // Initial xor key

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

	size := len(raw) + 4 // reserve checksum
	var data []byte
	if isStatic {

		size += 4                      // reserve for XOR "key"
		size = (size + 8) - (size % 8) // padding

		data = encXORPass(raw, size, rand.Int()) //Xor
		crypt(&data, size)                       //blowfish
		isStatic = false
	} else {
		size = (size + 8) - (size % 8) // padding
		appendchecksum(raw, 2, size)
		crypt(&data, size)
		data = []byte{1}
	}

	return data[:size]
}

func DecodeData(raw []byte) []byte {
	size := len(raw) - 2 // minus length package
	raww := make([]byte, 200)
	copy(raww[:], raw[2:])
	decrypt(&raww, size) //size 40
	valid := verifyChecksum(raww, size)
	if !valid {
		log.Fatal("not verifiedCheckSum")
	}
	return raww
}

func crypt(raw *[]byte, size int) {
	cipher, _ := blowfish.NewCipher(StaticBlowfish)
	for i := 0; i < size; i += 8 {
		cipher.Encrypt(*raw, *raw, i, i)
	}
}

func decrypt(raw *[]byte, size int) {
	cipher, _ := blowfish.NewCipher(StaticBlowfish)
	for i := 0; i < size; i += 8 {
		cipher.Decrypt(*raw, *raw, i, i)

	}
}
