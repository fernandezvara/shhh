package ops

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/scrypt"
)

func newSalt() (salt []byte, err error) {

	salt = make([]byte, 32)
	_, err = rand.Read(salt)
	return

}

func keyFromPassword(password, salt []byte) ([]byte, error) {

	key, err := scrypt.Key(password, salt, 1048576, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	return key, nil

}

func toSHA256(text string) []byte {

	alg := sha256.New()
	alg.Write([]byte(text))
	return alg.Sum(nil)

}

func (o *Ops) encrypt(v []byte) (result []byte, err error) {

	var (
		cipherKey cipher.Block
		gcm       cipher.AEAD
		nonce     []byte
	)

	cipherKey, err = aes.NewCipher(o.key)
	if err != nil {
		return
	}

	gcm, err = cipher.NewGCM(cipherKey)
	if err != nil {
		return
	}

	nonce = make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	result = gcm.Seal(nonce, nonce, v, nil)

	return

}

func (o *Ops) decrypt(v []byte) (result []byte, err error) {

	var (
		cipherKey  cipher.Block
		gcm        cipher.AEAD
		nonceSize  int
		nonce      []byte
		ciphertext []byte
	)

	cipherKey, err = aes.NewCipher(o.key)
	if err != nil {
		return
	}

	gcm, err = cipher.NewGCM(cipherKey)
	if err != nil {
		return
	}
	nonceSize = gcm.NonceSize()

	nonce, ciphertext = v[:nonceSize], v[nonceSize:]
	result, err = gcm.Open(nil, nonce, ciphertext, nil)

	return

}
