package scrabble

var LetterScores = map[rune]int{
	'A': 1,
	'B': 3,
	'C': 3,
	'D': 2,
	'E': 1,
	'F': 4,
	'G': 2,
	'H': 4,
	'I': 1,
	'J': 8,
	'K': 5,
	'L': 1,
	'M': 3,
	'N': 1,
	'O': 1,
	'P': 3,
	'Q': 10,
	'R': 1,
	'S': 1,
	'T': 1,
	'U': 1,
	'V': 4,
	'W': 4,
	'X': 8,
	'Y': 4,
	'Z': 10,
	'_': 0,
}

var LetterDistribution = map[rune]int{
	'A': 9,
	'B': 2,
	'C': 2,
	'D': 4,
	'E': 12,
	'F': 2,
	'G': 3,
	'H': 2,
	'I': 9,
	'J': 1,
	'K': 1,
	'L': 4,
	'M': 2,
	'N': 6,
	'O': 8,
	'P': 2,
	'Q': 1,
	'R': 6,
	'S': 4,
	'T': 6,
	'U': 4,
	'V': 2,
	'W': 2,
	'X': 1,
	'Y': 2,
	'Z': 1,
	'_': 2,
}

// StandardBonusMap is a map of bonuses by cell index
var StandardBonusMap = map[int]CellBonusType{
	1:   TripleWordScoreType,
	8:   TripleWordScoreType,
	15:  TripleWordScoreType,
	106: TripleWordScoreType,
	120: TripleWordScoreType,
	211: TripleWordScoreType,
	218: TripleWordScoreType,
	225: TripleWordScoreType,

	// Triple Letter Scores
	21:  TripleLetterScoreType,
	25:  TripleLetterScoreType,
	81:  TripleLetterScoreType,
	85:  TripleLetterScoreType,
	141: TripleLetterScoreType,
	145: TripleLetterScoreType,
	201: TripleLetterScoreType,
	205: TripleLetterScoreType,
	77:  TripleLetterScoreType,
	137: TripleLetterScoreType,
	89:  TripleLetterScoreType,
	149: TripleLetterScoreType,

	// Double Word Score
	17:  DoubleWordScoreType,
	33:  DoubleWordScoreType,
	49:  DoubleWordScoreType,
	65:  DoubleWordScoreType,
	29:  DoubleWordScoreType,
	43:  DoubleWordScoreType,
	57:  DoubleWordScoreType,
	71:  DoubleWordScoreType,
	197: DoubleWordScoreType,
	183: DoubleWordScoreType,
	169: DoubleWordScoreType,
	155: DoubleWordScoreType,
	209: DoubleWordScoreType,
	193: DoubleWordScoreType,
	177: DoubleWordScoreType,
	161: DoubleWordScoreType,

	// Double Letter Score
	4:   DoubleLetterScoreType,
	12:  DoubleLetterScoreType,
	214: DoubleLetterScoreType,
	222: DoubleLetterScoreType,
	46:  DoubleLetterScoreType,
	166: DoubleLetterScoreType,
	60:  DoubleLetterScoreType,
	180: DoubleLetterScoreType,
	93:  DoubleLetterScoreType,
	123: DoubleLetterScoreType,
	109: DoubleLetterScoreType,
	103: DoubleLetterScoreType,
	133: DoubleLetterScoreType,
	117: DoubleLetterScoreType,
	97:  DoubleLetterScoreType,
	99:  DoubleLetterScoreType,
	127: DoubleLetterScoreType,
	129: DoubleLetterScoreType,
	37:  DoubleLetterScoreType,
	53:  DoubleLetterScoreType,
	39:  DoubleLetterScoreType,
	187: DoubleLetterScoreType,
	189: DoubleLetterScoreType,
	173: DoubleLetterScoreType,
}

func makeLetterBag() []rune {
	bag := []rune{}
	for letter, count := range LetterDistribution {
		bag = append(bag, repeatLetter(letter, count)...)
	}
	return bag
}

func repeatLetter(letter rune, num int) []rune {
	letters := make([]rune, num)
	for i := range num {
		letters[i] = letter
	}
	return letters
}
