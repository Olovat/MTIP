package main

// Таблицы

// pi — прямая таблица замены S (ГОСТ Р 34.12-2015, приложение А) я хз чё тут происходит, она с хабра украдена
var pi = [256]byte{
	0xFC, 0xEE, 0xDD, 0x11, 0xCF, 0x6E, 0x31, 0x16,
	0xFB, 0xC4, 0xFA, 0xDA, 0x23, 0xC5, 0x04, 0x4D,
	0xE9, 0x77, 0xF0, 0xDB, 0x93, 0x2E, 0x99, 0xBA,
	0x17, 0x36, 0xF1, 0xBB, 0x14, 0xCD, 0x5F, 0xC1,
	0xF9, 0x18, 0x65, 0x5A, 0xE2, 0x5C, 0xEF, 0x21,
	0x81, 0x1C, 0x3C, 0x42, 0x8B, 0x01, 0x8E, 0x4F,
	0x05, 0x84, 0x02, 0xAE, 0xE3, 0x6A, 0x8F, 0xA0,
	0x06, 0x0B, 0xED, 0x98, 0x7F, 0xD4, 0xD3, 0x1F,
	0xEB, 0x34, 0x2C, 0x51, 0xEA, 0xC8, 0x48, 0xAB,
	0xF2, 0x2A, 0x68, 0xA2, 0xFD, 0x3A, 0xCE, 0xCC,
	0xB5, 0x70, 0x0E, 0x56, 0x08, 0x0C, 0x76, 0x12,
	0xBF, 0x72, 0x13, 0x47, 0x9C, 0xB7, 0x5D, 0x87,
	0x15, 0xA1, 0x96, 0x29, 0x10, 0x7B, 0x9A, 0xC7,
	0xF3, 0x91, 0x78, 0x6F, 0x9D, 0x9E, 0xB2, 0xB1,
	0x32, 0x75, 0x19, 0x3D, 0xFF, 0x35, 0x8A, 0x7E,
	0x6D, 0x54, 0xC6, 0x80, 0xC3, 0xBD, 0x0D, 0x57,
	0xDF, 0xF5, 0x24, 0xA9, 0x3E, 0xA8, 0x43, 0xC9,
	0xD7, 0x79, 0xD6, 0xF6, 0x7C, 0x22, 0xB9, 0x03,
	0xE0, 0x0F, 0xEC, 0xDE, 0x7A, 0x94, 0xB0, 0xBC,
	0xDC, 0xE8, 0x28, 0x50, 0x4E, 0x33, 0x0A, 0x4A,
	0xA7, 0x97, 0x60, 0x73, 0x1E, 0x00, 0x62, 0x44,
	0x1A, 0xB8, 0x38, 0x82, 0x64, 0x9F, 0x26, 0x41,
	0xAD, 0x45, 0x46, 0x92, 0x27, 0x5E, 0x55, 0x2F,
	0x8C, 0xA3, 0xA5, 0x7D, 0x69, 0xD5, 0x95, 0x3B,
	0x07, 0x58, 0xB3, 0x40, 0x86, 0xAC, 0x1D, 0xF7,
	0x30, 0x37, 0x6B, 0xE4, 0x88, 0xD9, 0xE7, 0x89,
	0xE1, 0x1B, 0x83, 0x49, 0x4C, 0x3F, 0xF8, 0xFE,
	0x8D, 0x53, 0xAA, 0x90, 0xCA, 0xD8, 0x85, 0x61,
	0x20, 0x71, 0x67, 0xA4, 0x2D, 0x2B, 0x09, 0x5B,
	0xCB, 0x9B, 0x25, 0xD0, 0xBE, 0xE5, 0x6C, 0x52,
	0x59, 0xA6, 0x74, 0xD2, 0xE6, 0xF4, 0xB4, 0xC0,
	0xD1, 0x66, 0xAF, 0xC2, 0x39, 0x4B, 0x63, 0xB6,
}

// piInv — обратная таблица замены S
var piInv [256]byte

func init() {
	for i := 0; i < 256; i++ {
		piInv[pi[i]] = byte(i)
	}
}

// lCoeff — коэффициенты линейного преобразования L в порядке big-endian хранения.

var lCoeff = [16]byte{
	0x94, 0x20, 0x85, 0x10, 0xC2, 0xC0, 0x01, 0xFB,
	0x01, 0xC0, 0xC2, 0x10, 0x85, 0x20, 0x94, 0x01,
}

// Арифметика в GF(2^8) — неприводимый многочлен p(x) = x^8+x^7+x^6+x+1 (0x1C3)

func gfMul(a, b byte) byte {
	var result byte
	for b != 0 {
		if b&1 != 0 {
			result ^= a
		}
		if a&0x80 != 0 {
			a = (a << 1) ^ 0xC3
		} else {
			a <<= 1
		}
		b >>= 1
	}
	return result
}

// Примитивные преобразования

// xorBlock — X[k]: побайтовый XOR блока с ключевым словом
func xorBlock(a, b [16]byte) [16]byte {
	var out [16]byte
	for i := range a {
		out[i] = a[i] ^ b[i]
	}
	return out
}

