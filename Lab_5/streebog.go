// Источник констант - RFC 6986
// необходим при проверке электронной цифровой подписи
package main

import (
	"encoding/binary"
)

// pi — таблица нелинейной биекции
var pi = [256]byte{
	252, 238, 221, 17, 207, 110, 49, 22, 251, 196, 250, 218, 35, 197, 4, 77,
	233, 119, 240, 219, 147, 46, 153, 186, 23, 54, 241, 187, 20, 205, 95, 193,
	249, 24, 101, 90, 226, 92, 239, 33, 129, 28, 60, 66, 139, 1, 142, 79,
	5, 132, 2, 174, 227, 106, 143, 160, 6, 11, 237, 152, 127, 212, 211, 31,
	235, 52, 44, 81, 234, 200, 72, 171, 242, 42, 104, 162, 253, 58, 206, 204,
	181, 112, 14, 86, 8, 12, 118, 18, 191, 114, 19, 71, 156, 183, 93, 135,
	21, 161, 150, 41, 16, 123, 154, 199, 243, 145, 120, 111, 157, 158, 178, 177,
	50, 117, 25, 61, 255, 53, 138, 126, 109, 84, 198, 128, 195, 189, 13, 87,
	223, 245, 36, 169, 62, 168, 67, 201, 215, 121, 214, 246, 124, 34, 185, 3,
	224, 15, 236, 222, 122, 148, 176, 188, 220, 232, 40, 80, 78, 51, 10, 74,
	167, 151, 96, 115, 30, 0, 98, 68, 26, 184, 56, 130, 100, 159, 38, 65,
	173, 69, 70, 146, 39, 94, 85, 47, 140, 163, 165, 125, 105, 213, 149, 59,
	7, 88, 179, 64, 134, 172, 29, 247, 48, 55, 107, 228, 136, 217, 231, 137,
	225, 27, 131, 73, 76, 63, 248, 254, 141, 83, 170, 144, 202, 216, 133, 97,
	32, 113, 103, 164, 45, 43, 9, 91, 203, 155, 37, 208, 190, 229, 108, 82,
	89, 166, 116, 210, 230, 244, 180, 192, 209, 102, 175, 194, 57, 75, 99, 182,
}

//  Матрица A линейного преобразования L

var aMatrix = [64]uint64{
	0x8e20faa72ba0b470, 0x47107ddd9b505a38, 0xad08b0e0c3282d1c, 0xd8045870ef14980e,
	0x6c022c38f90a4c07, 0x3601161cf205268d, 0x1b8e0b0e798c13c8, 0x83478b07b2468764,
	0xa011d380818e8f40, 0x5086e740ce47c920, 0x2843fd2067adea10, 0x14aff010bdd87508,
	0x0ad97808d06cb404, 0x05e23c0468365a02, 0x8c711e02341b2d01, 0x46b60f011a83988e,
	0x90dab52a387ae76f, 0x486dd4151c3dfdb9, 0x24b86a840e90f0d2, 0x125c354207487869,
	0x092e94218d243cba, 0x8a174a9ec8121e5d, 0x4585254f64090fa0, 0xaccc9ca9328a8950,
	0x9d4df05d5f661451, 0xc0a878a0a1330aa6, 0x60543c50de970553, 0x302a1e286fc58ca7,
	0x18150f14b9ec46dd, 0x0c84890ad27623e0, 0x0642ca05693b9f70, 0x0321658cba93c138,
	0x86275df09ce8aaa8, 0x439da0784e745554, 0xafc0503c273aa42a, 0xd960281e9d1d5215,
	0xe230140fc0802984, 0x71180a8960409a42, 0xb60c05ca30204d21, 0x5b068c651810a89e,
	0x456c34887a3805b9, 0xac361a443d1c8cd2, 0x561b0d22900e4669, 0x2b838811480723ba,
	0x9bcf4486248d9f5d, 0xc3e9224312c8c1a0, 0xeffa11af0964ee50, 0xf97d86d98a327728,
	0xe4fa2054a80b329c, 0x727d102a548b194e, 0x39b008152acb8227, 0x9258048415eb419d,
	0x492c024284fbaec0, 0xaa16012142f35760, 0x550b8e9e21f7a530, 0xa48b474f9ef5dc18,
	0x70a6a56e2440598e, 0x3853dc371220a247, 0x1ca76e95091051ad, 0x0edd37c48a08a6d8,
	0x07e095624504536c, 0x8d70c431ac02a736, 0xc83862965601dd1b, 0x641c314b2b8ee083,
}

// iterC — итерационные константы C[1]..C[12].

