package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
)

var (
	currentKeyPair *KeyPair
	scanner        = bufio.NewScanner(os.Stdin)
)

func readLine(prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func readChoice(prompt string, min, max int) int {
	for {
		s := readLine(prompt)
		var n int
		if _, err := fmt.Sscan(s, &n); err == nil && n >= min && n <= max {
			return n
		}
		fmt.Printf("  Введите число от %d до %d\n", min, max)
	}
}

func printKeys() {
	if currentKeyPair == nil {
		fmt.Println("  Ключи не сгенерированы.")
		return
	}
	fmt.Println("  Открытый ключ:")
	fmt.Printf("    e = %s\n", currentKeyPair.Public.E.String())
	fmt.Printf("    N = %s\n", currentKeyPair.Public.N.String())
	fmt.Println("  Закрытый ключ:")
	fmt.Printf("    d = %s\n", currentKeyPair.Private.D.String())
	fmt.Printf("    N = %s\n", currentKeyPair.Private.N.String())
}

// 1. Генерация ключей
func menuGenerateKeys() {
	fmt.Println("\nГенерация ключей")
	fmt.Println("  1 — 512 бит  (учебный)")
	fmt.Println("  2 — 1024 бит")
	fmt.Println("  3 — 2048 бит (рекомендуется)")
	choice := readChoice("  Выбор: ", 1, 3)
	bits := []int{512, 1024, 2048}[choice-1]

	fmt.Printf("\n  Генерация %d-битного ключа...\n", bits)
	kp, err := GenerateKeyPair(bits)
	if err != nil {
		fmt.Printf("  Ошибка: %v\n", err)
		return
	}
	currentKeyPair = &kp
	fmt.Println()
	printKeys()
}

// 2. Показать текущие ключи
func menuShowKeys() {
	fmt.Println("\nТекущие ключи")
	printKeys()
}

// 3. Зашифровать число: c = m^e mod N
func menuEncryptNumber() {
	fmt.Println("\nШифрование числа (c = m^e mod N)")
	if currentKeyPair == nil {
		fmt.Println("  Сначала сгенерируйте ключи")
		return
	}

	raw := readLine("  m (десятичное, 0 ≤ m < N): ")
	m := new(big.Int)
	if _, ok := m.SetString(raw, 10); !ok {
		fmt.Println("  Неверный формат числа.")
		return
	}
	if m.Sign() < 0 || m.Cmp(currentKeyPair.Public.N) >= 0 {
		fmt.Println("  Число должно быть в диапазоне [0, N).")
		return
	}

	c := EncryptInt(m, currentKeyPair.Public)
	fmt.Printf("\n  m        = %s\n", m)
	fmt.Printf("  c = m^e mod N\n  c        = %s\n", c)
}

// 4. Расшифровать число: m = c^d mod N
func menuDecryptNumber() {
	fmt.Println("\nДешифрование числа (m = c^d mod N)")
	if currentKeyPair == nil {
		fmt.Println("  Сначала сгенерируйте ключи")
		return
	}

	raw := readLine("  c (десятичное): ")
	c := new(big.Int)
	if _, ok := c.SetString(raw, 10); !ok {
		fmt.Println("  Неверный формат числа.")
		return
	}

	m := DecryptInt(c, currentKeyPair.Private)
	fmt.Printf("\n  c        = %s\n", c)
	fmt.Printf("  m = c^d mod N\n  m        = %s\n", m)
}

// 5. Зашифровать текст (блочный режим)
func menuEncryptText() {
	fmt.Println("\nШифрование текста")
	if currentKeyPair == nil {
		fmt.Println("  Сначала сгенерируйте ключи")
		return
	}

	msg := readLine("  Открытый текст: ")
	if msg == "" {
		fmt.Println("  Сообщение пустое.")
		return
	}

	blocks, err := EncryptBytes([]byte(msg), currentKeyPair.Public)
	if err != nil {
		fmt.Printf("  Ошибка шифрования: %v\n", err)
		return
	}

	parts := make([]string, len(blocks))
	for i, b := range blocks {
		parts[i] = hex.EncodeToString(b)
	}
	fmt.Printf("\n  Блоков: %d\n", len(blocks))
	fmt.Printf("  Шифртекст (hex, блоки через ':')\n  %s\n", strings.Join(parts, ":"))
}

// 6. Расшифровать текст
func menuDecryptText() {
	fmt.Println("\nДешифрование текста")
	if currentKeyPair == nil {
		fmt.Println("  Сначала сгенерируйте ключи.")
		return
	}

	raw := readLine("  Шифртекст (hex, блоки через ':'): ")
	if raw == "" {
		fmt.Println("  Ввод пуст.")
		return
	}

	parts := strings.Split(raw, ":")
	var blocks [][]byte
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		b, err := hex.DecodeString(p)
		if err != nil {
			fmt.Printf("  Ошибка hex: %v\n", err)
			return
		}
		blocks = append(blocks, b)
	}

	plain, err := DecryptBytes(blocks, currentKeyPair.Private)
	if err != nil {
		fmt.Printf("  Ошибка дешифрования: %v\n", err)
		return
	}
	fmt.Printf("\n  Открытый текст: %s\n", string(plain))
}

func main() {
	fmt.Println("Программа 1 — Шифрование RSA")

	for {
		fmt.Println()
		fmt.Println("  1  Сгенерировать ключи")
		fmt.Println("  2  Показать текущие ключи")
		fmt.Println("  3  Зашифровать число  (c = m^e mod N)")
		fmt.Println("  4  Расшифровать число (m = c^d mod N)")
		fmt.Println("  5  Зашифровать текст  (блочный режим)")
		fmt.Println("  6  Расшифровать текст")
		fmt.Println("  7  Демонстрация полного цикла")
		fmt.Println("  0  Выход")

		choice := readChoice("  Выбор: ", 0, 7)
		switch choice {
		case 0:
			fmt.Println("  До свидания!")
			return
		case 1:
			menuGenerateKeys()
		case 2:
			menuShowKeys()
		case 3:
			menuEncryptNumber()
		case 4:
			menuDecryptNumber()
		case 5:
			menuEncryptText()
		case 6:
			menuDecryptText()
		}
	}
}
