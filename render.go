package scrabble

import (
	"fmt"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/goregular"
	"image/color"
	"log"
	"strings"
	"time"
)

var font *truetype.Font

func init() {
	var err error
	font, err = truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
}

func resolveRenderOptions(opts ...RenderOption) *renderOpts {
	opt := &renderOpts{
		backgroundColor:     color.RGBA{193, 181, 173, 255},
		wordBackgroundColor: color.RGBA{R: 246, G: 219, B: 158, A: 255},
		cellBackgroundColor: color.RGBA{R: 225, G: 225, B: 211, A: 255},
		wordColor:           color.Black,
		labelColor:          color.RGBA{R: 200, G: 10, B: 10, A: 255},
		borderWidth:         20,
	}
	for _, v := range opts {
		v(opt)
	}
	return opt
}

type renderOpts struct {
	borderWidth         int
	backgroundColor     color.Color
	wordBackgroundColor color.Color
	cellBackgroundColor color.Color
	wordColor           color.Color
	labelColor          color.Color
}

type RenderOption func(opts *renderOpts)

func WithBorder(width int) RenderOption {
	return func(opts *renderOpts) {
		opts.borderWidth = width
	}
}

func WithBackgroundColor(cl color.Color) RenderOption {
	return func(opts *renderOpts) {
		opts.backgroundColor = cl
	}
}

func WithWordBackgroundColor(cl color.Color) RenderOption {
	return func(opts *renderOpts) {
		opts.wordBackgroundColor = cl
	}
}

func WithCellBackgroundColor(cl color.Color) RenderOption {
	return func(opts *renderOpts) {
		opts.cellBackgroundColor = cl
	}
}

func WithWordColor(cl color.Color) RenderOption {
	return func(opts *renderOpts) {
		opts.wordColor = cl
	}
}

func WithLabelColor(cl color.Color) RenderOption {
	return func(opts *renderOpts) {
		opts.labelColor = cl
	}
}

func RenderClassicPNG(c *Classic, width, height int, opts ...RenderOption) (*gg.Context, error) {
	options := resolveRenderOptions(opts...)

	gridWidth := height - options.borderWidth
	gridHeight := height - options.borderWidth

	cellWidth := float64(gridWidth / len(c.Board))
	cellHeight := float64(gridHeight / len(c.Board))
	cellOffset := 0.0
	if options.borderWidth > 0 {
		cellOffset = float64(options.borderWidth) / 2
	}

	dc := gg.NewContext(width, height)
	dc.SetColor(options.backgroundColor)
	dc.Clear()

	// board

	for gridY := 0; gridY < len(c.Board); gridY++ {
		for gridX, cell := range c.Board[gridY] {

			// draw cell with border
			var cellColor color.Color = options.cellBackgroundColor
			if cell.Bonus != NoBonusType {
				switch cell.Bonus {
				case DoubleLetterScoreType:
					cellColor = color.RGBA{R: 183, G: 215, B: 230, A: 255}
				case DoubleWordScoreType:
					cellColor = color.RGBA{R: 216, G: 143, B: 139, A: 255}
				case TripleLetterScoreType:
					cellColor = color.RGBA{R: 84, G: 164, B: 198, A: 255}
				case TripleWordScoreType:
					cellColor = color.RGBA{R: 208, G: 44, B: 32, A: 255}
				}
			}
			dc.SetColor(cellColor)
			dc.DrawRectangle(cellOffset+(float64(gridX)*cellWidth), cellOffset+(float64(gridY)*cellHeight), cellWidth, cellHeight)
			dc.FillPreserve()

			dc.SetColor(options.wordColor)
			dc.SetLineWidth(0.3)
			dc.Stroke()

			if !cell.Empty() {
				dc.DrawRectangle(cellOffset+(float64(gridX)*cellWidth), cellOffset+(float64(gridY)*cellHeight), cellWidth, cellHeight)
				dc.SetColor(options.wordBackgroundColor)
				dc.FillPreserve()

				// draw the word
				dc.SetColor(options.wordColor)
				dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 24}))
				dc.DrawStringAnchored(
					strings.ToUpper(cell.String()),
					cellOffset+float64(gridX)*cellWidth+cellWidth/2,
					cellOffset+float64(gridY)*cellHeight+cellHeight/2,
					0.5,
					0.5,
				)

				// draw letter score
				dc.SetColor(options.wordColor)
				dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 12}))
				dc.DrawStringAnchored(
					cell.LetterScoreString(),
					cellOffset+float64(gridX)*cellWidth+cellWidth-12,
					cellOffset+float64(gridY)*cellHeight+cellHeight-12,
					0.5,
					0.5,
				)

				dc.Stroke()
			}

			// draw cell index
			dc.SetColor(color.RGBA{107, 107, 99, 255})
			dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 12}))
			dc.DrawStringAnchored(
				cell.IndexString(),
				cellOffset+float64(gridX)*cellWidth+12,
				cellOffset+float64(gridY)*cellHeight+12,
				0.5,
				0.5,
			)

		}
	}

	// game information
	dc.SetColor(color.Black)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 20}))
	dc.DrawString(
		"LEGEND",
		float64(gridWidth)+float64(options.borderWidth),
		20+float64(options.borderWidth)/2,
	)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 18}))

	dc.SetColor(color.RGBA{R: 208, G: 44, B: 32, A: 255})
	dc.DrawString(
		"Triple Word Score",
		float64(gridWidth)+float64(options.borderWidth),
		50+float64(options.borderWidth)/2,
	)
	dc.SetColor(color.RGBA{R: 216, G: 143, B: 139, A: 255})
	dc.DrawString(
		"Double Word Score",
		float64(gridWidth)+float64(options.borderWidth),
		70+float64(options.borderWidth)/2,
	)
	dc.SetColor(color.RGBA{R: 84, G: 164, B: 198, A: 255})
	dc.DrawString(
		"Triple Letter Score",
		float64(gridWidth)+float64(options.borderWidth),
		90+float64(options.borderWidth)/2,
	)
	dc.SetColor(color.RGBA{R: 183, G: 215, B: 230, A: 255})
	dc.DrawString(
		"Double Letter Score",
		float64(gridWidth)+float64(options.borderWidth),
		110+float64(options.borderWidth)/2,
	)

	dc.SetColor(color.Black)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 20}))
	dc.DrawString(
		fmt.Sprintf("TILES LEFT: %d", len(c.SpareLetters)),
		float64(gridWidth)+float64(options.borderWidth),
		150+float64(options.borderWidth)/2,
	)

	//scores
	dc.SetColor(color.Black)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 20}))
	dc.DrawString(
		"PLAYER SCORES",
		float64(gridWidth)+float64(options.borderWidth),
		180+float64(options.borderWidth)/2,
	)

	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 18}))
	for i, p := range c.Players {
		suffix := ""
		if c.getCurrentPlayerName() == p.Name {
			dc.SetColor(colornames.Darkblue)
			suffix = " [current player]"
		} else {
			dc.SetColor(colornames.Black)
		}
		dc.DrawString(
			fmt.Sprintf("%s: %d%s", p.Name, p.Score, suffix),
			float64(gridWidth)+float64(options.borderWidth),
			190+float64(options.borderWidth)/2+(25*float64(i+1)),
		)
	}

	return dc, nil
}

