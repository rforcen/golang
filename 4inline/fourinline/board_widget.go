package fourinline

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type BoardWidget struct {
	widget.BaseWidget

	content  *fyne.Container
	lastSize fyne.Size

	fil *Fourinline
	mouseCallBack func(event *desktop.MouseEvent)
}

func NewBoardWidget(fil *Fourinline) *BoardWidget {
	rw := &BoardWidget{
		lastSize: fyne.NewSize(0, 0),
		content:  container.NewWithoutLayout(), // manual position set
		fil:   fil,
	}
	rw.ExtendBaseWidget(rw)
	return rw
}

func (r *BoardWidget) SetMouseCallBack(mouseCallBack func(event *desktop.MouseEvent)) {
	r.mouseCallBack = mouseCallBack
}

func (r *BoardWidget) MouseDown(event *desktop.MouseEvent) {
	if event.Button == desktop.MouseButtonPrimary {
		if r.mouseCallBack != nil {
			
			event.Position.X = float32(math.Trunc(float64(N_COL) * float64(event.Position.X) / float64(r.Size().Width)))
			event.Position.Y = float32(math.Trunc(float64(N_ROW) * float64(event.Position.Y) / float64(r.Size().Height)))
			
			r.mouseCallBack(event)

			r.Resize(fyne.NewSize(r.Size().Width+1, r.Size().Height+1))
			r.Resize(fyne.NewSize(r.Size().Width-1, r.Size().Height-1))
		}
	}
}

func (r *BoardWidget) MouseUp(event *desktop.MouseEvent) {}

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
	boardSize := fyne.Min(size.Width, size.Height)
	gridSize := boardSize - 2*padding
	tileSize := gridSize / float32(N_COL)

	for i := 0; i <= N_ROW; i++ { // lines		
		// Horizontal lines
		yPos := padding + float32(i)*tileSize
		lh := canvas.NewLine(color.White)
		lh.StrokeWidth = 1
		lh.Position1 = fyne.NewPos(padding, yPos)
		lh.Position2 = fyne.NewPos(boardSize-padding, yPos)
		r.widget.content.Add(lh)
	}

	for i:=0 ; i<=N_COL ; i++ {
		// Vertical lines
		xPos := padding + float32(i)*tileSize
		lv := canvas.NewLine(color.White)
		lv.StrokeWidth = 1
		lv.Position1 = fyne.NewPos(xPos, padding)
		lv.Position2 = fyne.NewPos(xPos, boardSize-padding)
		r.widget.content.Add(lv)
	}

	for row := range N_ROW {
		for col := range N_COL {
			var qc *canvas.Circle
			var chip Chip = r.widget.fil.board.board[row][col]

			switch chip {
			case Human:
				qc = canvas.NewCircle(color.RGBA{R: 255, G: 0, B: 0, A: 255})
				qc.StrokeColor = color.RGBA{A: 0}
				qc.FillColor = color.RGBA{R: 255, G: 215, B: 0, A: 255}
			case Machine:
				qc = canvas.NewCircle(color.RGBA{R: 0, G: 0, B: 255, A: 255})
				qc.StrokeColor = color.RGBA{A: 0}
				qc.FillColor = color.RGBA{R: 0, G: 215, B: 255, A: 255}
			case Empty:
				continue
			}
			qc.StrokeWidth = 0
			qc.Resize(fyne.NewSize(tileSize/1.5, tileSize/1.5))

			xPos := padding + float32(col)*tileSize + tileSize/2 - qc.Size().Width/2
			yPos := padding + float32(row)*tileSize + tileSize/2 - qc.Size().Height/2
			qc.Move(fyne.NewPos(xPos, boardSize-yPos-qc.Size().Height))

			r.widget.content.Add(qc)
		}
	}

	r.widget.content.Resize(size)
}

func (r *BoardWidgetRenderer) MinSize() fyne.Size           { return fyne.NewSize(200, 200) } // Minimum size
func (r *BoardWidgetRenderer) Refresh()                     { r.widget.content.Refresh() }
func (r *BoardWidgetRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *BoardWidgetRenderer) Destroy()                     {}
