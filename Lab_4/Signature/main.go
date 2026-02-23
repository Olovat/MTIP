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
	lastSig        *big.Int
	lastSigMsg     string
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
	fmt.Printf("    e = %s\n", currentKeyPair.Public.E)
	fmt.Printf("    N = %s\n", currentKeyPair.Public.N)
	fmt.Println("  Закрытый ключ:")
	fmt.Printf("    d = %s\n", currentKeyPair.Private.D)
	fmt.Printf("    N = %s\n", currentKeyPair.Private.N)
}

// 1. Генерация ключей
func menuGenerateKeys() {
	fmt.Println("\nГенерация ключей")
	fmt.Println("  1 — 512 бит")
	fmt.Println("  2 — 1024 бит")
	fmt.Println("  3 — 2048 бит (Безопаснее)")
	choice := readChoice("  Выбор: ", 1, 3)
	bits := []int{512, 1024, 2048}[choice-1]

	fmt.Printf("\n  Генерация %d-битного ключа...\n", bits)
	kp, err := GenerateKeyPair(bits)
	if err != nil {
		fmt.Printf("  Ошибка: %v\n", err)
		return
	}
	currentKeyPair = &kp
	lastSig = nil
	lastSigMsg = ""
	fmt.Println()
	printKeys()
}

// 2. Показать текущие ключи
func menuShowKeys() {
	fmt.Println("\nТекущие ключи")
	printKeys()
}

// 3. Формирование подписи
func menuSign() {
	fmt.Println("\nФормирование ЭЦП")
	if currentKeyPair == nil {
		fmt.Println("  Сначала сгенерируйте ключи.")
		return
	}

	msg := readLine("  Сообщение: ")
	if msg == "" {
		fmt.Println("  Сообщение пустое.")
		return
	}

	sig := SignMessage([]byte(msg), currentKeyPair.Private)
	h := sha256Hash([]byte(msg))
	h = normalizeHash(h, currentKeyPair.Private.N)

	lastSig = sig
	lastSigMsg = msg

	fmt.Println()
	fmt.Printf("  Сообщение    : %s\n", msg)
	fmt.Printf("  SHA-256(msg) : %s\n", h.Text(16))
	fmt.Printf("  H^d mod N    : (подпись)\n")
	fmt.Printf("  Подпись (hex): %s\n", hex.EncodeToString(sig.Bytes()))
	fmt.Println()
	fmt.Println("  Подпись сохранена для пункта 4.")
}

// 4. Проверка подписи
func menuVerify() {
	fmt.Println("\nПроверка ЭЦП")
	if currentKeyPair == nil {
		fmt.Println("  Сначала сгенерируйте ключи")
		return
	}

	fmt.Println("  1 — Использовать подпись из текущей сессии")
	fmt.Println("  2 — Ввести подпись вручную")
	choice := readChoice("  Выбор: ", 1, 2)

	var msg string
	var sig *big.Int

	switch choice {
	case 1:
		if lastSig == nil {
			fmt.Println("  В текущей сессии подпись не создана.")
			return
		}
		sig = lastSig
		msg = lastSigMsg
		fmt.Printf("\n  Сохранённое сообщение: %s\n", msg)
		fmt.Printf("  Сохранённая подпись  : %s\n", hex.EncodeToString(sig.Bytes()))

	case 2:
		msg = readLine("  Сообщение: ")
		sigHex := readLine("  Подпись (hex): ")
		b, err := hex.DecodeString(strings.TrimSpace(sigHex))
		if err != nil {
			fmt.Printf("  Ошибка декодирования: %v\n", err)
			return
		}
		sig = new(big.Int).SetBytes(b)
	}

	valid, h, hRecovered := VerifySignature([]byte(msg), sig, currentKeyPair.Public)

	fmt.Println()
	fmt.Printf("  Проверяемое сообщение : %s\n", msg)
	fmt.Printf("  SHA-256(msg)     H    : %s\n", h.Text(16))
	fmt.Printf("  S^e mod N        H'   : %s\n", hRecovered.Text(16))
	fmt.Println()
	if valid {
		fmt.Println("ПОДПИСЬ ВЕРНА")
	} else {
		fmt.Println("ПОДПИСЬ НЕВЕРНА")
	}
}

// 5. Демонстрация потом убрать когда лабу показывать время
func menuDemo() {
	fmt.Println("\nДемонстрация ЭЦП")
	fmt.Println("  Генерация 512-битного ключа...")
	kp, err := GenerateKeyPair(512)
	if err != nil {
		fmt.Printf("  Ошибка: %v\n", err)
		return
	}
	currentKeyPair = &kp
	fmt.Println()
	printKeys()

	msg := "Подписанный документ №1"
	fmt.Printf("\n  Сообщение: %s\n", msg)

	sig := SignMessage([]byte(msg), kp.Private)
	fmt.Printf("  Подпись (hex): %s\n", hex.EncodeToString(sig.Bytes()))

	// Проверка оригинала
	ok, h, hr := VerifySignature([]byte(msg), sig, kp.Public)
	fmt.Printf("\n  [Оригинальное сообщение]\n")
	fmt.Printf("  H  = %s\n", h.Text(16))
	fmt.Printf("  H' = %s\n", hr.Text(16))
	fmt.Printf("  Результат: %v", ok)
	if ok {
		fmt.Println("  ✓")
	} else {
		fmt.Println("  ✗")
	}

	// Проверка изменённого сообщения
	tampered := msg + " (изменено)"
	ok2, h2, hr2 := VerifySignature([]byte(tampered), sig, kp.Public)
	fmt.Printf("\n  [Изменённое сообщение: \"%s\"]\n", tampered)
	fmt.Printf("  H  = %s\n", h2.Text(16))
	fmt.Printf("  H' = %s\n", hr2.Text(16))
	fmt.Printf("  Результат: %v", ok2)
	if ok2 {
		fmt.Println("  ✓")
	} else {
		fmt.Println("  ✗")
	}

	lastSig = sig
	lastSigMsg = msg
	fmt.Println("\n  Ключи и подпись сохранены в сессии.")
}

func main() {
	fmt.Println("Программа 2 — ЭЦП RSA")

	for {
		fmt.Println()
		fmt.Println("  1  Сгенерировать ключи")
		fmt.Println("  2  Показать текущие ключи")
		fmt.Println("  3  Подписать сообщение")
		fmt.Println("  4  Проверить подпись")
		fmt.Println("  5  Демонстрация")
		fmt.Println("  0  Выход")

		choice := readChoice("  Выбор: ", 0, 5)
		switch choice {
		case 0:
			return
		case 1:
			menuGenerateKeys()
		case 2:
			menuShowKeys()
		case 3:
			menuSign()
		case 4:
			menuVerify()
		case 5:
			menuDemo()
		}
	}
}
