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

type PlacementResult struct {
	cells []Cell
}

func (r *PlacementResult) Score() int {
	wordTotal := 0
	var wordBonuses []CellBonusType
	for _, c := range r.cells {
		letterScore := LetterScores[c.Char]
		switch c.Bonus {
		case NoBonusType:
			wordTotal += letterScore
		case DoubleLetterScoreType:
			wordTotal += letterScore * 2
		case TripleLetterScoreType:
			wordTotal += letterScore * 3
		case DoubleWordScoreType:
			wordTotal += letterScore
			wordBonuses = append(wordBonuses, c.Bonus)
		case TripleWordScoreType:
			wordTotal += letterScore
			wordBonuses = append(wordBonuses, c.Bonus)
		}
	}
	for _, b := range wordBonuses {
		switch b {
		case DoubleWordScoreType:
			wordTotal = wordTotal * 2
		case TripleWordScoreType:
			wordTotal = wordTotal * 3
		}
	}
	if len(r.cells) == NumPlayerLetters {
		wordTotal += 50
	}
	return wordTotal
}

type CellBonusType string

const (
	NoBonusType           CellBonusType = ""
	DoubleLetterScoreType CellBonusType = "double_letter_score"
	DoubleWordScoreType   CellBonusType = "double_word_score"
	TripleLetterScoreType CellBonusType = "triple_letter_score"
	TripleWordScoreType   CellBonusType = "triple_word_score"
)

type Cell struct {
	Index int
	Char  rune
	Bonus CellBonusType
}

func (c Cell) Empty() bool {
	return c.Char == 0
}

func (c Cell) String() string {
	return string(c.Char)
}

func (c Cell) IndexString() string {
	return fmt.Sprintf("%d", c.Index)
}

func (c Cell) LetterScoreString() string {
	return fmt.Sprintf("%d", LetterScores[c.Char])
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

func (b Board) isValidWordPlacement(placement Placement, word string, firstWord bool) ([]rune, error) {

	lettersSpent := []rune{}
	overlaps := 0
	cellsCovered := []int64{}

	for i, letter := range word {
		isOverlapping := false

		cellIndex, err := b.getCellIndex(placement, i)
		if err != nil {
			return nil, err
		}
		cellsCovered = append(cellsCovered, cellIndex)
		// 1. does the word fit within the board
		cell := b.GetCell(cellIndex, CellAny)
		if cell == nil {
			return nil, fmt.Errorf("word does not fit on the board")
		}

		// 2. is there a valid overlap or empty space
		if cell.Char != 0 && cell.Char != letter {
			return nil, fmt.Errorf("invalid overlap, %s cannot be placed on %s", string(letter), string(cell.Char))
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
				return nil, fmt.Errorf("first letter is too close to another word")
			}
			if i == len(word)-1 && neighbours.R {
				return nil, fmt.Errorf("last letter is too close to another word")
			}
			// if the letter is overlapping there will be other letters touching the sides
			if !isOverlapping && (neighbours.A || neighbours.B) {
				return nil, fmt.Errorf("another word is too close above or below")
			}
		} else {
			if i == 0 && neighbours.A {
				return nil, fmt.Errorf("first letter is too close to another word")
			}
			if i == len(word)-1 && neighbours.B {
				return nil, fmt.Errorf("last letter is too close to another word")
			}
			if !isOverlapping && (neighbours.R || neighbours.L) {
				return nil, fmt.Errorf("another word is too close to the left or right")
			}
		}
	}
	if overlaps == len(word) {
		return nil, fmt.Errorf("word completely overlaps another word")
	}

	if !firstWord && overlaps == 0 {
		return nil, fmt.Errorf("word must overlap at least one other word")
	}

	if firstWord {
		centerCell := b.getCenterCellIdx()
		centerCellCovered := false
		for _, cellIdx := range cellsCovered {
			if cellIdx == centerCell {
				centerCellCovered = true
			}
		}
		if !centerCellCovered {
			return nil, fmt.Errorf("first word must overlap the center of the board (cell %d)", centerCell)
		}
	}

	return lettersSpent, nil
}

func (b Board) placeWord(placement Placement, word string) (*PlacementResult, error) {
	result := &PlacementResult{cells: make([]Cell, 0)}
	for i, letter := range word {
		cellIndex, err := b.getCellIndex(placement, i)
		if err != nil {
			return nil, err
		}

		cell, _ := b.SetCell(cellIndex, letter)
		result.cells = append(result.cells, cell)
	}
	return result, nil
}

