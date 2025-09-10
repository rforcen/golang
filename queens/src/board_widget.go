package queens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

type BoardWidget struct {
	widget.BaseWidget

	content  *fyne.Container
	lastSize fyne.Size
	onResize func(fyne.Size)

	board []int8
}

func NewBoardWidget(board []int8) *BoardWidget {
	rw := &BoardWidget{
		lastSize: fyne.NewSize(0, 0),
		content:  container.NewWithoutLayout(), // manual position set
		board:    board,
	}
	rw.ExtendBaseWidget(rw)
	return rw
}

func (r *BoardWidget) SetBoard(board []int8) {
	r.board = board
	r.Refresh()
}

func (r *BoardWidget) SetOnResize(f func(fyne.Size)) { r.onResize = f }
func (r *BoardWidget) CreateRenderer() fyne.WidgetRenderer { return &BoardWidgetRenderer{widget: r, objects: []fyne.CanvasObject{r.content}} }

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
	tileSize := gridSize / float32(len(r.widget.board))

	nQueens := len(r.widget.board)

	for i := 0; i <= nQueens; i++ { // lines
		yPos := padding + float32(i)*tileSize
		lh := canvas.NewLine(color.White)
		lh.StrokeWidth = 1
		lh.Position1 = fyne.NewPos(padding, yPos)
		lh.Position2 = fyne.NewPos(boardSize-padding, yPos)
		r.widget.content.Add(lh)

		// Vertical lines
		xPos := padding + float32(i)*tileSize
		lv := canvas.NewLine(color.White)
		lv.StrokeWidth = 1
		lv.Position1 = fyne.NewPos(xPos, padding)
		lv.Position2 = fyne.NewPos(xPos, boardSize-padding)
		r.widget.content.Add(lv)
	}

	for i, pos := range r.widget.board {
		row, col := i, pos
		qc := canvas.NewCircle(color.RGBA{R: 255, A: 255})
		qc.StrokeColor = color.RGBA{A: 0}
		qc.FillColor = color.RGBA{R: 255, G: 215, B: 0, A: 255}
		qc.StrokeWidth = 0
		qc.Resize(fyne.NewSize(tileSize/1.5, tileSize/1.5))

		xPos := padding + float32(col)*tileSize + tileSize/2 - qc.Size().Width/2
		yPos := padding + float32(row)*tileSize + tileSize/2 - qc.Size().Height/2
		qc.Move(fyne.NewPos(xPos, boardSize-yPos-qc.Size().Height))

		r.widget.content.Add(qc)
	}

	r.widget.content.Resize(size)
}

func (r *BoardWidgetRenderer) MinSize() fyne.Size           { return fyne.NewSize(200, 200) } // Minimum size
func (r *BoardWidgetRenderer) Refresh()                     { r.widget.content.Refresh() }
func (r *BoardWidgetRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *BoardWidgetRenderer) Destroy()                     {}
