package utils

import (
	"bytes"
	"golang.org/x/exp/constraints"
)

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
