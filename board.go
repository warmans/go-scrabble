package scrabble

import (
	"fmt"
	"math"
	"slices"
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

type Orientation string

const (
	Down   Orientation = "D"
	Across Orientation = "A"
)

type Direction string

const (
	U Direction = "up"
	D Direction = "down"
	L Direction = "left"
	R Direction = "right"
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
	L bool // left
	R bool // right
	A bool // above
	B bool // below
}

type Board [][]Cell

func (b Board) getCellIndex(placement Placement, offset int) int64 {
	if placement.Direction == Down {
		return b.getNextVerticalCellId(placement.CellId, offset)
	}
	return b.getNextHorizontalCellId(placement.CellId, offset)
}

func (b Board) getNextVerticalCellId(cellID int64, offset int) int64 {
	// the board is a square so the next down square we should be able to just add the length of a row.
	// no need to check if the word spans multiple columns because the IDs go from L2R.
	// if the offset is negative then move backwards.
	verticalOffset := int64(len(b) * offset)
	if offset < 0 {
		verticalOffset = 0 - verticalOffset
	}
	return cellID + verticalOffset
}

func (b Board) getNextHorizontalCellId(cellID int64, offset int) int64 {
	cellIndex := cellID + int64(offset)
	// check that all letters are on the same row as the first one
	if math.Ceil(float64(cellIndex)/float64(len(b))) != math.Ceil(float64(cellID)/float64(len(b))) {
		return -1
	}
	return cellIndex
}

func (b Board) isValidWordPlacement(placement Placement, word string, firstWord bool) (*PlacementResult, error) {

	overlaps := 0
	cellsCovered := []int64{}
	result := &PlacementResult{
		Cells:        make([]Cell, 0),
		LettersSpent: make([]rune, 0),
		Touching:     make([][]Cell, 0),
	}

	for i, letter := range []rune(word) {
		isOverlapping := false

		cellIndex := b.getCellIndex(placement, i)
		if cellIndex == -1 {
			return nil, fmt.Errorf("word had invalid cell range")
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

		thisCell := Cell{Index: cell.Index, Char: letter, Bonus: cell.Bonus}

		// add to result
		result.Cells = append(
			result.Cells,
			thisCell,
		)

		neighbours := b.nonEmptyNeighbouringCells(cellIndex)

		// if this is the last letter of a DOWN word, there cannot be any letters directly below
		if placement.Direction == Down && i == len(word)-1 && neighbours.B {
			return nil, fmt.Errorf("word below %s is too close", string(letter))
		}
		if placement.Direction == Down && i == 0 && neighbours.A {
			return nil, fmt.Errorf("word above %s is too close", string(letter))
		}
		if placement.Direction == Across && i == len(word)-1 && neighbours.R {
			return nil, fmt.Errorf("word to the right of %s is too close", string(letter))
		}
		if placement.Direction == Across && i == 0 && neighbours.L {
			return nil, fmt.Errorf("word to the left of %s is too close", string(word))
		}

		// words can only join for non-overlapping letters
		if !isOverlapping {

			// get joined horizontal words
			if placement.Direction != Across {
				touchingX := make([]Cell, 0)
				if neighbours.L {
					lhs := b.NeighboringWord(int64(cell.Index), L)
					slices.Reverse(lhs)
					touchingX = append(touchingX, lhs...)
				}
				if neighbours.L || neighbours.R {
					if i > 0 && placement.Direction == Across {
						touchingX = append(touchingX, result.Cells...)
					} else {
						touchingX = append(touchingX, thisCell)
					}
				}
				if neighbours.R {
					touchingX = append(touchingX, b.NeighboringWord(int64(cell.Index), R)...)
				}
				if len(touchingX) > 0 {
					result.Touching = append(result.Touching, touchingX)
				}
			}

			// get joined vertical words

			if placement.Direction != Down {
				touchingY := make([]Cell, 0)
				if neighbours.A {
					lhs := b.NeighboringWord(int64(cell.Index), U)
					slices.Reverse(lhs)
					touchingY = append(lhs, *cell)
				}
				if neighbours.A || neighbours.B {
					touchingY = append(touchingY, thisCell)
				}
				if neighbours.B {
					touchingY = append(touchingY, b.NeighboringWord(int64(cell.Index), D)...)
				}
				if len(touchingY) > 0 {
					result.Touching = append(result.Touching, touchingY)
				}
			}
		}
	}
	if overlaps == len(word) {
		return nil, fmt.Errorf("word completely overlaps another word")
	}

	if !firstWord && overlaps == 0 && len(result.Touching) == 0 {
		return nil, fmt.Errorf("word must overlap or touch at least one other word")
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
		cellIndex := b.getCellIndex(placement, i)
		if cellIndex == -1 {
			return nil, fmt.Errorf("word has invalid cell range")
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

func (b Board) NeighboringWord(cellID int64, direction Direction) []Cell {
	cells := make([]Cell, 0)
	offset := 0
	for {
		if direction == L || direction == U {
			offset -= 1
		} else {
			offset += 1
		}
		var nextCellId int64
		if direction == U || direction == D {
			nextCellId = b.getNextVerticalCellId(cellID, offset)
		} else {
			nextCellId = b.getNextHorizontalCellId(cellID, offset)
		}
		if nextCellId < 0 {
			return cells
		}
		nextCell := b.GetCell(nextCellId, CellFull)
		if nextCell == nil {
			return cells
		}
		cells = append(cells, *nextCell)
	}
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
	Direction Orientation
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
	Touching     [][]Cell
}

func (r *PlacementResult) Score() int {
	total := 0
	words := [][]Cell{r.Cells}
	if len(r.Touching) > 0 {
		words = append(words, r.Touching...)
	}
	for _, word := range words {
		if len(word) == 1 {
			continue
		}
		var wordTotal int
		var wordBonuses []CellBonusType
		for _, c := range word {
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
		total = total + wordTotal
	}
	if len(r.LettersSpent) == NumPlayerLetters {
		total += 50
	}
	return total
}
