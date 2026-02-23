package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"cliutil"
	"descore"
)

//  DES-CBC: шифрование и дешифрование

// xorBlocks выполняет побайтовый XOR двух 8-байтных блоков.
func xorBlocks(a, b [8]byte) [8]byte {
	var result [8]byte
	for i := 0; i < 8; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

// encryptCBC шифрует открытый текст в режиме CBC.
// Возвращает IV (8 байт) + шифртекст, объединённые в одном срезе.
// Схема: C[i] = E_K(P[i] XOR C[i-1]),  C[0] = IV
func encryptCBC(plaintext []byte, key [8]byte, iv [8]byte) []byte {
	subkeys := descore.GenerateSubkeys(key)
	padded := descore.PadPKCS7(plaintext)

	out := make([]byte, 8+len(padded))
	copy(out[:8], iv[:])

	prev := iv
	for i := 0; i < len(padded); i += 8 {
		var block [8]byte
		copy(block[:], padded[i:i+8])

		// XOR открытого блока с предыдущим блоком шифртекста (или IV)
		xored := xorBlocks(block, prev)

		// Шифрование одного блока DES
		encrypted := descore.DesBlock(xored, subkeys)
		copy(out[8+i:], encrypted[:])

		prev = encrypted
	}
	return out
}

// decryptCBC дешифрует шифртекст в режиме CBC.
// Принимает IV (8 байт) + шифртекст, объединённые в одном срезе.
// Схема: P[i] = D_K(C[i]) XOR C[i-1],  C[0] = IV
func decryptCBC(data []byte, key [8]byte) ([]byte, error) {
	if len(data) < 16 || (len(data)-8)%8 != 0 {
		return nil, fmt.Errorf("неверная длина данных (ожидается IV + шифртекст, кратный 8 байтам)")
	}

	// Извлекаем IV и шифртекст
	var iv [8]byte
	copy(iv[:], data[:8])
	ciphertext := data[8:]

	revSubkeys := descore.ReverseSubkeys(descore.GenerateSubkeys(key))

	plaintext := make([]byte, len(ciphertext))
	prev := iv
	for i := 0; i < len(ciphertext); i += 8 {
		var block [8]byte
		copy(block[:], ciphertext[i:i+8])

		// Дешифрование одного блока DES
		decrypted := descore.DesBlock(block, revSubkeys)

		// XOR расшифрованного блока с предыдущим блоком шифртекста (или IV)
		xored := xorBlocks(decrypted, prev)
		copy(plaintext[i:], xored[:])

		prev = block
	}
	return descore.UnpadPKCS7(plaintext)
}

func main() {
	fmt.Println()
	fmt.Println("  Ключ : до 8 символов  ИЛИ  16 hex-символов (8 байт)")
	fmt.Println("  IV   : 16 hex-символов (8 байт); оставьте пустым для случайного IV")
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

			result := encryptCBC([]byte(text), key, iv)
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

			plaintext, err := decryptCBC(data, key)
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
