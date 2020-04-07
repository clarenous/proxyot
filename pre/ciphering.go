package pre

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"io"
	"io/ioutil"
)

type EncryptClosure func(key []byte) (err error)

type DecryptClosure func(key []byte) (err error)

var (
	// errInvalidMAC occurs when Message Authentication Check (MAC) fails
	// during decryption. This happens because of either invalid private key or
	// corrupt ciphertext.
	errInvalidMAC = errors.New("invalid mac hash")

	// errInputTooShort occurs when the input ciphertext to the Decrypt
	// function is less than 134 bytes long.
	errInputTooShort = errors.New("ciphertext too short")

	errInvalidPadding = errors.New("invalid PKCS#7 padding")
)

func NewEncryptClosure(input io.Reader, output io.Writer) EncryptClosure {
	return func(key []byte) (err error) {
		// generate derived key
		derivedKey := sha512.Sum512(key)
		keyE := derivedKey[:32]
		keyM := derivedKey[32:]

		// read plaintext
		plaintext, err := ioutil.ReadAll(input)
		if err != nil {
			return err
		}

		// pad plain text
		paddedIn := addPKCSPadding(plaintext)

		// Text = IV + padded_cipher_text + HMAC-256
		text := make([]byte, aes.BlockSize+len(paddedIn)+sha256.Size)

		// read entropy as iv
		iv := text[:aes.BlockSize]
		if _, err = io.ReadFull(rand.Reader, iv); err != nil {
			return err
		}

		// start encryption
		block, err := aes.NewCipher(keyE)
		if err != nil {
			return err
		}
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(text[aes.BlockSize:len(text)-sha256.Size], paddedIn)

		// start HMAC-SHA-256
		hm := hmac.New(sha256.New, keyM)
		hm.Write(text[:len(text)-sha256.Size])          // everything is hashed
		copy(text[len(text)-sha256.Size:], hm.Sum(nil)) // write checksum

		_, err = output.Write(text)
		return
	}
}

func NewDecryptClosure(input io.Reader, output io.Writer) DecryptClosure {
	return func(key []byte) (err error) {
		text, err := ioutil.ReadAll(input)
		if err != nil {
			return err
		}

		// IV + 1 block + HMAC-256
		if len(text) < aes.BlockSize+aes.BlockSize+sha256.Size {
			return errInputTooShort
		}

		// check for cipher text length
		if (len(text)-aes.BlockSize-sha256.Size)%aes.BlockSize != 0 {
			return errInvalidPadding // not padded to 16 bytes
		}

		// generate derived key
		derivedKey := sha512.Sum512(key[:])
		keyE := derivedKey[:32]
		keyM := derivedKey[32:]

		// read hmac
		messageMAC := text[len(text)-sha256.Size:]

		// verify mac
		hm := hmac.New(sha256.New, keyM)
		hm.Write(text[:len(text)-sha256.Size]) // everything is hashed
		expectedMAC := hm.Sum(nil)
		if !hmac.Equal(messageMAC, expectedMAC) {
			return errInvalidMAC
		}

		// read iv
		iv := text[:aes.BlockSize]

		// start decryption
		block, err := aes.NewCipher(keyE)
		if err != nil {
			return err
		}
		mode := cipher.NewCBCDecrypter(block, iv)
		paddedOut := make([]byte, len(text)-aes.BlockSize-sha256.Size)
		mode.CryptBlocks(paddedOut, text[aes.BlockSize:len(text)-sha256.Size])

		plaintext, err := removePKCSPadding(paddedOut)
		if err != nil {
			return err
		}
		_, err = output.Write(plaintext)
		return
	}
}

// Implement PKCS#7 padding with block size of 16 (AES block size).

// addPKCSPadding adds padding to a block of data
func addPKCSPadding(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// removePKCSPadding removes padding from data that was added with addPKCSPadding
func removePKCSPadding(src []byte) ([]byte, error) {
	length := len(src)
	padLength := int(src[length-1])
	if padLength > aes.BlockSize || length < aes.BlockSize {
		return nil, errInvalidPadding
	}

	return src[:length-padLength], nil
}
