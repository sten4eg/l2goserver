package utils

import (
	"bytes"
	"golang.org/x/exp/constraints"
	"reflect"
	"unsafe"
)

func B2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func S2b(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

func Contains[T constraints.Integer](slice []T, need T) bool {
	for i := range slice {
		if slice[i] == need {
			return true
		}
	}
	return false
}

func CompareHexId(hexId []byte, hexIds [][]byte) bool {
	for i := range hexIds {
		if bytes.Equal(hexId, hexIds[i]) {
			return true
		}
	}
	return false
}
