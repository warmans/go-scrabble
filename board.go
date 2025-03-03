package scrabble

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type CellState string

const (
	CellFull  CellState = "full"
	CellEmpty CellState = "empty"
	CellAny   CellState = ""
)

type CellBonusType string

const (
	NoBonusType           CellBonusType = ""
	DoubleLetterScoreType CellBonusType = "double_letter_score"
	DoubleWordScoreType   CellBonusType = "double_word_score"
	TripleLetterScoreType CellBonusType = "triple_letter_score"
	TripleWordScoreType   CellBonusType = "triple_word_score"
)

type Direction string

const (
	Down   Direction = "D"
	Across Direction = "A"
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

func (b Board) isValidWordPlacement(placement Placement, word string, firstWord bool) (*PlacementResult, error) {

	overlaps := 0
	cellsCovered := []int64{}
	result := &PlacementResult{Cells: make([]Cell, 0), LettersSpent: make([]rune, 0)}

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
		if !cell.Empty() && cell.Char != letter {
			return nil, fmt.Errorf("invalid overlap, %s cannot be placed on %s", string(letter), string(cell.Char))
		}
		if cell.Char == letter {
			// letter already exits
			overlaps++
			isOverlapping = true
		} else {
			// user must have letter
			result.LettersSpent = append(result.LettersSpent, letter)
		}

		// add to result
		result.Cells = append(
			result.Cells,
			Cell{Index: cell.Index, Char: letter, Bonus: cell.Bonus},
		)

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

	return result, nil
}

func (b Board) placeWord(placement Placement, word string) (*PlacementResult, error) {
	result := &PlacementResult{Cells: make([]Cell, 0)}
	for i, letter := range word {
		cellIndex, err := b.getCellIndex(placement, i)
		if err != nil {
			return nil, err
		}

		cell, placed := b.SetCell(cellIndex, letter)
		result.Cells = append(result.Cells, cell)

		if placed {
			result.LettersSpent = append(result.LettersSpent, letter)
		}

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

type Placement struct {
	CellId    int64
	Direction Direction
}

func (p Placement) String() string {
	return fmt.Sprintf("%s%d", p.Direction, p.CellId)
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

type PlacementResult struct {
	Cells        []Cell
	LettersSpent []rune
}

func (r *PlacementResult) Score() int {
	wordTotal := 0
	var wordBonuses []CellBonusType
	for _, c := range r.Cells {
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
	if len(r.Cells) == NumPlayerLetters {
		wordTotal += 50
	}
	return wordTotal
}