var iterC = [12][8]uint64{
	{ // C[1]
		0xb1085bda1ecadae9, 0xebcb2f81c0657c1f,
		0x2f6a76432e45d016, 0x714eb88d7585c4fc,
		0x4b7ce09192676901, 0xa2422a08a460d315,
		0x05767436cc744d23, 0xdd806559f2a64507,
	},
	{ // C[2]
		0x6fa3b58aa99d2f1a, 0x4fe39d460f70b5d7,
		0xf3feea720a232b98, 0x61d55e0f16b50131,
		0x9ab5176b12d69958, 0x5cb561c2db0aa7ca,
		0x55dda21bd7cbcd56, 0xe679047021b19bb7,
	},
	{ // C[3]
		0xf574dcac2bce2fc7, 0x0a39fc286a3d8435,
		0x06f15e5f529c1f8b, 0xf2ea7514b1297b7b,
		0xd3e20fe490359eb1, 0xc1c93a376062db09,
		0xc2b6f443867adb31, 0x991e96f50aba0ab2,
	},
	{ // C[4]
		0xef1fdfb3e81566d2, 0xf948e1a05d71e4dd,
		0x488e857e335c3c7d, 0x9d721cad685e353f,
		0xa9d72c82ed03d675, 0xd8b71333935203be,
		0x3453eaa193e837f1, 0x220cbebc84e3d12e,
	},
	{ // C[5]
		0x4bea6bacad474799, 0x9a3f410c6ca92363,
		0x7f151c1f1686104a, 0x359e35d7800fffbd,
		0xbfcd1747253af5a3, 0xdfff00b723271a16,
		0x7a56a27ea9ea63f5, 0x601758fd7c6cfe57,
	},
	{ // C[6]
		0xae4faeae1d3ad3d9, 0x6fa4c33b7a3039c0,
		0x2d66c4f95142a46c, 0x187f9ab49af08ec6,
		0xcffaa6b71c9ab7b4, 0x0af21f66c2bec6b6,
		0xbf71c57236904f35, 0xfa68407a46647d6e,
	},
	{ // C[7]
		0xf4c70e16eeaac5ec, 0x51ac86febf240954,
		0x399ec6c7e6bf87c9, 0xd3473e33197a93c9,
		0x0992abc52d822c37, 0x06476983284a0504,
		0x3517454ca23c4af3, 0x8886564d3a14d493,
	},
	{ // C[8]
		0x9b1f5b424d93c9a7, 0x03e7aa020c6e4141,
		0x4eb7f8719c36de1e, 0x89b4443b4ddbc49a,
		0xf4892bcb929b0690, 0x69d18d2bd1a5c42f,
		0x36acc2355951a8d9, 0xa47f0dd4bf02e71e,
	},
	{ // C[9]
		0x378f5a541631229b, 0x944c9ad8ec165fde,
		0x3a7d3a1b25894224, 0x3cd955b7e00d0984,
		0x800a440bdbb2ceb1, 0x7b2b8a9aa6079c54,
		0x0e38dc92cb1f2a60, 0x7261445183235adb,
	},
	{ // C[10]
		0xabbedea680056f52, 0x382ae548b2e4f3f3,
		0x8941e71cff8a78db, 0x1fffe18a1b336103,
		0x9fe76702af69334b, 0x7a1e6c303b7652f4,
		0x3698fad1153bb6c3, 0x74b4c7fb98459ced,
	},
	{ // C[11]
		0x7bcd9ed0efc889fb, 0x3002c6cd635afe94,
		0xd8fa6bbbebab0761, 0x2001802114846679,
		0x8a1d71efea48b9ca, 0xefbacd1d7d476e98,
		0xdea2594ac06fd85d, 0x6bcaa4cd81f32d1b,
	},
	{ // C[12]
		0x378ee767f11631ba, 0xd21380b00449b17a,
		0xcda43c32bcdf1d77, 0xf82012d430219f9b,
		0x5d80ef9d1891cc86, 0xe71da4aa88e12852,
		0xfaf417d5d9b21b99, 0x48bc924af11bd720,
	},
}

// стырил
func sbWords(s [8]uint64) [8]uint64 {
	var out [8]uint64
	for i := 0; i < 8; i++ {
		v := s[i]
		out[i] = uint64(pi[v>>56])<<56 |
			uint64(pi[(v>>48)&0xFF])<<48 |
			uint64(pi[(v>>40)&0xFF])<<40 |
			uint64(pi[(v>>32)&0xFF])<<32 |
			uint64(pi[(v>>24)&0xFF])<<24 |
			uint64(pi[(v>>16)&0xFF])<<16 |
			uint64(pi[(v>>8)&0xFF])<<8 |
			uint64(pi[v&0xFF])
	}
	return out
}

