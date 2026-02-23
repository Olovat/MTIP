package main

import (
	"math/big"

	"rsacore"
)

// Псевдонимы типов — полностью совместимы с оригиналом
type PublicKey = rsacore.PublicKey
type PrivateKey = rsacore.PrivateKey
type KeyPair = rsacore.KeyPair

// Делегирующие функции
func GenerateKeyPair(bits int) (KeyPair, error) { return rsacore.GenerateKeyPair(bits) }

// rsaExp — используется в signature
func rsaExp(x, exp, n *big.Int) *big.Int { return rsacore.RsaExp(x, exp, n) }
