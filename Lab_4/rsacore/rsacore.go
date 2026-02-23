// rsacore.go — общие примитивы RSA для программ шифрования и ЭЦП
// Используется Lab_4/RSA и Lab_4/Signature как локальная зависимость
package rsacore

import (
	"crypto/rand"
	"errors"
	"math/big"
)

// PublicKey — открытый ключ RSA (e, N)
type PublicKey struct {
	N *big.Int
	E *big.Int
}

// PrivateKey — закрытый ключ RSA (d, N)
type PrivateKey struct {
	N *big.Int
	D *big.Int
}

// KeyPair объединяет оба ключа
type KeyPair struct {
	Public  PublicKey
	Private PrivateKey
}

// GenerateKeyPair генерирует RSA-ключи длиной bits бит (e = 65537)
func GenerateKeyPair(bits int) (KeyPair, error) {
	e := big.NewInt(65537)
	var p, q *big.Int
	var err error

	for {
		p, err = rand.Prime(rand.Reader, bits/2)
		if err != nil {
			return KeyPair{}, err
		}
		for {
			q, err = rand.Prime(rand.Reader, bits/2)
			if err != nil {
				return KeyPair{}, err
			}
			if p.Cmp(q) != 0 {
				break
			}
		}

		n := new(big.Int).Mul(p, q)
		phi := new(big.Int).Mul(
			new(big.Int).Sub(p, big.NewInt(1)),
			new(big.Int).Sub(q, big.NewInt(1)),
		)

		if new(big.Int).GCD(nil, nil, e, phi).Cmp(big.NewInt(1)) != 0 {
			continue
		}
		d := new(big.Int).ModInverse(e, phi)
		if d == nil {
			continue
		}
		return KeyPair{
			Public:  PublicKey{N: n, E: new(big.Int).Set(e)},
			Private: PrivateKey{N: n, D: d},
		}, nil
	}
}

// RsaExp вычисляет универсальный RSA-примитив
func RsaExp(x, exp, n *big.Int) *big.Int {
	return new(big.Int).Exp(x, exp, n)
}

// EncryptInt
func EncryptInt(m *big.Int, pub PublicKey) *big.Int {
	return new(big.Int).Exp(m, pub.E, pub.N)
}

// DecryptInt
func DecryptInt(c *big.Int, priv PrivateKey) *big.Int {
	return new(big.Int).Exp(c, priv.D, priv.N)
}

// blockSize — макс. байт в открытом блоке.
func blockSize(n *big.Int) int {
	return (n.BitLen() - 1) / 8
}

// EncryptBytes шифрует произвольные данные блочным RSA
func EncryptBytes(data []byte, pub PublicKey) ([][]byte, error) {
	bs := blockSize(pub.N)
	if bs < 1 {
		return nil, errors.New("ключ слишком мал")
	}

	length := len(data)
	encoded := make([]byte, 4+length)
	encoded[0] = byte(length >> 24)
	encoded[1] = byte(length >> 16)
	encoded[2] = byte(length >> 8)
	encoded[3] = byte(length)
	copy(encoded[4:], data)

	outBlockSize := (pub.N.BitLen() + 7) / 8
	var blocks [][]byte

	for i := 0; i < len(encoded); i += bs {
		end := i + bs
		if end > len(encoded) {
			end = len(encoded)
		}
		m := new(big.Int).SetBytes(encoded[i:end])
		c := EncryptInt(m, pub)

		cb := c.Bytes()
		block := make([]byte, outBlockSize)
		copy(block[outBlockSize-len(cb):], cb)
		blocks = append(blocks, block)
	}
	return blocks, nil
}

// DecryptBytes дешифрует блоки, полученные от EncryptBytes
func DecryptBytes(blocks [][]byte, priv PrivateKey) ([]byte, error) {
	bs := blockSize(priv.N)
	if bs < 1 {
		return nil, errors.New("ключ слишком мал")
	}

	var decoded []byte
	for _, block := range blocks {
		c := new(big.Int).SetBytes(block)
		m := DecryptInt(c, priv)
		mb := m.Bytes()
		chunk := make([]byte, bs)
		copy(chunk[bs-len(mb):], mb)
		decoded = append(decoded, chunk...)
	}

	if len(decoded) < 4 {
		return nil, errors.New("повреждённые данные: слишком коротко")
	}
	length := int(decoded[0])<<24 | int(decoded[1])<<16 | int(decoded[2])<<8 | int(decoded[3])
	if length < 0 || 4+length > len(decoded) {
		return nil, errors.New("повреждённые данные: неверная длина")
	}
	return decoded[4 : 4+length], nil
}
