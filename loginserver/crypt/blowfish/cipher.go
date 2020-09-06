// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package blowfish implements Bruce Schneier's Blowfish encryption algorithm.
//
// Blowfish is a legacy cipher and its short block size makes it vulnerable to
// birthday bound attacks (see https://sweet32.info). It should only be used
// where compatibility with legacy systems, not security, is the goal.
//
// Deprecated: any new system should use AES (from crypto/aes, if necessary in
// an AEAD mode like crypto/cipher.NewGCM) or XChaCha20-Poly1305 (from
// golang.org/x/crypto/chacha20poly1305).
package blowfish // import "golang.org/x/crypto/blowfish"

// The code is a port of Bruce Schneier's C implementation.
// See https://www.schneier.com/blowfish.html.

import "strconv"

// The Blowfish block size in bytes.
const BlockSize = 8

// A Cipher is an instance of Blowfish encryption using a particular key.
type Cipher struct {
	p              [18]uint32
	s0, s1, s2, s3 [256]uint32
}

type KeySizeError int

func (k KeySizeError) Error() string {
	return "crypto/blowfish: invalid key size " + strconv.Itoa(int(k))
}

// NewCipher creates and returns a Cipher.
// The key argument should be the Blowfish key, from 1 to 56 bytes.
func NewCipher(key []byte) (*Cipher, error) {
	var result Cipher
	if k := len(key); k < 1 || k > 56 {
		return nil, KeySizeError(k)
	}
	initCipher(&result)
	ExpandKey(key, &result)
	return &result, nil
}

// NewSaltedCipher creates a returns a Cipher that folds a salt into its key
// schedule. For most purposes, NewCipher, instead of NewSaltedCipher, is
// sufficient and desirable. For bcrypt compatibility, the key can be over 56
// bytes.
func NewSaltedCipher(key, salt []byte) (*Cipher, error) {
	if len(salt) == 0 {
		return NewCipher(key)
	}
	var result Cipher
	if k := len(key); k < 1 {
		return nil, KeySizeError(k)
	}
	initCipher(&result)
	expandKeyWithSalt(key, salt, &result)
	return &result, nil
}

// BlockSize returns the Blowfish block size, 8 bytes.
// It is necessary to satisfy the Block interface in the
// package "crypto/cipher".
func (c *Cipher) BlockSize() int { return BlockSize }

// Encrypt encrypts the 8-byte buffer src using the key k
// and stores the result in dst.
// Note that for amounts of data larger than a block,
// it is not safe to just call Encrypt on successive blocks;
// instead, use an encryption mode like CBC (see crypto/cipher/cbc.go).
func (c *Cipher) Encrypt(dst, src []byte, sIndex, dIndex int) {
	l := uint32(src[sIndex+3])<<24 | uint32(src[sIndex+2])<<16 | uint32(src[sIndex+1])<<8 | uint32(src[sIndex+0])
	r := uint32(src[sIndex+7])<<24 | uint32(src[sIndex+6])<<16 | uint32(src[sIndex+5])<<8 | uint32(src[sIndex+4])
	l, r = encryptBlock(l, r, c)
	c.bits32ToBytes(int(r), dst, dIndex)
	c.bits32ToBytes(int(l), dst, dIndex+4)
	//	dst[3], dst[2], dst[1], dst[0] = byte(l>>24), byte(l>>16), byte(l>>8), byte(l)
	//	dst[7], dst[6], dst[5], dst[4] = byte(r>>24), byte(r>>16), byte(r>>8), byte(r)
}

func (c *Cipher) bits32ToBytes(in int, dst []byte, dstIndex int) {
	dst[dstIndex] = byte(in)
	dst[dstIndex+1] = byte(in >> 8)
	dst[dstIndex+2] = byte(in >> 16)
	dst[dstIndex+3] = byte(in >> 24)
}

// Decrypt decrypts the 8-byte buffer src using the key k
// and stores the result in dst.
func (c *Cipher) Decrypt(dst, src []byte, sIndex, dIndex int) {
	l := uint32(src[sIndex+3])<<24 | uint32(src[sIndex+2])<<16 | uint32(src[sIndex+1])<<8 | uint32(src[sIndex+0])
	r := uint32(src[sIndex+7])<<24 | uint32(src[sIndex+6])<<16 | uint32(src[sIndex+5])<<8 | uint32(src[sIndex+4])
	l, r = decryptBlock(l, r, c)
	c.bits32ToBytes(int(r), dst, dIndex)
	c.bits32ToBytes(int(l), dst, dIndex+4)
	//	dst[0], dst[1], dst[2], dst[3] = byte(l>>24), byte(l>>16), byte(l>>8), byte(l)
	//	dst[4], dst[5], dst[6], dst[7] = byte(r>>24), byte(r>>16), byte(r>>8), byte(r)
}

func initCipher(c *Cipher) {
	copy(c.p[0:], p[0:])
	copy(c.s0[0:], s0[0:])
	copy(c.s1[0:], s1[0:])
	copy(c.s2[0:], s2[0:])
	copy(c.s3[0:], s3[0:])
}
