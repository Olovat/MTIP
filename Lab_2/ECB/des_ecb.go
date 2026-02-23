package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"cliutil"
	"descore"
)

//  DES-ECB: шифрование и дешифрование произвольного сообщения

// encryptECB шифрует текст в режиме ECB.
func encryptECB(plaintext []byte, key [8]byte) []byte {
	subkeys := descore.GenerateSubkeys(key)
	padded := descore.PadPKCS7(plaintext)
	ciphertext := make([]byte, len(padded))
	for i := 0; i < len(padded); i += 8 {
		var block [8]byte
		copy(block[:], padded[i:i+8])
		result := descore.DesBlock(block, subkeys)
		copy(ciphertext[i:], result[:])
	}
	return ciphertext
}

// decryptECB дешифрует текст в режиме ECB.
func decryptECB(ciphertext []byte, key [8]byte) ([]byte, error) {
	if len(ciphertext)%8 != 0 {
		return nil, fmt.Errorf("длина шифртекста должна быть кратна 8 байтам")
	}
	revSubkeys := descore.ReverseSubkeys(descore.GenerateSubkeys(key))
	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += 8 {
		var block [8]byte
		copy(block[:], ciphertext[i:i+8])
		result := descore.DesBlock(block, revSubkeys)
		copy(plaintext[i:], result[:])
	}
	return descore.UnpadPKCS7(plaintext)
}

func main() {
	fmt.Println()
	fmt.Println("Шифр DES — режим ЭКК (Электронная кодовая книга, ECB)")
	fmt.Println("  Ключ  : до 8 символов  ИЛИ  16 hex-символов (8 байт)")
	fmt.Println("  Дополнение: PKCS#7")
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
			key, err := cliutil.ParseKey(keyStr)
			if err != nil {
				fmt.Println("Ошибка ключа:", err)
				continue
			}
			ciphertext := encryptECB([]byte(text), key)
			fmt.Println("\nЗашифрованный текст (hex):", hex.EncodeToString(ciphertext))
			fmt.Println()

		case "2":
			hexText := strings.TrimSpace(cliutil.ReadLine("Введите hex-шифртекст: "))
			ciphertext, err := hex.DecodeString(hexText)
			if err != nil {
				fmt.Println("Ошибка: неверный hex-формат:", err)
				continue
			}
			keyStr := cliutil.ReadLine("Введите ключ:          ")
			key, err := cliutil.ParseKey(keyStr)
			if err != nil {
				fmt.Println("Ошибка ключа:", err)
				continue
			}
			plaintext, err := decryptECB(ciphertext, key)
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
