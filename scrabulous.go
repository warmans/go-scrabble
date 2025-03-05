package scrabble

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"
)

type ScrabulousState string

const (
	StateIdle     ScrabulousState = ""
	StateStealing ScrabulousState = "stealing"
)

type Word struct {
	Submitter string
	Word      []rune
	Place     Placement
	Result    *PlacementResult
	Stolen    bool
	joined    []string
}

func (w Word) String() string {
	return fmt.Sprintf("%s -> %s (%s) | %d", w.Submitter, string(w.Word), w.Place.String(), w.Result.Score())
}

type Score struct {
	PlayerName string
	Score      int
	Words      int
}

func NewScrabulousGame(stealTime time.Duration) *Scrabulous {
	game := &Scrabulous{
		StealTime: stealTime,
	}
	game.ResetGame()

	return game
}

type Scrabulous struct {
	Board        Board
	SpareLetters []rune
	Letters      []rune
	PlacedWords  []*Word
	PendingWords []*Word
	PlaceWordAt  *time.Time
	Complete     bool
	GameState    ScrabulousState
	StealTime    time.Duration
}

func (s *Scrabulous) IsPlayerAllowed(playerName string) bool {
	if len(s.PlacedWords) == 0 {
		return true
	}
	return s.PlacedWords[len(s.PlacedWords)-1].Submitter != playerName
}

func (s *Scrabulous) IsNewBestWord(score int) bool {
	if len(s.PendingWords) == 0 {
		return true
	}
	return score > s.PendingWords[len(s.PendingWords)-1].Result.Score()
}

func (s *Scrabulous) BestPendingWord() *Word {
	if len(s.PendingWords) == 0 {
		return nil
	}
	return s.PendingWords[len(s.PendingWords)-1]
}

func (s *Scrabulous) TryPlacePendingWord() error {
	if s.PlaceWordAt != nil && time.Now().After(*s.PlaceWordAt) {
		return s.PlacePendingWord()
	}
	return nil
}

func (s *Scrabulous) CreatePendingWord(place Placement, word string, playerName string) (*PlacementResult, error) {
	word = strings.ToUpper(word)

	// is the word valid
	result, err := s.Board.isValidWordPlacement(place, word, len(s.PlacedWords) == 0)
	if err != nil {
		return nil, err
	}

	// do they have the letters required to make the word considering overlaps
	if !s.haveLetters(result.LettersSpent) {
		return nil, fmt.Errorf("you do not have all letters of word: %s", word)
	}

	// if this is the beginning of a new
	firstPendingWord := false
	if len(s.PendingWords) == 0 {
		s.startStealTime()
		firstPendingWord = true
	}

	// if it's a new best word add it to the end fo the list
	if s.IsNewBestWord(result.Score()) {
		s.PendingWords = append(
			s.PendingWords,
			&Word{
				Word:      []rune(word),
				Submitter: playerName,
				Place:     place,
				Result:    result,
				Stolen:    !firstPendingWord,
			},
		)
		return result, nil
	}

	return nil, nil
}

func (s *Scrabulous) PlacePendingWord() error {
	var best *Word
	for _, v := range s.PendingWords {
		if best == nil || v.Result.Score() > best.Result.Score() {
			best = v
		}
	}
	if best == nil {
		return fmt.Errorf("no pending words")
	}

	result, err := s.Board.placeWord(best.Place, string(best.Word))
	if err != nil {
		return err
	}

	s.PlacedWords = append(s.PlacedWords, best)

	if err := s.removeLetters(result.LettersSpent); err != nil {
		fmt.Printf("Failed to remove letters, this is likely a bug: %s", err.Error())
	}

	s.ResetLetters()

	if len(s.Letters) == 0 && len(s.SpareLetters) == 0 {
		s.Complete = true
	}

	s.setGameIdle()

	return nil
}

func (s *Scrabulous) GetScores() []*Score {
	scores := make([]*Score, 0)
	for _, v := range s.PlacedWords {
		found := false
		for _, score := range scores {

			if v.Submitter == score.PlayerName {
				found = true
				score.Score += v.Result.Score()
				score.Words++
				break
			}
		}
		if !found {
			scores = append(scores, &Score{
				PlayerName: v.Submitter,
				Score:      v.Result.Score(),
				Words:      1,
			})
		}
	}

	slices.SortFunc(scores, func(a, b *Score) int {
		return b.Score - a.Score
	})

	return scores
}

func (s *Scrabulous) setGameIdle() {
	s.PendingWords = make([]*Word, 0)
	s.GameState = StateIdle
	s.PlaceWordAt = nil
}

func (s *Scrabulous) startStealTime() {
	finishAt := time.Now().Add(s.StealTime)
	s.PlaceWordAt = &finishAt
	s.GameState = StateStealing
}

func (s *Scrabulous) haveLetters(letters []rune) bool {
	_, foundAll := s.getUsedLetters(letters)
	return foundAll
}

func (s *Scrabulous) ResetLetters() {
	// return any letters to pool
	for _, v := range s.Letters {
		s.SpareLetters = append(s.SpareLetters, v)
	}

	// reset
	s.Letters = make([]rune, 0)

	// add new ones from the pool
	for {
		if len(s.Letters)+1 > NumPlayerLetters || len(s.SpareLetters) == 0 {
			return
		}
		var letterIdx int
		if len(s.SpareLetters)-1 == 0 {
			letterIdx = 0
		} else {
			letterIdx = rand.IntN(len(s.SpareLetters) - 1)
		}
		s.Letters = append(s.Letters, s.SpareLetters[letterIdx])
		s.SpareLetters = slices.Delete(s.SpareLetters, letterIdx, letterIdx+1)
	}
}

func (s *Scrabulous) GetLastPendingWord() *Word {
	if len(s.PendingWords) == 0 {
		return nil
	}
	return s.PendingWords[len(s.PendingWords)-1]
}

func (s *Scrabulous) ResetGame() {
	s.Board = NewBoard(15)
	s.SpareLetters = makeLetterBag()
	s.PlacedWords = make([]*Word, 0)
	s.PendingWords = make([]*Word, 0)
	s.PlaceWordAt = nil
	s.GameState = StateIdle
	s.Complete = false
	s.ResetLetters()
}

func (s *Scrabulous) removeLetters(letters []rune) error {
	newLetters := make([]rune, 0, 7)
	usageMap, foundAll := s.getUsedLetters(letters)
	if !foundAll {
		return fmt.Errorf("all letters were not found to be removed")
	}
	for letter, avail := range usageMap {
		for range avail {
			newLetters = append(newLetters, letter)
		}
	}
	s.Letters = newLetters
	return nil
}

func (s *Scrabulous) getUsedLetters(word []rune) (map[rune]int, bool) {
	lettermap := map[rune]int{}

	foundAll := true
	for _, l := range s.Letters {
		if _, ok := lettermap[l]; ok {
			lettermap[l]++
		} else {
			lettermap[l] = 1
		}
	}
	for _, want := range word {
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
