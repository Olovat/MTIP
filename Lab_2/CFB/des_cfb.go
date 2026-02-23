package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"cliutil"
	"descore"
)

//  DES-CFB: шифрование и дешифрование

// xorBytes применяет побайтовый XOR среза a со срезом b длиной n байт.
func xorBytes(a, b []byte, n int) []byte {
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}

// encryptCFB шифрует открытый текст в режиме CFB-64 (полноблочный, 64-битный сдвиг).
// Схема для каждого 8-байтного блока:
//
//	O[i] = E_K( I[i] )          — шифрование сдвигового регистра
//	C[i] = P[i] XOR O[i]        — XOR с открытым текстом
//	I[i+1] = C[i]               — сдвиговый регистр <- блок шифртекста
//
// I[0] = IV.
// Дополнение не требуется: последний неполный блок обрабатывается частичным XOR.
// Возвращает IV (8 байт) || шифртекст.
func encryptCFB(plaintext []byte, key [8]byte, iv [8]byte) []byte {
	subkeys := descore.GenerateSubkeys(key)

	// Результат: IV || шифртекст (без дополнения)
	out := make([]byte, 8+len(plaintext))
	copy(out[:8], iv[:])

	shiftReg := iv // сдвиговый регистр, начальное значение = IV

	for i := 0; i < len(plaintext); i += 8 {
		// Шифруем сдвиговый регистр
		keystream := descore.DesBlock(shiftReg, subkeys)

		// Определяем длину текущего блока (последний может быть короче 8 байт)
		end := i + 8
		if end > len(plaintext) {
			end = len(plaintext)
		}
		blockLen := end - i

		// C[i] = P[i] XOR O[i] (только нужное количество байт)
		cBlock := xorBytes(plaintext[i:end], keystream[:], blockLen)
		copy(out[8+i:], cBlock)

		// Следующий сдвиговый регистр = блок шифртекста
		// Для неполного последнего блока используем частичный шифртекст,
		// остаток дополняем нулями (реально следующего блока нет)
		copy(shiftReg[:blockLen], cBlock)
		// хвост обнуляем только при необходимости (неполный блок)
		for j := blockLen; j < 8; j++ {
			shiftReg[j] = 0
		}
	}
	return out
}

// decryptCFB дешифрует шифртекст в режиме CFB-64.
// Принимает IV (8 байт) || шифртекст, объединённые в одном срезе.
// Схема:
//
//	O[i] = E_K( I[i] )          — шифрование сдвигового регистра (то же, что при шифровании!)
//	P[i] = C[i] XOR O[i]        — восстановление открытого текста
//	I[i+1] = C[i]               — сдвиговый регистр <- блок шифртекста
func decryptCFB(data []byte, key [8]byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("данные слишком короткие: ожидается минимум IV (8 байт)")
	}

	// Извлекаем IV и шифртекст
	var iv [8]byte
	copy(iv[:], data[:8])
	ciphertext := data[8:]

	if len(ciphertext) == 0 {
		return []byte{}, nil
	}

	// В режиме CFB для дешифрования используется то же E_K (шифрование DES),
	// подключи в прямом порядке.
	subkeys := descore.GenerateSubkeys(key)

	plaintext := make([]byte, len(ciphertext))
	shiftReg := iv // сдвиговый регистр

	for i := 0; i < len(ciphertext); i += 8 {
		// Шифруем сдвиговый регистр
		keystream := descore.DesBlock(shiftReg, subkeys)

		end := i + 8
		if end > len(ciphertext) {
			end = len(ciphertext)
		}
		blockLen := end - i

		// Следующий сдвиговый регистр = блок ШИФРТЕКСТА (до XOR)
		var nextReg [8]byte
		copy(nextReg[:blockLen], ciphertext[i:end])

		// P[i] = C[i] XOR O[i]
		pBlock := xorBytes(ciphertext[i:end], keystream[:], blockLen)
		copy(plaintext[i:], pBlock)

		shiftReg = nextReg
	}
	return plaintext, nil
}

func main() {
	fmt.Println()
	fmt.Println("Шифр DES — режим ОСШ (Обратная связь по шифру, CFB-64)")
	fmt.Println("  Ключ : до 8 символов  ИЛИ  16 hex-символов (8 байт)")
	fmt.Println("  IV   : 16 hex-символов (8 байт); оставьте пустым для случайного IV")
	fmt.Println("  Дополнение: не требуется (поточный режим)")
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

			result := encryptCFB([]byte(text), key, iv)
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

			plaintext, err := decryptCFB(data, key)
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
