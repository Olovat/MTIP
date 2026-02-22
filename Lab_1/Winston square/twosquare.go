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

type tsTable struct {
	grid [][]rune
	rows int
	cols int
	pos  map[rune][2]int // для быстрого поиска координат буквы в таблице
}

func buildTable(key string, rows, cols int, alphabet string, norm func(rune) rune) *tsTable {
	seen := make(map[rune]bool)
	var order []rune

	// Буквы ключа — по порядку появления, без повторов
	for _, r := range []rune(key) {
		nr := norm(r)
		if strings.ContainsRune(alphabet, nr) && !seen[nr] {
			seen[nr] = true
			order = append(order, nr)
		}
	}
	// Оставшиеся буквы алфавита
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
	return &tsTable{grid: grid, rows: rows, cols: cols, pos: pos}
}

// printTable выводит таблицу на экран (для наглядности).
func (t *tsTable) printTable(title string) {
	fmt.Printf("  %s:\n", title)
	for _, row := range t.grid {
		fmt.Print("    ")
		for _, r := range row {
			fmt.Printf("%c ", r)
		}
		fmt.Println()
	}
}

// processBigram выполняет над парой букв. Пара таблиц tL и tR соответствует алфавиту, к которому принадлежат буквы a и b.
func processBigram(a, b rune, tL, tR *tsTable) (rune, rune) {
	pa := tL.pos[a]
	pb := tR.pos[b]

	if pa[0] == pb[0] {
		// Буквы в одной строке — прямоугольник вырождается, замена не производится.
		return a, b
	}
	return tL.grid[pa[0]][pb[1]], tR.grid[pb[0]][pa[1]]
}

// Нормализация и проверка принадлежности к алфавиту

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

func isLatinLetter(r rune) bool {
	lr := unicode.ToLower(r)
	return lr >= 'a' && lr <= 'z'
}

func isCyrillicLetter(r rune) bool {
	lr := unicode.ToLower(r)
	return strings.ContainsRune(cyrillicAlphabet, lr) || lr == 'ъ'
}

// Обработка потока букв
// Если количество букв нечётное — в конец добавляется заполнитель (filler), который уже должен быть нормализован.
func processLetters(letters []rune, filler rune, tL, tR *tsTable) []rune {
	if len(letters)%2 != 0 {
		letters = append(letters, filler)
	}
	result := make([]rune, 0, len(letters))
	for i := 0; i < len(letters); i += 2 {
		c1, c2 := processBigram(letters[i], letters[i+1], tL, tR)
		result = append(result, c1, c2)
	}
	return result
}

// Главная функция обработки текста

// process шифрует или дешифрует текст шифром
// keyL — ключ для левой таблицы, keyR — ключ для правой.
func process(text, keyL, keyR string) (string, error) {
	if keyL == "" || keyR == "" {
		return "", fmt.Errorf("оба ключа не могут быть пустыми")
	}

	//  Строим пары таблиц (левая / правая) для каждого алфавита
	tLatL := buildTable(keyL, 5, 5, latinAlphabet, normLatin)
	tLatR := buildTable(keyR, 5, 5, latinAlphabet, normLatin)

	tCyrL := buildTable(keyL, 4, 8, cyrillicAlphabet, normCyrillic)
	tCyrR := buildTable(keyR, 4, 8, cyrillicAlphabet, normCyrillic)

	//  Разбираем текст на токены
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

	// Собираем потоки нормализованных букв
	var latLetters, cyrLetters []rune
	for _, tok := range tokens {
		if tok.isLat {
			latLetters = append(latLetters, normLatin(tok.r))
		} else if tok.isCyr {
			cyrLetters = append(cyrLetters, normCyrillic(tok.r))
		}
	}

	// Обрабатываем каждый поток
	latResult := processLetters(latLetters, 'x', tLatL, tLatR)
	cyrResult := processLetters(cyrLetters, 'а', tCyrL, tCyrR)

	// Восстанавливаем вывод (позиции не-букв сохраняются)
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
	// Если при нечётном количестве букв был добавлен заполнитель — дописываем его
	for ; latIdx < len(latResult); latIdx++ {
		sb.WriteRune(latResult[latIdx])
	}
	for ; cyrIdx < len(cyrResult); cyrIdx++ {
		sb.WriteRune(cyrResult[cyrIdx])
	}

	return sb.String(), nil
}

// Вспомогательные функции ввода/вывода

var stdinScanner = bufio.NewScanner(os.Stdin)

func readLine(prompt string) string {
	fmt.Print(prompt)
	stdinScanner.Scan()
	return stdinScanner.Text()
}

func printTables(keyL, keyR string) {
	fmt.Println("\nТаблицы шифра:")
	tLatL := buildTable(keyL, 5, 5, latinAlphabet, normLatin)
	tLatR := buildTable(keyR, 5, 5, latinAlphabet, normLatin)
	tCyrL := buildTable(keyL, 4, 8, cyrillicAlphabet, normCyrillic)
	tCyrR := buildTable(keyR, 4, 8, cyrillicAlphabet, normCyrillic)
	tLatL.printTable("Левая  (ключ 1)")
	tLatR.printTable("Правая (ключ 2)")
	tCyrL.printTable("Левая  (ключ 1)")
	tCyrR.printTable("Правая (ключ 2)")
	fmt.Println()
}

func main() {
	fmt.Println()

	for {
		fmt.Println("1 — Зашифровать / Расшифровать")
		fmt.Println("2 — Показать таблицы по ключам")
		fmt.Println("0 — Выход")
		choice := strings.TrimSpace(readLine(": "))

		switch choice {
		case "1":
			text := readLine("Введите текст:         ")
			keyL := strings.TrimSpace(readLine("Введите ключ 1 (левой таблицы):  "))
			keyR := strings.TrimSpace(readLine("Введите ключ 2 (правой таблицы): "))

			result, err := process(text, keyL, keyR)
			if err != nil {
				fmt.Println("Ошибка:", err)
				continue
			}
			fmt.Printf("\nРезультат: %s\n\n", result)

		case "2":
			keyL := strings.TrimSpace(readLine("Введите ключ 1 (левой таблицы):  "))
			keyR := strings.TrimSpace(readLine("Введите ключ 2 (правой таблицы): "))
			printTables(keyL, keyR)

		case "0":
			fmt.Println("Выход.")
			return

		default:
			fmt.Println("Неверный выбор, попробуйте снова.")
		}
	}
}
