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

func RenderPNG(c *Game, width, height int, opts ...RenderOption) (*gg.Context, error) {
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

	//scores
	dc.SetColor(color.Black)
	dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 20}))
	dc.DrawString(
		"PLAYER SCORES",
		float64(gridWidth)+float64(options.borderWidth),
		160+float64(options.borderWidth)/2,
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
			170+float64(options.borderWidth)/2+(25*float64(i+1)),
		)
	}

	return dc, nil
}
