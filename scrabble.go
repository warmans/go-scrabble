package scrabble

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
)

const NumPlayerLetters = 7

type CellState string

const (
	CellFull  CellState = "full"
	CellEmpty CellState = "empty"
	CellAny   CellState = ""
)

type Cell struct {
	Index int
	Char  rune
}

type Overlap struct {
	L bool // top
	R bool // right
	A bool // above
	B bool // below
}

type Board [][]Cell

func (b Board) getCellIndex(placement Placement, letterIdx int) (int64, error) {
	var cellIndex int64
	if placement.Direction == Down {
		// the board is a square so the next down square we should be able to just add the length of a row.
		// no need to check if the word spans multiple columns because the IDs go from L2R.
		// if the letterIdx is negative then move backwards.
		verticalOffset := int64(len(b) * letterIdx)
		if letterIdx < 0 {
			verticalOffset = 0 - verticalOffset
		}
		cellIndex = placement.CellId + verticalOffset
	} else {
		cellIndex = placement.CellId + int64(letterIdx)
		// check that all letters are on the same row as the first one
		if math.Ceil(float64(cellIndex)/float64(len(b))) != math.Ceil(float64(placement.CellId)/float64(len(b))) {
			return 0, fmt.Errorf("word cannot span multiple rows")
		}
	}
	return cellIndex, nil
}

func (b Board) isValidWordPlacement(placement Placement, word string) ([]rune, int, error) {

	lettersSpent := []rune{}
	overlaps := 0
	for i, letter := range word {
		isOverlapping := false

		cellIndex, err := b.getCellIndex(placement, i)
		if err != nil {
			return nil, 0, err
		}
		// 1. does the word fit within the board
		cell := b.GetCell(cellIndex, CellAny)
		if cell == nil {
			return nil, 0, fmt.Errorf("word does not fit on the board")
		}

		// 2. is there a valid overlap or empty space
		if cell.Char != 0 && cell.Char != letter {
			return nil, 0, fmt.Errorf("invalid overlap, %s cannot be placed on %s", string(letter), string(cell.Char))
		}
		if cell.Char == letter {
			overlaps++
			isOverlapping = true
		} else {
			lettersSpent = append(lettersSpent, letter)
		}

		neighbours := b.nonEmptyNeighbouringCells(cellIndex)

		// 3. Does the word fit into the given placement considering neighbouring cells
		if placement.Direction == Across {
			if i == 0 && neighbours.L {
				return nil, 0, fmt.Errorf("first letter is too close to another word")
			}
			if i == len(word)-1 && neighbours.R {
				return nil, 0, fmt.Errorf("last letter is too close to another word")
			}
			// if the letter is overlapping there will be other letters touching the sides
			if !isOverlapping && (neighbours.A || neighbours.B) {
				return nil, 0, fmt.Errorf("another word is too close above or below")
			}
		} else {
			if i == 0 && neighbours.A {
				return nil, 0, fmt.Errorf("first letter is too close to another word")
			}
			if i == len(word)-1 && neighbours.B {
				return nil, 0, fmt.Errorf("last letter is too close to another word")
			}
			if !isOverlapping && (neighbours.R || neighbours.L) {
				return nil, 0, fmt.Errorf("another word is too close to the left or right")
			}
		}
	}
	if overlaps == len(word) {
		return nil, 0, fmt.Errorf("word completely overlaps another word")
	}

	return lettersSpent, overlaps, nil
}

func (b Board) placeWord(placement Placement, word string) error {
	for i, letter := range word {
		cellIndex, err := b.getCellIndex(placement, i)
		if err != nil {
			return err
		}

		b.SetCell(cellIndex, letter)
	}
	return nil
}

func (b Board) GetCell(cellID int64, state CellState) *Cell {
	// cells indexes start at 1
	if cellID < 1 {
		return nil
	}
	for _, row := range b {
		for _, cell := range row {
			if int64(cell.Index) == cellID {
				if state == CellEmpty && cell.Char != 0 {
					return nil
				}
				if state == CellFull && cell.Char == 0 {
					return nil
				}
				return &cell
			}
		}
	}
	return nil
}

func (b Board) SetCell(cellID int64, letter rune) bool {
	curID := int64(0)
	set := false
	for rowIdx, row := range b {
		for colIdx := range row {
			curID++
			if curID == cellID {
				if b[rowIdx][colIdx].Char == 0 {
					set = true
				}
				b[rowIdx][colIdx].Char = letter
			}
		}
	}
	// todo: perhaps it should return any bonuses for "set" letters here?
	// e.g. return [doubleLetterScore(letter), doubleWordScore()]
	return set
}

func (b Board) nonEmptyNeighbouringCells(cellID int64) Overlap {
	return Overlap{
		L: b.GetCell(cellID-1, CellFull) != nil,
		R: b.GetCell(cellID+1, CellFull) != nil,
		A: b.GetCell(cellID-int64(len(b)), CellFull) != nil,
		B: b.GetCell(cellID+int64(len(b)), CellFull) != nil,
	}
}

