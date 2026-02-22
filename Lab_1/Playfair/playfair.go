package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

const (
	latinAlphabet    = "abcdefghiklmnopqrstuvwxyz"
	cyrillicAlphabet = "абвгдеёжзийклмнопрстуфхцчшщыьэюя"
)

type playfairTable struct {
	grid [][]rune
	rows int
	cols int
	pos  map[rune][2]int // буква - {строка, столбец}
}

func buildTable(key string, rows, cols int, alphabet string, norm func(rune) rune) *playfairTable {
	seen := make(map[rune]bool)
	var order []rune

	// Сначала добавляем буквы из ключа (в порядке появления, без повторов)
	for _, r := range []rune(key) {
		nr := norm(r)
		if strings.ContainsRune(alphabet, nr) && !seen[nr] {
			seen[nr] = true
			order = append(order, nr)
		}
	}

	// Затем — оставшиеся буквы алфавита
	for _, r := range []rune(alphabet) {
		if !seen[r] {
			seen[r] = true
			order = append(order, r)
		}
	}

	grid := make([][]rune, rows)
	pos := make(map[rune][2]int)
	idx := 0
	for i := range grid {
		grid[i] = make([]rune, cols)
		for j := range grid[i] {
			if idx < len(order) {
				grid[i][j] = order[idx]
				pos[order[idx]] = [2]int{i, j}
				idx++
			}
		}
	}

	return &playfairTable{grid: grid, rows: rows, cols: cols, pos: pos}
}

// encBigram шифрует пару букв по правилам Плейфейра.
func (t *playfairTable) encBigram(a, b rune) (rune, rune) {
	pa, pb := t.pos[a], t.pos[b]
	switch {
	case pa[0] == pb[0]: // одна строка — сдвиг вправо
		return t.grid[pa[0]][(pa[1]+1)%t.cols],
			t.grid[pb[0]][(pb[1]+1)%t.cols]
	case pa[1] == pb[1]: // один столбец — сдвиг вниз
		return t.grid[(pa[0]+1)%t.rows][pa[1]],
			t.grid[(pb[0]+1)%t.rows][pb[1]]
	default: // прямоугольник — меняем столбцы
		return t.grid[pa[0]][pb[1]], t.grid[pb[0]][pa[1]]
	}
}

// decBigram дешифрует пару букв по правилам Плейфейра.
func (t *playfairTable) decBigram(a, b rune) (rune, rune) {
	pa, pb := t.pos[a], t.pos[b]
	switch {
	case pa[0] == pb[0]:
		return t.grid[pa[0]][(pa[1]-1+t.cols)%t.cols],
			t.grid[pb[0]][(pb[1]-1+t.cols)%t.cols]
	case pa[1] == pb[1]:
		return t.grid[(pa[0]-1+t.rows)%t.rows][pa[1]],
			t.grid[(pb[0]-1+t.rows)%t.rows][pb[1]]
	default: // прямоугольник — меняем столбцы
		return t.grid[pa[0]][pb[1]], t.grid[pb[0]][pa[1]]
	}
}

// вспомогательные, для определения алфавита и нормализации букв
func isLatinLetter(r rune) bool {
	lr := unicode.ToLower(r)
	return lr >= 'a' && lr <= 'z'
}

func isCyrillicLetter(r rune) bool {
	lr := unicode.ToLower(r)
	return strings.ContainsRune(cyrillicAlphabet, lr) || lr == 'ъ'
}

func normLatin(r rune) rune {
	r = unicode.ToLower(r)
	if r == 'j' {
		return 'i'
	}
	return r
}

func normCyrillic(r rune) rune {
	r = unicode.ToLower(r)
	if r == 'ъ' {
		return 'ь'
	}
	return r
}

// prepareBigrams подготавливает срез букв к шифрованию:
// filter - дополняет до чётной длины заполнителем.
func prepareBigrams(letters []rune, filler rune, norm func(rune) rune) []rune {
	normalized := make([]rune, 0, len(letters))
	for _, r := range letters {
		normalized = append(normalized, norm(r))
	}

	var result []rune
	i := 0
	for i < len(normalized) {
		a := normalized[i]
		if i+1 >= len(normalized) {
			b := filler
			if a == b {
				b = 'q'
				if !isLatinLetter(b) {
					b = 'й'
				}
			}
			result = append(result, a, b)
			i++
		} else {
			b := normalized[i+1]
			if a == b {
				f := filler
				if a == f {
					f = 'q'
					if !isLatinLetter(f) {
						f = 'й'
					}
				}
				result = append(result, a, f)
				i++ // вторую букву оставляем для следующей биграммы
			} else {
				result = append(result, a, b)
				i += 2
			}
		}
	}
	return result
}

