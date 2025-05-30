package crypt

import (
	"l2goserver/crypt/blowfish"
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

func VerifyCheckSum(raw []byte) bool {
	size := len(raw)
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

func AppendCheckSum(raw []byte, size int) []byte {
	var chksum int64
	var count = size - 4
	var i int

	for i = 0; i < count; i += 4 {
		var ecx = int64(raw[i])
		ecx |= (int64(raw[i+1]) << 8) & 0xff00
		ecx |= (int64(raw[i+2]) << 0x10) & 0xff0000
		ecx |= (int64(raw[i+3]) << 0x18) & 0xff000000
		chksum ^= ecx
	}

	raw[i] = (byte)(chksum & 0xff)
	raw[i+1] = (byte)((chksum >> 0x08) & 0xff)
	raw[i+2] = (byte)((chksum >> 0x10) & 0xff)
	raw[i+3] = (byte)((chksum >> 0x18) & 0xff)
	return raw
}

func encXORPass(raw []byte, size, key int) []byte {

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

func EncodeDataInit(raw []byte) (data []byte) {
	size := len(raw) + 4 // reserve checksum

	size += 4                      // reserve for XOR "key"
	size = (size + 8) - (size % 8) // padding
	data = make([]byte, size)
	copy(data, raw)

	data = encXORPass(data, size, rand.Int()) // Xor
	crypt(data, size, StaticBlowfish)         // blowfish

	return
}

func EncodeData(raw []byte, blowfishKey []byte) (data []byte) {
	size := len(raw) + 4 // reserve checksum

	size = (size + 8) - (size % 8) // padding

	data = make([]byte, size)
	copy(data, raw)

	AppendCheckSum(data, size)
	crypt(data, size, blowfishKey)

	return
}

func DecodeData(data []byte, blowfishKey []byte) bool {
	if len(data) == 0 {
		return false
	}

	decrypt(data, blowfishKey)

	if !VerifyCheckSum(data) {
		log.Println("checksum verification failed")
		return false
	}

	return true
}
func crypt(raw []byte, size int, blowfishKey []byte) {
	cipher, _ := blowfish.NewCipher(blowfishKey)
	for i := 0; i < size; i += 8 {
		cipher.Encrypt(raw, raw, i, i)
	}
}

func decrypt(raw []byte, blowfishKey []byte) {
	cipher, _ := blowfish.NewCipher(blowfishKey)
	size := len(raw)
	for i := 0; i < size; i += 8 {
		cipher.Decrypt(raw, raw, i, i)
	}
}

func ScrambleModulus(modulus []byte) []byte {

	scrambledMod := modulus
	var temp []byte
	copy(temp, scrambledMod)

	// step 1 : 0x4d-0x50 <-> 0x00-0x04

	for i := 0; i < 4; i++ {
		scrambledMod[0x00+i], scrambledMod[0x4d+i] = scrambledMod[0x4d+i], scrambledMod[0x00+i]
	}
	// step 2 : xor first 0x40 bytes with last 0x40 bytes
	for i := 0; i < 0x40; i++ {
		scrambledMod[i] = (byte)(scrambledMod[i] ^ scrambledMod[0x40+i])
	}
	// step 3 : xor bytes 0x0d-0x10 with bytes 0x34-0x38
	for i := 0; i < 4; i++ {
		scrambledMod[0x0d+i] = (byte)(scrambledMod[0x0d+i] ^ scrambledMod[0x34+i])
	}
	// step 4 : xor last 0x40 bytes with first 0x40 bytes
	for i := 0; i < 0x40; i++ {
		scrambledMod[0x40+i] = (byte)(scrambledMod[0x40+i] ^ scrambledMod[i])
	}
	return scrambledMod
}
