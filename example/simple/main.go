package main

import (
	"fmt"
	"github.com/warmans/go-scrabble"
	"os"
)

const defaultPlayerName = "default"

func main() {
	game := scrabble.NewGame()
	if err := game.AddPlayer(defaultPlayerName); err != nil {
		panic(err)
	}
	game.Players[0].Letters = []rune{'f', 'o', 'o', 'f'}
	if err := game.PlaceWord(scrabble.Placement{CellId: 1, Direction: scrabble.Across}, defaultPlayerName, "foof"); err != nil {
		panic(err)
	}
	fmt.Println("Letters remaining: ", scrabble.RuneSliceAsString(game.Players[0].Letters))

	game.Players[0].Letters = []rune{'f', 'o', 'o', 'f'}
	if err := game.PlaceWord(scrabble.Placement{CellId: 4, Direction: scrabble.Down}, defaultPlayerName, "foof"); err != nil {
		panic(err)
	}

	game.Players[0].Letters = []rune{'f', 'o', 'o', 'f'}
	if err := game.PlaceWord(scrabble.Placement{CellId: 33, Direction: scrabble.Across}, defaultPlayerName, "foof"); err != nil {
		panic(err)
	}

	scrabble.PrintGame(game, os.Stdout)
}
