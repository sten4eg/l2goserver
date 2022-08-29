package packets

import (
	"encoding/binary"
	"math"
	"unicode/utf16"
)

type Buffer struct {
	B []byte
}

func (b *Buffer) Len() int {
	return len(b.B)
}

func (b *Buffer) Bytes() []byte {
	return b.B
}
func (b *Buffer) CopyBytes() []byte {
	m := make([]byte, b.Len()+1)
	_ = copy(m, b.B)
	return m
}
func (b *Buffer) Reset() {
	b.B = b.B[:0]
}

func float64ToByte(f float64) []byte {
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[0:], math.Float64bits(f))
	return buffer[0:]
}

func (b *Buffer) WriteF(value float64) {
	b.B = append(b.B, float64ToByte(value)...)
}

func (b *Buffer) WriteH(value int16) {
	b.B = append(b.B, byte(value&0xff), byte(value>>8))
}

func (b *Buffer) WriteHU(value uint16) {
	b.B = append(b.B, byte(value&0xff), byte(value>>8))
}

func (b *Buffer) WriteQ(value int64) {
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[:], uint64(value))
	b.B = append(b.B, buffer[0:]...)
}

func (b *Buffer) WriteD(value int32) {
	var buffer [4]byte
	binary.LittleEndian.PutUint32(buffer[:], uint32(value))
	b.B = append(b.B, buffer[0:]...)
}
func (b *Buffer) WriteDU(value uint32) {
	var buffer [4]byte
	binary.LittleEndian.PutUint32(buffer[:], value)
	b.B = append(b.B, buffer[0:]...)
}

func (b *Buffer) WriteSlice(value []byte) {
	b.B = append(b.B, value...)
}
func (b *Buffer) WriteSliceTest(value []byte) {
	b.B = append(b.B, value[0:]...)
}

func (b *Buffer) WriteSingleByte(value byte) {
	b.B = append(b.B, value)
}

const EmptyByte byte = 0

func (b *Buffer) WriteS(value string) {
	utf16Slice := utf16.Encode([]rune(value))

	var buffer []byte
	for _, v := range utf16Slice {
		if v < math.MaxInt8 {
			buffer = append(buffer, byte(v), 0)
		} else {
			f, s := uint8(v&0xff), uint8(v>>8)
			buffer = append(buffer, f, s)
		}
	}

	buffer = append(buffer, EmptyByte, EmptyByte)

	b.B = append(b.B, buffer...)
}
