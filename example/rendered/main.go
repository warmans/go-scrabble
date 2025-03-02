package main

import (
	"fmt"
	"github.com/warmans/go-scrabble"
	"os"
)

func main() {
	game := scrabble.NewGame()
	if err := game.AddPlayer("player 1"); err != nil {
		panic(err)
	}
	if err := game.AddPlayer("player 2"); err != nil {
		panic(err)
	}
	if err := game.AddPlayer("player 3"); err != nil {
		panic(err)
	}
	game.Players[0].Letters = []rune{'F', 'O', 'O', 'F'}
	if err := game.PlaceWord(scrabble.MustParsePlacement("A113"), "foof"); err != nil {
		panic(err)
	}
	fmt.Println("Letters remaining: ", scrabble.RuneSliceAsString(game.Players[0].Letters))
	fmt.Println("Score: ", game.Players[0].Score)

	game.Players[1].Letters = []rune{'F', 'O', 'O', 'F'}
	if err := game.PlaceWord(scrabble.MustParsePlacement("D116"), "foof"); err != nil {
		panic(err)
	}

	//game.Players[2].Letters = []rune{'F', 'O', 'O', 'F'}
	//if err := game.PlaceWord(scrabble.MustParsePlacement("A145"), "foof"); err != nil {
	//	panic(err)
	//}

	scrabble.PrintGame(game, os.Stdout)

	canvas, err := scrabble.RenderPNG(game, 1500, 1000)
	if err != nil {
		panic(err)
	}
	if err := canvas.SavePNG("./example-1.png"); err != nil {
		panic(err)
	}
}
