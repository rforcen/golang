package sudoku

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type BoardWidget struct {
	widget.BaseWidget

	content  *fyne.Container
	lastSize fyne.Size

	sudoku *Sudoku
}

func NewBoardWidget(sudoku *Sudoku) *BoardWidget {
	rw := &BoardWidget{
		lastSize: fyne.NewSize(0, 0),
		content:  container.NewWithoutLayout(), // manual position set
		sudoku:   sudoku,
	}
	rw.ExtendBaseWidget(rw)
	return rw
}

func (r *BoardWidget) CreateRenderer() fyne.WidgetRenderer {
	return &BoardWidgetRenderer{widget: r, objects: []fyne.CanvasObject{r.content}}
}

type BoardWidgetRenderer struct {
	widget  *BoardWidget
	objects []fyne.CanvasObject
}

func (r *BoardWidgetRenderer) Layout(size fyne.Size) {
	// This method is called whenever the widget is resized
	if size.Width == r.widget.lastSize.Width && size.Height == r.widget.lastSize.Height {
		return
	}
	r.widget.lastSize = size

	// bg
	boardBackground := canvas.NewRectangle(color.Black)
	boardBackground.Resize(size)
	r.widget.content.Add(boardBackground)

	// lines
	minFloat32 := func(a, b float32) float32 {
		if a < b {
			return a
		}
		return b
	}
	boardSize := minFloat32(size.Width, size.Height)
	gridSize := boardSize - 2*padding
	tileSize := gridSize / float32(len(r.widget.sudoku.board))

	n := r.widget.sudoku.n
	szBox := r.widget.sudoku.szBox

	for i := 0; i <= n; i++ { // lines
		var col color.Color
		if i%szBox == 0 {
			col = color.RGBA{R: 255, G: 0, B: 0, A: 255} // red color
		} else {
			col = color.White // white color
		}
		// Horizontal lines
		yPos := padding + float32(i)*tileSize
		lh := canvas.NewLine(col)
		lh.StrokeWidth = 1
		lh.Position1 = fyne.NewPos(padding, yPos)
		lh.Position2 = fyne.NewPos(boardSize-padding, yPos)
		if i%szBox == 0 {
			lh.StrokeWidth = 3
		}
		r.widget.content.Add(lh)

		// Vertical lines
		xPos := padding + float32(i)*tileSize
		lv := canvas.NewLine(col)
		lv.StrokeWidth = 1
		lv.Position1 = fyne.NewPos(xPos, padding)
		lv.Position2 = fyne.NewPos(xPos, boardSize-padding)
		if i%szBox == 0 {
			lv.StrokeWidth = 3
		}
		r.widget.content.Add(lv)
	}

	for row := range n {
		for col := range n {
			if r.widget.sudoku.board[row][col] == 0 {
				continue
			}
			st := canvas.NewText(r.widget.sudoku.getSymbol(row, col), color.White)
			st.TextSize = tileSize / 2
			st.Move(fyne.NewPos(padding+float32(col)*tileSize+tileSize/3., padding+float32(row)*tileSize+tileSize/8))
			r.widget.content.Add(st)
		}
	}

	r.widget.content.Resize(size)
}

func (r *BoardWidgetRenderer) MinSize() fyne.Size           { return fyne.NewSize(200, 200) } // Minimum size
func (r *BoardWidgetRenderer) Refresh()                     { r.widget.content.Refresh() }
func (r *BoardWidgetRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *BoardWidgetRenderer) Destroy()                     {}
