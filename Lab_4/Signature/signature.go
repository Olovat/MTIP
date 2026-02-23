package main

import (
	"crypto/sha256"
	"math/big"
)

// sha256Hash возвращает SHA-256(data) как *big.Int.
func sha256Hash(data []byte) *big.Int {
	sum := sha256.Sum256(data)
	return new(big.Int).SetBytes(sum[:])
}

// normalizeHash приводит хэш h к диапазону [1, N-1].
// Необходимо при малых ключах (512 бит), когда 256-битный хэш может оказаться >= N.
func normalizeHash(h, n *big.Int) *big.Int {
	nMinus1 := new(big.Int).Sub(n, big.NewInt(1))
	if h.Cmp(nMinus1) >= 0 {
		h = new(big.Int).Mod(h, nMinus1)
		h.Add(h, big.NewInt(1))
	}
	return h
}

// SignMessage формирует ЭЦП сообщения msg с помощью закрытого ключа.
//
// Алгоритм:
//  1. H = SHA-256(msg), представить как целое число
//  2. Нормализовать H: если H >= N, то H = H mod (N-1) + 1
//  3. S = H^d mod N  — подпись
func SignMessage(msg []byte, priv PrivateKey) *big.Int {
	h := sha256Hash(msg)
	h = normalizeHash(h, priv.N)
	return rsaExp(h, priv.D, priv.N) // S = H^d mod N
}

// VerifySignature проверяет подпись S сообщения msg с помощью открытого ключа.
//
// Алгоритм:
//  1. H = SHA-256(msg), нормализовать аналогично подписанию
//  2. H' = S^e mod N  — восстановленный хэш
//  3. Подпись верна, если H == H'
func VerifySignature(msg []byte, sig *big.Int, pub PublicKey) (bool, *big.Int, *big.Int) {
	h := sha256Hash(msg)
	h = normalizeHash(h, pub.N)
	hRecovered := rsaExp(sig, pub.E, pub.N) // H' = S^e mod N
	return h.Cmp(hRecovered) == 0, h, hRecovered
}