// subBytes — S: замена каждого байта через таблицу π
func subBytes(a [16]byte) [16]byte {
	var out [16]byte
	for i := range a {
		out[i] = pi[a[i]]
	}
	return out
}

// subBytesInv - обратная замена
func subBytesInv(a [16]byte) [16]byte {
	var out [16]byte
	for i := range a {
		out[i] = piInv[a[i]]
	}
	return out
}

// rStep — R: один шаг сдвигового регистра
func rStep(a [16]byte) [16]byte {
	var feedback byte
	for i := 0; i < 16; i++ {
		feedback ^= gfMul(a[i], lCoeff[i])
	}
	var out [16]byte
	out[0] = feedback
	copy(out[1:], a[:15])
	return out
}

// rStepInv — R^{-1} (big-endian хранение).
// out[0..14] = a[1..15],  out[15] = a[0] XOR ∑_{i=0}^{14} lCoeff[i]*a[i+1]
func rStepInv(a [16]byte) [16]byte {
	var out [16]byte
	copy(out[:15], a[1:])
	out[15] = a[0]
	for i := 0; i < 15; i++ {
		out[15] ^= gfMul(a[i+1], lCoeff[i])
	}
	return out
}

// lTrans — L: 16 применений R
func lTrans(a [16]byte) [16]byte {
	for i := 0; i < 16; i++ {
		a = rStep(a)
	}
	return a
}

// lTransInv — L^{-1}
func lTransInv(a [16]byte) [16]byte {
	for i := 0; i < 16; i++ {
		a = rStepInv(a)
	}
	return a
}

// iterC возвращает константу C_i = L(vec(i)), i = 1…32.
func iterC(i int) [16]byte {
	var v [16]byte
	v[15] = byte(i)
	return lTrans(v)
}

// ExpandKey разворачивает 256-битный ключ в 10 раундовых ключей (по 128 бит каждый).
func ExpandKey(key [32]byte) [10][16]byte {
	var rk [10][16]byte

	var k0, k1 [16]byte
	copy(k0[:], key[:16])
	copy(k1[:], key[16:])

	rk[0] = k0
	rk[1] = k1

	for i := 0; i < 4; i++ {
		// Каждая пара из 8 итераций (F-функция Фейстеля)
		c1 := iterC(8*i + 1)
		c2 := iterC(8*i + 2)
		c3 := iterC(8*i + 3)
		c4 := iterC(8*i + 4)
		c5 := iterC(8*i + 5)
		c6 := iterC(8*i + 6)
		c7 := iterC(8*i + 7)
		c8 := iterC(8*i + 8)

		k0, k1 = feistelRound(k0, k1, c1)
		k0, k1 = feistelRound(k0, k1, c2)
		k0, k1 = feistelRound(k0, k1, c3)
		k0, k1 = feistelRound(k0, k1, c4)
		k0, k1 = feistelRound(k0, k1, c5)
		k0, k1 = feistelRound(k0, k1, c6)
		k0, k1 = feistelRound(k0, k1, c7)
		k0, k1 = feistelRound(k0, k1, c8)

		rk[2+2*i] = k0
		rk[3+2*i] = k1
	}
	return rk
}

// feistelRound — одна итерация F-функции для развёртки ключа:
//   new_k0 = L(S(X(k0, c))) XOR k1
//   new_k1 = k0
func feistelRound(k0, k1, c [16]byte) ([16]byte, [16]byte) {
	tmp := lTrans(subBytes(xorBlock(k0, c)))
	newK0 := xorBlock(tmp, k1)
	return newK0, k0
}

// EncryptBlock шифрует один 128-битный блок.
func EncryptBlock(block [16]byte, rk [10][16]byte) [16]byte {
	// Раунды 1–9: X → S → L
	a := block
	for i := 0; i < 9; i++ {
		a = xorBlock(a, rk[i])
		a = subBytes(a)
		a = lTrans(a)
	}
	// Раунд 10: только X
	a = xorBlock(a, rk[9])
	return a
}

// DecryptBlock дешифрует один 128-битный блок.
func DecryptBlock(block [16]byte, rk [10][16]byte) [16]byte {
	// Раунды в обратном порядке
	a := xorBlock(block, rk[9])
	for i := 8; i >= 0; i-- {
		a = lTransInv(a)
		a = subBytesInv(a)
		a = xorBlock(a, rk[i])
	}
	return a
}

// PadPKCS7 дополняет срез до кратного 16 байтам.
func PadPKCS7(data []byte) []byte {
	pad := 16 - len(data)%16
	out := make([]byte, len(data)+pad)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

// UnpadPKCS7 удаляет PKCS#7-дополнение.
func UnpadPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data)%16 != 0 {
		return nil, errBadPad
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > 16 {
		return nil, errBadPad
	}
	for i := len(data) - pad; i < len(data); i++ {
		if data[i] != byte(pad) {
			return nil, errBadPad
		}
	}
	return data[:len(data)-pad], nil
}

var errBadPad = &kuzError{"неверный PKCS#7-паддинг"}

type kuzError struct{ msg string }

func (e *kuzError) Error() string { return e.msg }