func RenderScrabulousPNG(c *Scrabulous, width, height int, opts ...RenderOption) (*gg.Context, error) {
	options := resolveRenderOptions(opts...)

	gridWidth := height - options.borderWidth
	gridHeight := height - options.borderWidth

	cellWidth := float64(gridWidth / len(c.Board))
	cellHeight := float64(gridHeight / len(c.Board))
	cellOffset := 0.0
	if options.borderWidth > 0 {
		cellOffset = float64(options.borderWidth) / 2
	}

	dc := gg.NewContext(width, height)
	dc.SetColor(options.backgroundColor)
	dc.Clear()

	pendingWord := map[int]Cell{}
	if best := c.BestPendingWord(); best != nil {
		for _, c := range best.Result.Cells {
			pendingWord[c.Index] = c
		}
	}

	// board

	for gridY := 0; gridY < len(c.Board); gridY++ {
		for gridX, cell := range c.Board[gridY] {

			// draw cell with border
			var cellColor color.Color = options.cellBackgroundColor
			if cell.Bonus != NoBonusType {
				switch cell.Bonus {
				case DoubleLetterScoreType:
					cellColor = color.RGBA{R: 183, G: 215, B: 230, A: 255}
				case DoubleWordScoreType:
					cellColor = color.RGBA{R: 216, G: 143, B: 139, A: 255}
				case TripleLetterScoreType:
					cellColor = color.RGBA{R: 84, G: 164, B: 198, A: 255}
				case TripleWordScoreType:
					cellColor = color.RGBA{R: 208, G: 44, B: 32, A: 255}
				}
			}
			dc.SetColor(cellColor)
			dc.DrawRectangle(cellOffset+(float64(gridX)*cellWidth), cellOffset+(float64(gridY)*cellHeight), cellWidth, cellHeight)
			dc.FillPreserve()

			dc.SetColor(options.wordColor)
			dc.SetLineWidth(0.3)
			dc.Stroke()

			pendingCell, pending := pendingWord[cell.Index]

			if !cell.Empty() || (pending) {
				dc.DrawRectangle(cellOffset+(float64(gridX)*cellWidth), cellOffset+(float64(gridY)*cellHeight), cellWidth, cellHeight)
				dc.SetColor(options.wordBackgroundColor)
				dc.FillPreserve()

				cellContent := cell.String()
				if pending {
					cellContent = pendingCell.String()
					dc.SetColor(colornames.Green)
				} else {
					dc.SetColor(options.wordColor)
				}

				// draw the word
				dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 24}))
				dc.DrawStringAnchored(
					strings.ToUpper(cellContent),
					cellOffset+float64(gridX)*cellWidth+cellWidth/2,
					cellOffset+float64(gridY)*cellHeight+cellHeight/2,
					0.5,
					0.5,
				)

				cellScore := cell.LetterScoreString()
				if pending {
					cellScore = pendingCell.LetterScoreString()
				}

				// draw letter score
				dc.SetColor(options.wordColor)
				dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 12}))
				dc.DrawStringAnchored(
					cellScore,
					cellOffset+float64(gridX)*cellWidth+cellWidth-12,
					cellOffset+float64(gridY)*cellHeight+cellHeight-12,
					0.5,
					0.5,
				)

				dc.Stroke()
			}

			// draw cell index
			dc.SetColor(color.RGBA{107, 107, 99, 255})
			dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 12}))
			dc.DrawStringAnchored(
				cell.IndexString(),
				cellOffset+float64(gridX)*cellWidth+12,
				cellOffset+float64(gridY)*cellHeight+12,
				0.5,
				0.5,
			)

		}
	}

	suffix := "[IDLE]"
	if c.GameState == StateStealing && c.PlaceWordAt != nil {
		suffix = fmt.Sprintf("[COUNTDOWN %s]", time.Until(*c.PlaceWordAt).Truncate(time.Second))
	}

	xOffset := float64(gridWidth) + float64(options.borderWidth)

	dc.SetColor(color.Black)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 20}))
	dc.DrawString(
		fmt.Sprintf("LETTERS (%d spare) | %s", len(c.SpareLetters), suffix),
		xOffset,
		50,
	)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 18}))
	dc.SetColor(colornames.Black)
	for i, v := range c.Letters {
		dc.SetColor(options.wordBackgroundColor)
		dc.DrawRectangle(xOffset+float64(60*i), 50+float64(options.borderWidth)/2, 55, 55)
		dc.Fill()

		dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 18}))
		dc.SetColor(colornames.Black)
		dc.DrawStringAnchored(
			string(v),
			xOffset+float64(60*i)+30,
			50+float64(options.borderWidth)/2+30,
			0.5,
			0.5,
		)

		dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 10}))
		dc.SetColor(colornames.Black)
		dc.DrawStringAnchored(
			fmt.Sprintf("%d", LetterScores[v]),
			xOffset+float64(60*i)+45,
			50+float64(options.borderWidth)/2+45,
			0.5,
			0.5,
		)
	}

	//scores
	dc.SetColor(color.Black)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 20}))
	dc.DrawString(
		"PLAYER SCORES",
		float64(gridWidth)+float64(options.borderWidth),
		150+float64(options.borderWidth)/2,
	)

	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 18}))
	dc.SetColor(colornames.Black)
	for i, score := range c.GetScores() {
		if !c.IsPlayerAllowed(score.PlayerName) {
			dc.SetColor(colornames.Red)
		} else {
			dc.SetColor(colornames.Black)
		}
		c.PendingWords
		dc.DrawString(
			fmt.Sprintf("%s: %d (%d words)", score.PlayerName, score.Score, score.Words),
			float64(gridWidth)+float64(options.borderWidth),
			150+float64(options.borderWidth)/2+(30*float64(i+1)),
		)
	}

	// tile legend
	dc.SetColor(color.Black)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 20}))
	dc.DrawString(
		"LEGEND",
		float64(gridWidth)+float64(options.borderWidth),
		(float64(gridHeight)-80)+float64(options.borderWidth)/2,
	)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 18}))
	for i, legend := range []struct {
		name   string
		colour color.Color
	}{
		{name: "Triple Word Score", colour: color.RGBA{R: 208, G: 44, B: 32, A: 255}},
		{name: "Double Word Score", colour: color.RGBA{R: 216, G: 143, B: 139, A: 255}},
		{name: "Triple Letter Score", colour: color.RGBA{R: 84, G: 164, B: 198, A: 255}},
		{name: "Double Letter Score", colour: color.RGBA{R: 183, G: 215, B: 230, A: 255}},
	} {
		dc.SetColor(legend.colour)
		dc.DrawString(
			legend.name,
			float64(gridWidth)+float64(options.borderWidth),
			(float64(gridHeight)-20*float64(i))+float64(options.borderWidth)/2,
		)
	}

	return dc, nil
}
