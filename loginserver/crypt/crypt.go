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

var IsStatic = true

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
	log.Println(checksum)
	log.Println(ecx)
	return ecx == checksum
}

func appendchecksum(raw []byte, size int) []byte {
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

func EncodeData(raw []byte, blowfishKey []byte) []byte {

	size := len(raw) + 4 // reserve checksum
	data := make([]byte, 200)
	copy(data, raw)
	if IsStatic {

		size += 4                      // reserve for XOR "key"
		size = (size + 8) - (size % 8) // padding

		data = encXORPass(data, size, rand.Int()) // Xor
		crypt(&data, size, StaticBlowfish)        // blowfish
		IsStatic = false
	} else {
		size = (size + 8) - (size % 8) // padding
		appendchecksum(data, size)
		crypt(&data, size, blowfishKey)
	}

	return data[:size]
}

func DecodeData(raw []byte, blowfishKey []byte) []byte {
	size := len(raw) - 2 // minus length package
	data := make([]byte, 200)
	copy(data, raw[2:])
	decrypt(&data, size, blowfishKey)

	valid := verifyChecksum(data, size)
	if !valid {
		log.Println("not verifiedCheckSum")
	}
	return data
}

func crypt(raw *[]byte, size int, blowfishKey []byte) {
	cipher, _ := blowfish.NewCipher(blowfishKey)
	for i := 0; i < size; i += 8 {
		cipher.Encrypt(*raw, *raw, i, i)
	}
}

func decrypt(raw *[]byte, size int, blowfishKey []byte) {
	cipher, _ := blowfish.NewCipher(blowfishKey)
	for i := 0; i < size; i += 8 {
		cipher.Decrypt(*raw, *raw, i, i)
	}
}

func ScrambleModulus(modulus []byte) []byte {

	scrambledMod := modulus
	var temp []byte
	copy(temp, scrambledMod)
	//	System.arraycopy(scrambledMod, 1, temp, 0, 0x80)
	//	scrambledMod = temp

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
