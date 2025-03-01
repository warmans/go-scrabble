package scrabble

import (
	"fmt"
	"io"
)

func PrintGame(game *Game, writer io.Writer) {
	celIdx := 1
	for _, row := range game.Board {
		fmt.Fprintf(writer, "|")
		for _, cell := range row {
			fmt.Fprintf(writer, "%d %s |", celIdx, string(cell.Char))
			celIdx++
		}
		fmt.Fprintf(writer, "\n")
	}
}

func RuneSliceAsString(runes []rune) string {
	out := ""
	for _, v := range runes {
		out += string(v)
	}
	return out
}
