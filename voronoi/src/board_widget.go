package voronoi

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type BoardWidget struct {
	widget.BaseWidget

	content  *fyne.Container
	lastSize fyne.Size
	onResize func(fyne.Size)

	Voronoi *Voronoi
}

func NewBoardWidget(Voronoi *Voronoi) *BoardWidget {
	rw := &BoardWidget{
		lastSize: fyne.NewSize(0, 0),
		content:  container.NewWithoutLayout(), // manual position set
		Voronoi:  Voronoi,
	}
	rw.ExtendBaseWidget(rw)
	return rw
}

func (r *BoardWidget) SetOnResize(f func(fyne.Size)) { r.onResize = f }
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

	v := r.widget.Voronoi
	img := canvas.NewImageFromImage(v.GenerateImage())
	img.FillMode = canvas.ImageFillOriginal
	img.Resize(size)

	r.widget.content.Resize(size)
	r.widget.content.RemoveAll()
	r.widget.content.Add(img)
}

func (r *BoardWidgetRenderer) MinSize() fyne.Size           { return fyne.NewSize(200, 200) } // Minimum size
func (r *BoardWidgetRenderer) Refresh()                     { r.widget.content.Refresh() }
func (r *BoardWidgetRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *BoardWidgetRenderer) Destroy()                     {}
