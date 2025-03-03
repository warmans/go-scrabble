package scrabble

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
)

const NumPlayerLetters = 7

type Player struct {
	Name    string
	Letters []rune
	Score   int
}

func (p *Player) getUsedLetters(letters []rune) (map[rune]int, bool) {
	lettermap := map[rune]int{}

	foundAll := true
	for _, l := range p.Letters {
		if _, ok := lettermap[l]; ok {
			lettermap[l]++
		} else {
			lettermap[l] = 1
		}
	}
	for _, want := range letters {
		found := false
		if avail, ok := lettermap[want]; ok && avail > 0 {
			found = true
			lettermap[want]--
		}
		if !found {
			// if they have a blank they can use that instead for any letter
			if avail, ok := lettermap['_']; ok && avail > 0 {
				found = true
				lettermap['_']--
			}
		}
		if !found {
			foundAll = false
		}
	}

	return lettermap, foundAll
}

func (p *Player) hasLetters(letters []rune) bool {
	_, foundAll := p.getUsedLetters(letters)
	return foundAll
}

func (p *Player) removeLetters(letters []rune) error {
	newLetters := make([]rune, 0, 7)
	usageMap, foundAll := p.getUsedLetters(letters)
	if !foundAll {
		return fmt.Errorf("all letters were not found to be removed")
	}
	for letter, avail := range usageMap {
		for range avail {
			newLetters = append(newLetters, letter)
		}
	}
	p.Letters = newLetters
	return nil
}

func NewClassicGame() *Classic {
	game := &Classic{
		Board:         NewBoard(15),
		CurrentPlayer: 0,
		SpareLetters:  makeLetterBag(),
		Players:       make([]*Player, 0),
	}

	return game
}

type Classic struct {
	Board          Board
	Players        []*Player
	CurrentPlayer  int
	SpareLetters   []rune
	NumWordsPlaced int
	Complete       bool
}

func (g *Classic) AddPlayer(name string) error {
	g.Players = append(g.Players, &Player{Name: name, Letters: make([]rune, 0)})
	if err := g.refillPlayerLetters(len(g.Players) - 1); err != nil {
		return err
	}
	return nil
}

// PlaceWord places a word on the game, the word must be a whole word even if it is just adding letters
// to an existing word. Any exising letters are not spent by the player.
func (g *Classic) PlaceWord(place Placement, word string) error {
	word = strings.ToUpper(word)

	player, err := g.GetCurrentPlayer()
	if err != nil {
		return err
	}

	// is the word valid
	result, err := g.Board.isValidWordPlacement(place, word, g.NumWordsPlaced == 0)
	if err != nil {
		return err
	}

	// do they have the letters required to make the word considering overlaps
	if !player.hasLetters(result.LettersSpent) {
		return fmt.Errorf("player does not have all letters of word: %s", word)
	}

	// spend the letters
	if err := player.removeLetters(result.LettersSpent); err != nil {
		return err
	}

	// update the board
	_, err = g.Board.placeWord(place, word)
	if err != nil {
		// we're in trouble here because the letters have already been removed from the player
		return err
	}

	if err := g.refillPlayerLetters(g.CurrentPlayer); err != nil {
		return err
	}

	// scoring
	player.Score += result.Score()

	g.NextPlayer()

	g.NumWordsPlaced++
	return nil
}

func (g *Classic) getPlayer(idx int) (*Player, error) {
	for k, v := range g.Players {
		if k == idx {
			return v, nil
		}
	}
	return nil, fmt.Errorf("unknown player index: %d", idx)
}

func (g *Classic) GetCurrentPlayer() (*Player, error) {
	for k, v := range g.Players {
		if k == g.CurrentPlayer {
			return v, nil
		}
	}
	return nil, fmt.Errorf("unknown current player index: %d", g.CurrentPlayer)
}

func (g *Classic) getCurrentPlayerName() string {
	player, err := g.GetCurrentPlayer()
	if err != nil {
		return "Unknown"
	}
	return player.Name
}

func (g *Classic) getNextPlayerIdx() int {
	if g.CurrentPlayer+1 > len(g.Players)-1 {
		return 0 // wrap around
	}
	return g.CurrentPlayer + 1
}

func (g *Classic) NextPlayer() {
	maxAttempts := len(g.Players)
	for maxAttempts > 0 {
		next := g.getNextPlayerIdx()
		if len(g.Players[next].Letters) > 0 {
			g.CurrentPlayer = next
			break
		}
		maxAttempts--
	}
	if maxAttempts == 0 {
		g.Complete = true
	}
}

func (g *Classic) refillPlayerLetters(idx int) error {
	player, err := g.getPlayer(idx)
	if err != nil {
		return err
	}
	for {
		if len(player.Letters)+1 > NumPlayerLetters || len(g.SpareLetters) == 0 {
			return nil
		}
		letterIdx := rand.IntN(len(g.SpareLetters) - 1)
		player.Letters = append(player.Letters, g.SpareLetters[letterIdx])
		g.SpareLetters = slices.Delete(g.SpareLetters, letterIdx, letterIdx+1)
	}
}