func (b Board) GetCell(cellID int64, state CellState) *Cell {
	// cells indexes start at 1
	if cellID < 1 {
		return nil
	}
	for _, row := range b {
		for _, cell := range row {
			if int64(cell.Index) == cellID {
				if state == CellEmpty && !cell.Empty() {
					return nil
				}
				if state == CellFull && cell.Empty() {
					return nil
				}
				return &cell
			}
		}
	}
	return nil
}

func (b Board) SetCell(cellID int64, letter rune) (Cell, bool) {
	curID := int64(0)
	set := false
	var cell Cell
	for rowIdx, row := range b {
		for colIdx := range row {
			curID++
			if curID == cellID {
				if b[rowIdx][colIdx].Empty() {
					set = true
				}
				b[rowIdx][colIdx].Char = letter
				cell = b[rowIdx][colIdx]
			}
		}
	}
	return cell, set
}

func (b Board) nonEmptyNeighbouringCells(cellID int64) Overlap {
	return Overlap{
		L: b.GetCell(cellID-1, CellFull) != nil,
		R: b.GetCell(cellID+1, CellFull) != nil,
		A: b.GetCell(cellID-int64(len(b)), CellFull) != nil,
		B: b.GetCell(cellID+int64(len(b)), CellFull) != nil,
	}
}

func (b Board) getCenterCellIdx() int64 {
	size := float64(len(b))
	middle := math.Ceil(size / float64(2))
	return int64(middle + (size * math.Floor(size/float64(2))))
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

			if bonus, ok := StandardBonusMap[index]; ok {
				grid[rowNum][colNum].Bonus = bonus
			}

			index++
		}
	}
	return grid
}

type Player struct {
	Name    string
	Letters []rune
	Score   int
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

func MustParsePlacement(placementStr string) Placement {
	place, err := ParsePlacement(placementStr)
	if err != nil {
		panic(err)
	}
	return place
}

func ParsePlacement(placementStr string) (Placement, error) {
	p := Placement{}
	if strings.HasPrefix(placementStr, "D") {
		p.Direction = "D"
	} else if strings.HasPrefix(placementStr, "A") {
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
	if err := g.refillPlayerLetters(len(g.Players) - 1); err != nil {
		return err
	}
	return nil
}

// PlaceWord places a word on the game, the word must be a whole word even if it is just adding letters
// to an existing word. Any exising letters are not spent by the player.
func (g *Game) PlaceWord(place Placement, word string) error {
	word = strings.ToUpper(word)
	
	player, err := g.getCurrentPlayer()
	if err != nil {
		return err
	}

	// is the word valid
	lettersRequired, err := g.Board.isValidWordPlacement(place, word, g.NumWordsPlaced == 0)
	if err != nil {
		return err
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
	result, err := g.Board.placeWord(place, word)
	if err != nil {
		// we're in trouble here because the letters have already been removed from the player
		return err
	}

	if err := g.refillPlayerLetters(g.CurrentPlayer); err != nil {
		return err
	}

	// scoring
	player.Score += result.Score()

	g.nextPlayer()

	g.NumWordsPlaced++
	return nil
}

func (g *Game) getPlayer(idx int) (*Player, error) {
	for k, v := range g.Players {
		if k == idx {
			return v, nil
		}
	}
	return nil, fmt.Errorf("unknown player index: %d", idx)
}

func (g *Game) getCurrentPlayer() (*Player, error) {
	for k, v := range g.Players {
		if k == g.CurrentPlayer {
			return v, nil
		}
	}
	return nil, fmt.Errorf("unknown current player index: %d", g.CurrentPlayer)
}

func (g *Game) getCurrentPlayerName() string {
	player, err := g.getCurrentPlayer()
	if err != nil {
		return "Unknown"
	}
	return player.Name
}

func (g *Game) getNextPlayerIdx() int {
	if g.CurrentPlayer+1 > len(g.Players)-1 {
		return 0 // wrap around
	}
	return g.CurrentPlayer + 1
}

func (g *Game) nextPlayer() {
	g.CurrentPlayer = g.getNextPlayerIdx()
}

func (g *Game) refillPlayerLetters(idx int) error {
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
		g.SpareLetters = slices.Delete(g.SpareLetters, letterIdx, letterIdx)
	}
}
