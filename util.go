package scrabble

import "fmt"

func dumpLetters(letters []rune) string {
	out := ""
	for _, v := range letters {
		out = out + string(v)
	}
	return out
}

func dumpLetterMap(letters map[rune]int) string {
	out := ""
	for letter, avail := range letters {
		out = out + fmt.Sprintf("%s:%d", string(letter), avail)
	}
	return out
}
