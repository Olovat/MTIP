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

// alphabetFor возвращает срез рун алфавита для данного символа (или nil).
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

// shiftRune сдвигает руну на сдвиг, заданный символом ключа.
func shiftRune(r rune, keyRune rune, encrypt bool) rune {
	alpha := alphabetFor(r)
	if alpha == nil {
		return r // не буква — оставляем как есть
	}

	alphaLen := len(alpha)
	lr := unicode.ToLower(r)

	pos := -1
	for i, c := range alpha {
		if c == lr {
			pos = i
			break
		}
	}
	keyAlpha := alphabetFor(keyRune)
	if keyAlpha == nil {
		return r
	}
	lk := unicode.ToLower(keyRune)
	keyPos := -1
	for i, c := range keyAlpha {
		if c == lk {
			keyPos = i
			break
		}
	}
	if pos < 0 || keyPos < 0 {
		return r
	}
	// применяем шифр Виженера: сложение/вычитание позиций по модулю N
	var newPos int
	if encrypt {
		newPos = (pos + keyPos) % alphaLen
	} else {
		newPos = (pos - keyPos + alphaLen) % alphaLen
	}

	result := alpha[newPos]
	if unicode.IsUpper(r) {
		result = unicode.ToUpper(result)
	}
	return result
}

// process шифрует или дешифрует текст методом Виженера.
func process(text, key string, encrypt bool) (string, error) {
	if key == "" {
		return "", fmt.Errorf("ключ не может быть пустым")
	}

	keyRunes := []rune(key)
	var filteredKey []rune
	for _, r := range keyRunes {
		if alphabetFor(r) != nil {
			filteredKey = append(filteredKey, r)
		}
	}
	if len(filteredKey) == 0 {
		return "", fmt.Errorf("ключ должен содержать хотя бы одну букву")
	}

	var sb strings.Builder
	keyIndex := 0
	for _, r := range []rune(text) {
		if alphabetFor(r) != nil {
			keyRune := filteredKey[keyIndex%len(filteredKey)]
			sb.WriteRune(shiftRune(r, keyRune, encrypt))
			keyIndex++
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
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
		fmt.Println("Выберите действие:")
		fmt.Println("  1 — Зашифровать")
		fmt.Println("  2 — Расшифровать")
		fmt.Println("  0 — Выход")
		choice := strings.TrimSpace(readLine(": "))

		switch choice {
		case "1", "2":
			encrypt := choice == "1"
			action := "Зашифровать"
			if !encrypt {
				action = "Расшифровать"
			}
			_ = action

			text := readLine("Введите текст: ")
			key := strings.TrimSpace(readLine("Введите ключ:  "))

			result, err := process(text, key, encrypt)
			if err != nil {
				fmt.Println("Ошибка:", err)
				continue
			}

			if encrypt {
				fmt.Println("\nЗашифрованный текст:", result)
			} else {
				fmt.Println("\nРасшифрованный текст:", result)
			}
			fmt.Println()

		case "0":
			fmt.Println("Выход.")
			return

		default:
			fmt.Println("Неверный выбор, попробуйте снова.")
		}
	}
}
