package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"cliutil"
	"descore"
)

// xorBytes выполняет побайтовый XOR двух срезов одинаковой длины.
func xorBytes(a, b []byte) []byte {
	result := make([]byte, len(a))
	for i := range a {
		result[i] = a[i] ^ b[i]
	}
	return result
}

// generateKeystream вырабатывает гамму (keystream) длиной n байт в режиме OFB.
func generateKeystream(n int, key [8]byte, iv [8]byte) []byte {
	subkeys := descore.GenerateSubkeys(key)
	keystream := make([]byte, 0, n)

	register := iv
	for len(keystream) < n {
		// Шифруем регистр обратной связи
		register = descore.DesBlock(register, subkeys)
		keystream = append(keystream, register[:]...)
	}
	return keystream[:n]
}

// encryptOFB шифрует открытый текст в режиме OFB.
// Возвращает IV (8 байт)  шифртекст.
func encryptOFB(plaintext []byte, key [8]byte, iv [8]byte) []byte {
	keystream := generateKeystream(len(plaintext), key, iv)
	ciphertext := xorBytes(plaintext, keystream)

	out := make([]byte, 8+len(ciphertext))
	copy(out[:8], iv[:])
	copy(out[8:], ciphertext)
	return out
}

// decryptOFB дешифрует шифртекст в режиме OFB.
// Принимает: IV (8 байт) || шифртекст.
func decryptOFB(data []byte, key [8]byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("данные слишком короткие: ожидается минимум 8 байт (IV)")
	}

	var iv [8]byte
	copy(iv[:], data[:8])
	ciphertext := data[8:]

	if len(ciphertext) == 0 {
		return []byte{}, nil
	}

	keystream := generateKeystream(len(ciphertext), key, iv)
	return xorBytes(ciphertext, keystream), nil
}

func main() {
	fmt.Println()
	fmt.Println("Шифр DES — режим ОСВ (Обратная связь по выходу, OFB)")
	fmt.Println("  Ключ : до 8 символов  ИЛИ  16 hex-символов (8 байт)")
	fmt.Println("  IV   : 16 hex-символов (8 байт); оставьте пустым для случайного IV")
	fmt.Println("  Дополнение: не требуется (потоковый режим)")
	fmt.Println("  Формат вывода: hex(IV) || hex(шифртекст)")
	fmt.Println()

	for {
		fmt.Println("Выберите действие:")
		fmt.Println("  1 — Зашифровать")
		fmt.Println("  2 — Расшифровать")
		fmt.Println("  0 — Выход")
		choice := strings.TrimSpace(cliutil.ReadLine(": "))

		switch choice {
		case "1":
			text := cliutil.ReadLine("Введите текст:  ")
			keyStr := cliutil.ReadLine("Введите ключ:   ")
			ivStr := cliutil.ReadLine("Введите IV (hex, пусто = случайный): ")

			key, err := cliutil.ParseKey(keyStr)
			if err != nil {
				fmt.Println("Ошибка ключа:", err)
				continue
			}
			iv, err := cliutil.ParseIV(ivStr)
			if err != nil {
				fmt.Println("Ошибка IV:", err)
				continue
			}

			result := encryptOFB([]byte(text), key, iv)
			fmt.Println("\nIV (hex):                  ", hex.EncodeToString(iv[:]))
			fmt.Println("Зашифрованный текст (hex): ", hex.EncodeToString(result[8:]))
			fmt.Println("IV || шифртекст (hex):     ", hex.EncodeToString(result))
			fmt.Println()

		case "2":
			hexData := strings.TrimSpace(cliutil.ReadLine("Введите hex (IV || шифртекст): "))
			data, err := hex.DecodeString(hexData)
			if err != nil {
				fmt.Println("Ошибка: неверный hex-формат:", err)
				continue
			}
			keyStr := cliutil.ReadLine("Введите ключ:                  ")
			key, err := cliutil.ParseKey(keyStr)
			if err != nil {
				fmt.Println("Ошибка ключа:", err)
				continue
			}

			plaintext, err := decryptOFB(data, key)
			if err != nil {
				fmt.Println("Ошибка дешифрования:", err)
				continue
			}
			fmt.Println("\nРасшифрованный текст:", string(plaintext))
			fmt.Println()

		case "0":
			fmt.Println("Выход.")
			return

		default:
			fmt.Println("Неверный выбор, попробуйте снова.")
		}
	}
}