// обработка входного потока в зависимости от флага encrypt.
func processStream(letters []rune, t *playfairTable, filler rune, norm func(rune) rune, encrypt bool) ([]rune, error) {
	if len(letters) == 0 {
		return nil, nil
	}

	if encrypt {
		prepared := prepareBigrams(letters, filler, norm)
		result := make([]rune, 0, len(prepared))
		for i := 0; i < len(prepared); i += 2 {
			a, b := t.encBigram(prepared[i], prepared[i+1])
			result = append(result, a, b)
		}
		return result, nil
	}

	normalized := make([]rune, 0, len(letters))
	for _, r := range letters {
		normalized = append(normalized, norm(r))
	}
	if len(normalized)%2 != 0 {
		return nil, fmt.Errorf("длина зашифрованного текста должна быть чётной (шифр Плейфейра работает с парами)")
	}
	result := make([]rune, 0, len(normalized))
	for i := 0; i < len(normalized); i += 2 {
		a, b := t.decBigram(normalized[i], normalized[i+1])
		result = append(result, a, b)
	}
	return result, nil
}

// Главная функция обработки текста, шифрует или дешифрует текст методом Плейфейра
// дополнительные символы-заполнители (при шифровании) добавляются в конец
func process(text, key string, encrypt bool) (string, error) {
	if key == "" {
		return "", fmt.Errorf("ключ не может быть пустым")
	}

	// Строим таблицы для обоих алфавитов
	tLat := buildTable(key, 5, 5, latinAlphabet, normLatin)
	tCyr := buildTable(key, 4, 8, cyrillicAlphabet, normCyrillic)

	// Разбираем исходный текст на позиционированные токены
	type token struct {
		r     rune
		isLat bool
		isCyr bool
		upper bool
	}
	runes := []rune(text)
	tokens := make([]token, len(runes))
	for i, r := range runes {
		tokens[i] = token{
			r:     r,
			isLat: isLatinLetter(r),
			isCyr: isCyrillicLetter(r),
			upper: unicode.IsUpper(r),
		}
	}

	// Собираем отдельные потоки букв
	var latLetters, cyrLetters []rune
	for _, tok := range tokens {
		if tok.isLat {
			latLetters = append(latLetters, tok.r)
		} else if tok.isCyr {
			cyrLetters = append(cyrLetters, tok.r)
		}
	}

	// Обрабатываем каждый поток
	latResult, err := processStream(latLetters, tLat, 'x', normLatin, encrypt)
	if err != nil {
		return "", err
	}
	cyrResult, err := processStream(cyrLetters, tCyr, 'а', normCyrillic, encrypt)
	if err != nil {
		return "", err
	}

	// буквы заменяются результатами шифрования/дешифрования (с сохранением регистра оригинала)

	var sb strings.Builder
	latIdx, cyrIdx := 0, 0

	for _, tok := range tokens {
		switch {
		case tok.isLat && latIdx < len(latResult):
			r := latResult[latIdx]
			if tok.upper {
				r = unicode.ToUpper(r)
			}
			sb.WriteRune(r)
			latIdx++
		case tok.isCyr && cyrIdx < len(cyrResult):
			r := cyrResult[cyrIdx]
			if tok.upper {
				r = unicode.ToUpper(r)
			}
			sb.WriteRune(r)
			cyrIdx++
		case !tok.isLat && !tok.isCyr:
			sb.WriteRune(tok.r)
		}
	}

	// Дописываем заполнители, если при шифровании текст стал длиннее
	for ; latIdx < len(latResult); latIdx++ {
		sb.WriteRune(latResult[latIdx])
	}
	for ; cyrIdx < len(cyrResult); cyrIdx++ {
		sb.WriteRune(cyrResult[cyrIdx])
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
	fmt.Println("Биграммный шифр Плейфейра")
	fmt.Println("  Латиница : таблица 5×5  (J = I)")
	fmt.Println("  Кириллица: таблица 4×8  (Ъ = Ь)")
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
