// Общие утилиты CLI для программ шифрования DES.
// Содержит ввод строк, разбор ключа и IV.
package cliutil

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

var stdinScanner = bufio.NewScanner(os.Stdin)

// ReadLine читает строку из стандартного ввода.
func ReadLine(prompt string) string {
	fmt.Print(prompt)
	stdinScanner.Scan()
	return stdinScanner.Text()
}

// ParseKey разбирает строку ключа:
//   - 16 hex-символов → 8 байт
//   - до 8 символов текста (дополняется нулями до 8 байт)
func ParseKey(input string) ([8]byte, error) {
	input = strings.TrimSpace(input)
	var key [8]byte

	if len(input) == 16 {
		b, err := hex.DecodeString(input)
		if err == nil {
			copy(key[:], b)
			return key, nil
		}
	}

	keyBytes := []byte(input)
	if len(keyBytes) > 8 {
		keyBytes = keyBytes[:8]
	}
	copy(key[:], keyBytes)
	return key, nil
}

// ParseIV разбирает строку IV: ровно 16 hex-символов (8 байт).
// Если строка пуста — генерирует случайный IV.
func ParseIV(input string) ([8]byte, error) {
	input = strings.TrimSpace(input)
	var iv [8]byte

	if input == "" {
		if _, err := rand.Read(iv[:]); err != nil {
			return iv, fmt.Errorf("не удалось сгенерировать случайный IV: %w", err)
		}
		return iv, nil
	}

	if len(input) != 16 {
		return iv, fmt.Errorf("IV должен быть задан в виде 16 hex-символов (8 байт)")
	}
	b, err := hex.DecodeString(input)
	if err != nil {
		return iv, fmt.Errorf("неверный hex-формат IV: %w", err)
	}
	copy(iv[:], b)
	return iv, nil
}