func NewBoard(size int) Board {
	grid := make(Board, size)
	for y := range size {
		grid[y] = make([]Cell, size)
	}
	index := 1
	for rowNum := range grid {
		for colNum := range grid[rowNum] {
			grid[rowNum][colNum].Index = index
			index++
		}
	}
	return grid
}

type Player struct {
	Name    string
	Letters []rune
}

func (p *Player) hasLetters(letters []rune) bool {
	numBlanks := 0
	for _, v := range p.Letters {
		if v == '_' {
			numBlanks++
		}
	}
	matchedLetters := 0
	for _, v := range letters {
		found := false
		for _, l := range p.Letters {
			if l == v {
				found = true
			}
		}
		if found {
			matchedLetters++
		}
	}
	return matchedLetters == len(letters) || matchedLetters+numBlanks > len(letters)
}

func (p *Player) removeLetters(letters []rune) error {
	originalLetters := make([]rune, 0, 7)
	copy(originalLetters, p.Letters)

	removed := 0
	for _, l := range letters {
		newLetters := make([]rune, 0, 7)
		for _, j := range p.Letters {
			if l != j {
				newLetters = append(newLetters, j)
			} else {
				removed++
			}
		}
		p.Letters = newLetters
	}

	// apparently there are blanks to be removed
	blanksRemoved := 0
	numBlanksToRemove := len(letters) - removed
	for range numBlanksToRemove {
		newLetters := make([]rune, 0, 7)
		for _, j := range p.Letters {
			if j == '_' {
				blanksRemoved++
			} else {
				newLetters = append(newLetters, j)
			}
		}
	}

	if numBlanksToRemove > blanksRemoved {
		// something is fucked up here, because we should have already asserted that the player has all the
		// required letters.
		// revert the changes and return an error.
		p.Letters = originalLetters
		return fmt.Errorf("not enough letters, or blanks to remove")
	}
	return nil
}

func ParsePlacement(placementStr string) (Placement, error) {
	p := Placement{}
	if strings.HasSuffix(placementStr, "D") {
		p.Direction = "D"
	} else if strings.HasSuffix(placementStr, "A") {
		p.Direction = "A"
	} else {
		return p, fmt.Errorf("placement must start with either D or A")
	}

	var err error
	p.CellId, err = strconv.ParseInt(placementStr[1:], 10, 32)
	if err != nil {
		return p, fmt.Errorf("failed to parse placement cell index: %w", err)
	}

	return p, nil
}

type Direction string

const (
	Down   Direction = "D"
	Across Direction = "A"
)

type Placement struct {
	CellId    int64
	Direction Direction
}

func (p Placement) String() string {
	return fmt.Sprintf("%s%d", p.Direction, p.CellId)
}

func NewGame() *Game {
	game := &Game{
		Board:         NewBoard(15),
		CurrentPlayer: 0,
		SpareLetters:  makeLetterBag(),
		Players:       make([]*Player, 0),
	}

	return game
}

type Game struct {
	Board          Board
	Players        []*Player
	CurrentPlayer  int
	SpareLetters   []rune
	NumWordsPlaced int
}

func (g *Game) AddPlayer(name string) error {
	g.Players = append(g.Players, &Player{Name: name, Letters: make([]rune, 0)})
	if err := g.refillPlayerLetters(name); err != nil {
		return err
	}
	return nil
}

// PlaceWord places a word on the game, the word must be a whole word even if it is just adding letters
// to an existing word. Any exising letters are not spent by the player.
func (g *Game) PlaceWord(place Placement, playerName string, word string) error {
	player, err := g.getPlayer(playerName)
	if err != nil {
		return err
	}

	// is the word valid
	lettersRequired, overlaps, err := g.Board.isValidWordPlacement(place, word)
	if err != nil {
		return err
	}

	if g.NumWordsPlaced > 0 && overlaps == 0 {
		return fmt.Errorf("word must overlap at least one other word")
	}

	// do they have the letters required to make the word considering overlaps
	if !player.hasLetters(lettersRequired) {
		return fmt.Errorf("player does not have all letters of word: %s", word)
	}

	// spend the letters
	if err := player.removeLetters(lettersRequired); err != nil {
		return err
	}

	// update the board
	if err := g.Board.placeWord(place, word); err != nil {
		// we're in trouble here because the letters have already been removed from the player
		return err
	}

	if err := g.refillPlayerLetters(playerName); err != nil {
		return err
	}

	//todo: if len(requiredLetters) == 7 they get an additional 50 points

	g.NumWordsPlaced++
	return nil
}

func (g *Game) getPlayer(name string) (*Player, error) {
	for _, v := range g.Players {
		if v.Name == name {
			return v, nil
		}
	}
	return nil, fmt.Errorf("unknown player: %s", name)
}

func (g *Game) refillPlayerLetters(name string) error {
	player, err := g.getPlayer(name)
	if err != nil {
		return err
	}
	for {
		if len(player.Letters)+1 > NumPlayerLetters || len(g.SpareLetters) == 0 {
			return nil
		}
		letterIdx := rand.IntN(len(g.SpareLetters) - 1)
		player.Letters = append(player.Letters, g.SpareLetters[letterIdx])
		g.SpareLetters = slices.Delete(g.SpareLetters, letterIdx, letterIdx)
	}
}
