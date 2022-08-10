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
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(f))
	return buf[:]
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
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(value))
	b.B = append(b.B, buf[:]...)
}

func (b *Buffer) WriteD(value int32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(value))
	b.B = append(b.B, buf[:]...)
}
func (b *Buffer) WriteDU(value uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	b.B = append(b.B, buf[:]...)
}

func (b *Buffer) WriteSlice(value []byte) {
	b.B = append(b.B, value...)
}

func (b *Buffer) WriteSingleByte(value byte) {
	b.B = append(b.B, value)
}

const EmptyByte byte = 0

func (b *Buffer) WriteS(value string) {
	utf16Slice := utf16.Encode([]rune(value))

	var buf []byte
	for _, v := range utf16Slice {
		if v < math.MaxInt8 {
			buf = append(buf, byte(v), 0)
		} else {
			f, s := uint8(v&0xff), uint8(v>>8)
			buf = append(buf, f, s)
		}
	}

	buf = append(buf, EmptyByte, EmptyByte)

	b.B = append(b.B, buf...)
}

//type Reader struct {
//	*bytes.Reader
//}
//
//func NewReader(buffer []byte) *Reader {
//	return &Reader{bytes.NewReader(buffer)}
//}
//
//func (r *Reader) ReadBytes(number int) []byte {
//	buffer := make([]byte, number)
//	n, _ := r.Read(buffer)
//	if n < number {
//		return []byte{}
//	}
//
//	return buffer
//}
//
//func (r *Reader) ReadInt64() int64 {
//	return int64(r.ReadUInt64())
//}
//func (r *Reader) ReadUInt64() uint64 {
//	buffer := make([]byte, 8)
//	n, err := r.Read(buffer)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if n < 8 {
//		return 0
//	}
//
//	return binary.LittleEndian.Uint64(buffer)
//}
//func (r *Reader) ReadUInt32() uint32 {
//	var result uint32
//
//	buffer := make([]byte, 4)
//	n, err := r.Read(buffer)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if n < 4 {
//		return 0
//	}
//
//	buf := bytes.NewBuffer(buffer)
//
//	err = binary.Read(buf, binary.LittleEndian, &result)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return result
//}
//
//func (r *Reader) ReadUInt16() uint16 {
//	var result uint16
//
//	buffer := make([]byte, 2)
//	n, err := r.Read(buffer)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if n < 2 {
//		return 0
//	}
//
//	buf := bytes.NewBuffer(buffer)
//
//	err = binary.Read(buf, binary.LittleEndian, &result)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return result
//}
//
//func (r *Reader) ReadUInt8() uint8 {
//	var result uint8
//
//	buffer := make([]byte, 1)
//	n, err := r.Read(buffer)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if n < 1 {
//		return 0
//	}
//
//	buf := bytes.NewBuffer(buffer)
//
//	err = binary.Read(buf, binary.LittleEndian, &result)
//	if err != nil {
//		log.Fatal(err)
//	}
//	return result
//}
//
//func (r *Reader) ReadString() string {
//	var result []byte
//	var secondByte byte
//	for {
//		firstByte, err := r.ReadByte()
//		if err != nil {
//			log.Fatal(err)
//		}
//		secondByte, err = r.ReadByte()
//		if err != nil {
//			log.Fatal(err)
//		}
//		if firstByte == 0x00 && secondByte == 0x00 {
//			break
//		} else {
//			result = append(result, firstByte, secondByte)
//		}
//	}
//
//	return string(result)
//}
//
//func (r *Reader) ReadUint32() uint32 {
//	buffer := make([]byte, 4)
//	n, err := r.Read(buffer)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if n < 4 {
//		return 0
//	}
//	return binary.LittleEndian.Uint32(buffer)
//}
