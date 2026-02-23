package main

import (
	"math/big"

	"rsacore"
)

// Псевдонимы типов — полностью совместимы с оригиналом
type PublicKey = rsacore.PublicKey
type PrivateKey = rsacore.PrivateKey
type KeyPair = rsacore.KeyPair

// Делегирующие функции — main.go вызывает их напрямую без префикса пакета
func GenerateKeyPair(bits int) (KeyPair, error)       { return rsacore.GenerateKeyPair(bits) }
func EncryptInt(m *big.Int, pub PublicKey) *big.Int   { return rsacore.EncryptInt(m, pub) }
func DecryptInt(c *big.Int, priv PrivateKey) *big.Int { return rsacore.DecryptInt(c, priv) }
func EncryptBytes(data []byte, pub PublicKey) ([][]byte, error) {
	return rsacore.EncryptBytes(data, pub)
}
func DecryptBytes(blocks [][]byte, priv PrivateKey) ([]byte, error) {
	return rsacore.DecryptBytes(blocks, priv)
}
