package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

const (
	latinAlphabet    = "abcdefghijklmnopqrstuvwxyz"
	cyrillicAlphabet = "абвгдеёжзийклмнопрстуфхцчшщъыьэюя"
)

// alphabetFor возвращает срез рун алфавита для данного символа (или nil)
func alphabetFor(r rune) []rune {
	lr := unicode.ToLower(r)
	if lr >= 'a' && lr <= 'z' {
		return []rune(latinAlphabet)
	}
	for _, c := range []rune(cyrillicAlphabet) {
		if c == lr {
			return []rune(cyrillicAlphabet)
		}
	}
	return nil
}

// возвращает индекс руны в срезе, или -1
func indexOf(alpha []rune, r rune) int {
	for i, c := range alpha {
		if c == r {
			return i
		}
	}
	return -1
}

// gammaCipher шифрует/дешифрует текст гамма-шифром (сложение по модулю k)
func gammaCipher(text, key string, encrypt bool) string {
	keyRunes := []rune(key)
	if len(keyRunes) == 0 {
		return text
	}
	result := []rune{}
	keyIdx := 0 // движется только по буквам

	for _, r := range []rune(text) {
		alpha := alphabetFor(r)
		if alpha == nil {
			result = append(result, r) // не буква — оставляем как есть
			continue
		}

		isUpper := unicode.IsUpper(r)
		lr := unicode.ToLower(r)
		k := len(alpha)

		// позиция символа текста
		tPos := indexOf(alpha, lr)

		// символ гаммы (циклически), пропускаем не-буквы в ключе
		var gammaRune rune
		for {
			gr := unicode.ToLower(keyRunes[keyIdx%len(keyRunes)])
			keyIdx++
			gammaAlpha := alphabetFor(gr)
			if gammaAlpha != nil {
				gammaRune = gr
				break
			}
		}

		// Ищем позицию гаммы в том же алфавите
		gammaPos := indexOf(alpha, gammaRune)
		if gammaPos == -1 {
			// гамма из другого алфавита — используем модуль от позиции в своём алфавите
			gammaAlpha := alphabetFor(gammaRune)
			if gammaAlpha != nil {
				gammaPos = indexOf(gammaAlpha, gammaRune) % k
			} else {
				gammaPos = 0
			}
		}

		var newPos int
		if encrypt {
			newPos = (tPos + gammaPos) % k
		} else {
			newPos = (tPos - gammaPos + k) % k
		}

		newRune := alpha[newPos]
		if isUpper {
			newRune = unicode.ToUpper(newRune)
		}
		result = append(result, newRune)
	}

	return string(result)
}

var stdinScanner = bufio.NewScanner(os.Stdin)

func readLine(prompt string) string {
	fmt.Print(prompt)
	stdinScanner.Scan()
	return stdinScanner.Text()
}

func main() {
	fmt.Println()

	for {
		fmt.Println("Шифрование методом гаммирования (ТШ = (ТО + ТГ) mod N)")
		fmt.Println("1 — Зашифровать")
		fmt.Println("2 — Дешифровать")
		fmt.Println("0 — Выход")
		op := strings.ToUpper(strings.TrimSpace(readLine(": ")))

		switch op {
		case "1", "2":
			text := readLine("Введите текст : ")
			key := strings.TrimSpace(readLine("Введите гамму : "))

			encrypt := op == "1"
			result := gammaCipher(text, key, encrypt)

			if encrypt {
				fmt.Printf("\nЗашифрованный текст : %s\n\n", result)
			} else {
				fmt.Printf("\nРасшифрованный текст: %s\n\n", result)
			}
		case "0":
			fmt.Println("Выход.")
			return
		default:
			fmt.Println("Неверный выбор, попробуйте снова.")
		}
	}
}
