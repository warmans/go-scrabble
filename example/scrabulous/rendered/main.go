package main

import (
	"github.com/warmans/go-scrabble"
	"time"
)

func main() {
	game := scrabble.NewScrabulousGame(time.Minute * 5)

	game.Letters = []rune{'F', 'O', 'O', 'F'}
	if _, err := game.CreatePendingWord(scrabble.MustParsePlacement("A113"), "foof", "player 1"); err != nil {
		panic(err)
	}

	//fmt.Println("Pending...")
	//for _, v := range game.PendingWords {
	//	fmt.Println(v)
	//}

	if err := game.PlacePendingWord(); err != nil {
		panic(err)
	}

	game.Letters = []rune{'F', 'O', 'O', 'F', 'S'}
	if _, err := game.CreatePendingWord(scrabble.MustParsePlacement("D57"), "foofs", "player 2"); err != nil {
		panic(err)
	}
	//for _, v := range game.PendingWords {
	//	fmt.Println(v)
	//}

	if err := game.PlacePendingWord(); err != nil {
		panic(err)
	}

	game.Letters = []rune{'F', 'O', 'O', 'F', 'S'}
	if _, err := game.CreatePendingWord(scrabble.MustParsePlacement("A57"), "FOO", "player 3"); err != nil {
		panic(err)
	}
	if err := game.PlacePendingWord(); err != nil {
		panic(err)
	}

	game.Letters = []rune{'F', 'O', 'O', 'F', 'S'}
	if _, err := game.CreatePendingWord(scrabble.MustParsePlacement("D15"), "SOOFF", "player 4"); err != nil {
		panic(err)
	}
	if err := game.PlacePendingWord(); err != nil {
		panic(err)
	}

	//game.Letters = []rune{'A', 'A'}
	//if _, err := game.CreatePendingWord(scrabble.MustParsePlacement("D86"), "AAF", "player 4"); err != nil {
	//	panic(err)
	//}
	//if err := game.PlacePendingWord(); err != nil {
	//	panic(err)
	//}

	canvas, err := scrabble.RenderScrabulousPNG(game, 1500, 1000)
	if err != nil {
		panic(err)
	}
	if err := canvas.SavePNG("./example-1.png"); err != nil {
		panic(err)
	}
}
