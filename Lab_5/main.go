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
	lastR, lastS   *big.Int
	lastMsg        string
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

//  Пункт 1: Генерация ключей

func menuGenerateKeys() {
	fmt.Println()
	fmt.Println("Генерация ключей ГОСТ Р 34.10-2018")
	fmt.Println()
	fmt.Println("  Параметры кривой:")
	fmt.Printf("    p  = %s...\n", Curve512Test.P.Text(16)[:32])
	fmt.Printf("    a  = %s\n", Curve512Test.A.Text(16))
	fmt.Printf("    q  = %s...\n", Curve512Test.Q.Text(16)[:32])
	fmt.Println()
	fmt.Println("  Генерируется случайный закрытый ключ d")
	fmt.Println("  Вычисляется открытый ключ Q")
	fmt.Println()

	kp, err := GenerateKeyPair(Curve512Test)
	if err != nil {
		fmt.Printf("  Ошибка: %v\n", err)
		return
	}
	currentKeyPair = &kp
	lastR, lastS = nil, nil
	lastMsg = ""

	printKeys()
}

func printKeys() {
	if currentKeyPair == nil {
		fmt.Println("  Ключи не сгенерированы")
		return
	}
	dHex := currentKeyPair.Private.D.Text(16)
	qxHex := currentKeyPair.Public.Q.X.Text(16)
	qyHex := currentKeyPair.Public.Q.Y.Text(16)

	fmt.Println("  Закрытый ключ d :")
	printHexWrapped(dHex, 4)
	fmt.Println()
	fmt.Println("  Открытый ключ Q ")
	fmt.Println("    x_Q =")
	printHexWrapped(qxHex, 6)
	fmt.Println("    y_Q =")
	printHexWrapped(qyHex, 6)
}

func menuShowKeys() {
	fmt.Println("\n Текущие ключи")
	printKeys()
}

func menuSign() {
	fmt.Println()
	fmt.Println("Формирование ЭЦП")
	if currentKeyPair == nil {
		fmt.Println("\n  Сначала сгенерируйте ключи ")
		return
	}

	msg := readLine("\n  Введите сообщение: ")
	if msg == "" {
		fmt.Println("  Сообщение не может быть пустым.")
		return
	}

	hBytes := streebog512([]byte(msg))
	alpha := new(big.Int).SetBytes(hBytes)
	e := new(big.Int).Mod(alpha, currentKeyPair.Private.Curve.Q)
	if e.Sign() == 0 {
		e.SetInt64(1)
	}

	fmt.Printf("H(M) = ")
	printHexWrapped(hex.EncodeToString(hBytes), 15)
	fmt.Println()
	fmt.Printf(" e = ")
	printHexWrapped(e.Text(16), 13)
	fmt.Println()

	r, s, err := Sign([]byte(msg), currentKeyPair.Private)
	if err != nil {
		fmt.Printf("  Ошибка: %v\n", err)
		return
	}

	lastR, lastS = r, s
	lastMsg = msg

	fmt.Println("Подпись сформирована!")
	fmt.Println()
	fmt.Println("  r (hex) =")
	printHexWrapped(r.Text(16), 4)
	fmt.Println()
	fmt.Println("  s (hex) =")
	printHexWrapped(s.Text(16), 4)
	fmt.Println()
	fmt.Println("  Сохранено в памяти для пункта 4.")
}

func menuVerify() {
	fmt.Println()
	fmt.Println("Проверка ЭЦП")
	if currentKeyPair == nil {
		fmt.Println("\n  Сначала сгенерируйте ключи ")
		return
	}

	fmt.Println()
	fmt.Println("  1 — Использовать подпись из текущей сессии")
	fmt.Println("  2 — Проверить подпись изменённого сообщения")
	choice := readChoice("  Выбор: ", 1, 2)

	var msg string
	var r, s *big.Int

	switch choice {
	case 1:
		if lastR == nil {
			fmt.Println("\n  В сессии нет сохранённой подписи. Сначала подпишите сообщение")
			return
		}
		msg = lastMsg
		r, s = lastR, lastS
		fmt.Printf("\n  Сообщение: %s\n", msg)

	case 2:
		if lastR == nil {
			fmt.Println("\n  Нет сохранённой подписи. Сначала подпишите сообщение.")
			return
		}
		r, s = lastR, lastS
		origMsg := lastMsg
		msg = readLine(fmt.Sprintf("\n  Исходное сообщение: %q\n  Введите изменённое: ", origMsg))
		fmt.Println("  (Результат проверки должен быть НЕВЕРНЫМ!! Для тестов пункт)")
	}

	fmt.Println()

	if r.Sign() <= 0 || r.Cmp(currentKeyPair.Public.Curve.Q) >= 0 ||
		s.Sign() <= 0 || s.Cmp(currentKeyPair.Public.Curve.Q) >= 0 {
		fmt.Println("Подпись НЕВЕРНА.")
		return
	}
	fmt.Println("OK")

	hBytes := streebog512([]byte(msg))
	alpha := new(big.Int).SetBytes(hBytes)
	e := new(big.Int).Mod(alpha, currentKeyPair.Public.Curve.Q)
	if e.Sign() == 0 {
		e.SetInt64(1)
	}

	valid := Verify([]byte(msg), r, s, currentKeyPair.Public)

	fmt.Println()
	if valid {
		fmt.Println("Подпись ВЕРНА")
	} else {
		fmt.Println("Подпись НЕВЕРНА")
	}
}

// Вспомогательный вывод
func printHexWrapped(hexStr string, indent int) {
	prefix := strings.Repeat(" ", indent)
	for len(hexStr) > 0 {
		chunk := 64
		if len(hexStr) < chunk {
			chunk = len(hexStr)
		}
		fmt.Printf("%s%s\n", prefix, hexStr[:chunk])
		hexStr = hexStr[chunk:]
	}
}

//  main

func main() {
	fmt.Println()
	fmt.Println("ГОСТ Р 34.10-2018 — Электронная цифровая подпись")

	for {
		fmt.Println()
		fmt.Println("  1  Генерация ключей")
		fmt.Println("  2  Показать текущие ключи")
		fmt.Println("  3  Сформировать подпись")
		fmt.Println("  4  Проверить подпись")
		fmt.Println("  0  Выход")

		choice := readChoice("  Выбор: ", 0, 4)
		switch choice {
		case 1:
			menuGenerateKeys()
		case 2:
			menuShowKeys()
		case 3:
			menuSign()
		case 4:
			menuVerify()
		case 0:

			return
		}
	}
}
