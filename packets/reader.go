package packets

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
)

type Reader struct {
	r *bytes.Reader
}

func NewReader(buffer []byte) *Reader {
	return &Reader{
		r: bytes.NewReader(buffer),
	}
}

func (r *Reader) CurrentIndex() int64 {
	l := r.r.Len()
	s := r.r.Size()
	currIndex := s - int64(l)
	return currIndex
}

func (r *Reader) UnreadBytes(b int) {
	for i := 0; i < b; i++ {
		err := r.r.UnreadByte()
		if err != nil {
			panic(err)
		}
	}
}
func (r *Reader) ReadBytes(number int) []byte {
	buffer := make([]byte, number)
	n, err := r.r.Read(buffer)
	if err != nil {
		panic(err)
	}
	if n < number {
		panic("n<number")
	}

	return buffer
}
func (r *Reader) ReadSingleByte() byte {
	n, err := r.r.ReadByte()
	if err != nil {
		panic(err)
	}
	return n
}
func (r *Reader) ReadUInt8() uint8 {
	return r.ReadSingleByte()
}
func (r *Reader) ReadInt8() int8 {
	return int8(r.ReadSingleByte())
}
func (r *Reader) ReadUInt16() uint16 {
	buffer := make([]byte, 2)
	n, err := r.r.Read(buffer)
	if err != nil {
		panic(err)
	}
	if n < 2 {
		return 0
	}

	return binary.LittleEndian.Uint16(buffer)
}
func (r *Reader) ReadInt16() int16 {
	buffer := make([]byte, 2)
	n, err := r.r.Read(buffer)
	if err != nil {
		panic(err)
	}
	if n < 2 {
		return 0
	}

	return int16(binary.LittleEndian.Uint16(buffer))
}
func (r *Reader) ReadUInt32() uint32 {
	buffer := make([]byte, 4)
	n, err := r.r.Read(buffer)
	if err != nil {
		panic(err)
	}
	if n < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(buffer)
}
func (r *Reader) ReadInt32() int32 {
	buffer := make([]byte, 4)
	n, err := r.r.Read(buffer)
	if err != nil {
		panic(err)
	}
	if n < 4 {
		return 0
	}

	return int32(binary.LittleEndian.Uint32(buffer))
}
func (r *Reader) ReadUInt64() uint64 {
	buffer := make([]byte, 8)
	n, err := r.r.Read(buffer)
	if err != nil {
		panic(err)
	}
	if n < 8 {
		return 0
	}

	return binary.LittleEndian.Uint64(buffer)
}
func (r *Reader) ReadInt64() int64 {
	return int64(r.ReadUInt64())
}
func (r *Reader) ReadString() string {
	var result []uint16
	buffer := make([]byte, 2)
	for {
		_, err := r.r.Read(buffer)
		if err != nil {
			panic(err)
		}
		q := binary.LittleEndian.Uint16(buffer)

		if q == 0 {
			break
		}
		result = append(result, q)

	}
	return string(utf16.Decode(result))
}