// pbWords применяет перестановку Tau (транспозиция матрицы байт 8×8).
func pbWords(s [8]uint64) [8]uint64 {
	// Разворачиваем в байтовую матрицу 8×8
	var mat [8][8]byte
	for row := 0; row < 8; row++ {
		v := s[row]
		for col := 0; col < 8; col++ {
			mat[row][col] = byte(v >> uint(56-col*8))
		}
	}
	// Транспозиция: out_row[row][col] = mat[col][row]
	var out [8]uint64
	for row := 0; row < 8; row++ {
		var v uint64
		for col := 0; col < 8; col++ {
			v = (v << 8) | uint64(mat[col][row])
		}
		out[row] = v
	}
	return out
}

// lbWord вычисляет линейное преобразование l(v) для одного 64-битного слова.
func lbWord(v uint64) uint64 {
	var r uint64
	for i := 0; i < 64; i++ {
		if (v>>uint(i))&1 == 1 {
			r ^= aMatrix[63-i]
		}
	}
	return r
}

// lbWords применяет линейное преобразование L к каждому слову.
func lbWords(s [8]uint64) [8]uint64 {
	var out [8]uint64
	for i := 0; i < 8; i++ {
		out[i] = lbWord(s[i])
	}
	return out
}

// xorWords выполняет побитовое XOR двух состояний.
func xorWords(a, b [8]uint64) [8]uint64 {
	var out [8]uint64
	for i := 0; i < 8; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}

// lps выполняет комбинированное преобразование LPS = L(P(S(state))).
func lps(s [8]uint64) [8]uint64 {
	return lbWords(pbWords(sbWords(s)))
}

//  Шифр E и функция сжатия g_N

func encryptE(K, m [8]uint64) [8]uint64 {
	var keys [13][8]uint64
	keys[0] = K
	for i := 1; i < 13; i++ {
		keys[i] = lps(xorWords(keys[i-1], iterC[i-1]))
	}
	y := m
	for i := 0; i < 12; i++ {
		y = lps(xorWords(y, keys[i]))
	}
	y = xorWords(y, keys[12])
	return y
}

// gN — функция сжатия:
func gN(N, h, m [8]uint64) [8]uint64 {
	K := lps(xorWords(h, N))
	enc := encryptE(K, m)
	return xorWords(xorWords(enc, h), m)
}

// add512 складывает два 512-битных числа, хранящихся как [8]uint64 (big-endian).
func add512(a, b [8]uint64) [8]uint64 {
	var result [8]uint64
	var carry uint64
	for i := 7; i >= 0; i-- {
		s := a[i] + b[i]
		c1 := uint64(0)
		if s < a[i] {
			c1 = 1
		}
		s2 := s + carry
		c2 := uint64(0)
		if s2 < s {
			c2 = 1
		}
		result[i] = s2
		carry = c1 | c2
	}
	return result
}

// add512U прибавляет uint64 к 512-битному числу (big-endian [8]uint64).
func add512U(a [8]uint64, b uint64) [8]uint64 {
	result := a
	for i := 7; i >= 0; i-- {
		s := result[i] + b
		if s >= b {
			result[i] = s
			break
		}
		result[i] = s
		b = 1 // перенос
	}
	return result
}

//  Вспомогательные: преобразование блока байт ↔ [8]uint64

func blockToWords(b []byte) [8]uint64 {
	var w [8]uint64
	for i := 0; i < 8; i++ {
		w[i] = binary.BigEndian.Uint64(b[i*8:])
	}
	return w
}

func wordsToBytes(w [8]uint64) []byte {
	out := make([]byte, 64)
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint64(out[i*8:], w[i])
	}
	return out
}

//  Основная функция хэширования Стрибог-512

// streebog512 вычисляет Стрибог-512 (ГОСТ Р 34.11-2012) от сообщения data.

func streebog512(data []byte) []byte {
	// Шаг 1: инициализация
	var h [8]uint64   // IV = 0^512
	var N [8]uint64   // счётчик бит
	var sig [8]uint64 // накопленная сумма

	// Рабочая копия данных
	msg := make([]byte, len(data))
	copy(msg, data)

	// Шаг 2: обработка полных 512-битных (64-байтных) блоков
	for len(msg) >= 64 {
		m := blockToWords(msg[:64])
		h = gN(N, h, m)
		N = add512U(N, 512)
		sig = add512(sig, m)
		msg = msg[64:]
	}

	// Шаг 3: добивка последнего неполного блока
	n := len(msg) // 0..63 байт
	padded := make([]byte, 64)
	copy(padded[64-n:], msg)
	if n < 64 {
		padded[64-n-1] = 0x01 // бит-разделитель
	}
	m := blockToWords(padded)
	h = gN(N, h, m)
	N = add512U(N, uint64(n*8)) // прибавляем число бит остатка
	sig = add512(sig, m)

	// Шаг 4: финализация
	var zero [8]uint64
	h = gN(zero, h, N)
	h = gN(zero, h, sig)

	return wordsToBytes(h)
}
