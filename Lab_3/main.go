package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

var stdinScanner = bufio.NewScanner(os.Stdin)

func readLine(prompt string) string {
	fmt.Print(prompt)
	stdinScanner.Scan()
	return stdinScanner.Text()
}

// parseKey разбирает строку ключа:
//   - 64 hex-символа -> 32 байта
//   - до 32 символов текста (дополняется нулями до 32 байт)
func parseKey(input string) ([32]byte, error) {
	input = strings.TrimSpace(input)
	var key [32]byte

	if len(input) == 64 {
		b, err := hex.DecodeString(input)
		if err == nil {
			copy(key[:], b)
			return key, nil
		}
	}

	keyBytes := []byte(input)
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}
	copy(key[:], keyBytes)
	return key, nil
}

func main() {
	fmt.Println()
	fmt.Println("Шифр «Кузнечик» (Grasshopper) — ГОСТ Р 34.12-2015")
	fmt.Println("  Ключ       : до 32 символов  ИЛИ  64 hex-символа (32 байта)")
	fmt.Println("  Блок       : 128 бит (16 байт)")
	fmt.Println("  Дополнение : PKCS#7")
	fmt.Println()

	for {
		fmt.Println("Выберите действие:")
		fmt.Println("  1 — Зашифровать")
		fmt.Println("  2 — Расшифровать")
		fmt.Println("  3 — Тест-вектор ГОСТ Р 34.12-2015")
		fmt.Println("  0 — Выход")
		choice := strings.TrimSpace(readLine(": "))

		switch choice {
		case "1":
			text := readLine("Введите текст:  ")
			keyStr := readLine("Введите ключ:   ")
			key, err := parseKey(keyStr)
			if err != nil {
				fmt.Println("Ошибка ключа:", err)
				continue
			}
			rk := ExpandKey(key)
			padded := PadPKCS7([]byte(text))
			var out []byte
			for i := 0; i < len(padded); i += 16 {
				var blk [16]byte
				copy(blk[:], padded[i:i+16])
				enc := EncryptBlock(blk, rk)
				out = append(out, enc[:]...)
			}
			fmt.Println("\nЗашифрованный текст (hex):", hex.EncodeToString(out))
			fmt.Println()

		case "2":
			hexText := strings.TrimSpace(readLine("Введите hex-шифртекст: "))
			ciphertext, err := hex.DecodeString(hexText)
			if err != nil {
				fmt.Println("Ошибка: неверный hex-формат:", err)
				continue
			}
			if len(ciphertext)%16 != 0 {
				fmt.Println("Ошибка: длина шифртекста должна быть кратна 16 байтам")
				continue
			}
			keyStr := readLine("Введите ключ:          ")
			key, err := parseKey(keyStr)
			if err != nil {
				fmt.Println("Ошибка ключа:", err)
				continue
			}
			rk := ExpandKey(key)
			var dec []byte
			for i := 0; i < len(ciphertext); i += 16 {
				var blk [16]byte
				copy(blk[:], ciphertext[i:i+16])
				d := DecryptBlock(blk, rk)
				dec = append(dec, d[:]...)
			}
			plain, err := UnpadPKCS7(dec)
			if err != nil {
				fmt.Println("Ошибка дешифрования:", err)
				continue
			}
			fmt.Println("\nРасшифрованный текст:", string(plain))
			fmt.Println()

		case "3":
			runSelfTest()
			fmt.Println()

		case "0":
			fmt.Println("Выход.")
			return

		default:
			fmt.Println("Неверный выбор, попробуйте снова.")
		}
	}
}

// runSelfTest проверяет тест-вектор из ГОСТ Р 34.12-2015
func runSelfTest() {
	keyBytes, _ := hex.DecodeString("8899aabbccddeeff0011223344556677fedcba98765432100123456789abcdef")
	ptBytes, _ := hex.DecodeString("1122334455667700ffeeddccbbaa9988")
	wantCT := "7f679d90bebc24305a468d42b9d4edcd"

	var key [32]byte
	copy(key[:], keyBytes)
	var pt [16]byte
	copy(pt[:], ptBytes)

	rk := ExpandKey(key)
	ct := EncryptBlock(pt, rk)
	gotCT := hex.EncodeToString(ct[:])

	fmt.Printf("Ключ          : %s\n", hex.EncodeToString(keyBytes))
	fmt.Printf("Открытый текст: %s\n", hex.EncodeToString(ptBytes))
	fmt.Printf("Ожидается     : %s\n", wantCT)
	fmt.Printf("Получено      : %s\n", gotCT)
	if gotCT == wantCT {
		fmt.Println("Результат     : УСПЕХ")
	} else {
		fmt.Println("Результат     : ОШИБКА")
	}

	dec := DecryptBlock(ct, rk)
	gotPT := hex.EncodeToString(dec[:])
	if gotPT == hex.EncodeToString(ptBytes) {
		fmt.Println("Дешифрование  :", gotPT, "OK")
	} else {
		fmt.Println("Дешифрование  : ОШИБКА, получено", gotPT)
	}
}
