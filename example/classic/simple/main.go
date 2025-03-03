package main

import (
	"fmt"
	"github.com/warmans/go-scrabble"
	"os"
)

const defaultPlayerName = "default"

func main() {
	game := scrabble.NewClassicGame()
	if err := game.AddPlayer(defaultPlayerName); err != nil {
		panic(err)
	}
	game.Players[0].Letters = []rune{'F', 'O', 'O', 'F'}
	if err := game.PlaceWord(scrabble.MustParsePlacement("A113"), "foof"); err != nil {
		panic(err)
	}
	fmt.Println("Letters remaining: ", scrabble.RuneSliceAsString(game.Players[0].Letters))
	fmt.Println("Score: ", game.Players[0].Score)

	game.Players[0].Letters = []rune{'F', 'O', 'O', 'F'}
	if err := game.PlaceWord(scrabble.MustParsePlacement("D116"), "foof"); err != nil {
		panic(err)
	}

	game.Players[0].Letters = []rune{'F', 'O', 'O', 'F'}
	if err := game.PlaceWord(scrabble.MustParsePlacement("A145"), "foof"); err != nil {
		panic(err)
	}

	scrabble.PrintGame(game, os.Stdout)
}
