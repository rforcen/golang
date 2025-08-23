// UI for displaying random DC
package dc

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	n          = 1
	w          = 1200 * n
	h          = 1200 * n
	complexity = 6
)

// tap mandel widget
type DCWidget struct {
	widget.BaseWidget // Embed BaseWidget to get default implementations for many widget methods
	img               *canvas.Image
	win               fyne.Window

	dc_        DC
	complexity int
}

func New_DC_Widget(w, h, complexity int, win fyne.Window) *DCWidget {
	dc_ := NewDC(w, h, "") // create mandel & Image
	dc_.Random(complexity)
	fmt.Println(dc_.GetExpression())

	ti := &DCWidget{
		img:        canvas.NewImageFromImage(dc_.GenerateImage()),
		dc_:        dc_,
		complexity: complexity,
		win:        win,
	}

	ti.ExtendBaseWidget(ti)
	ti.img.FillMode = canvas.ImageFillOriginal

	return ti
}

func (ti *DCWidget) update() {
	ti.img.Image = ti.dc_.GenerateImage()
	ti.img.Refresh()

	ti.win.SetTitle(fmt.Sprintf("Domain Coloring %v x %v, complexity: %v, lap: %v ms", w, h, ti.complexity, ti.dc_.Lap))
}

func (ti *DCWidget) save_last() {
	for fn := 0; ; fn++ {
		fname := fmt.Sprintf("dc%v.png", fn)
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			ti.dc_.WritePng(fname)
			break
		}
	}
}

func (ti *DCWidget) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(ti.img) }

func (ti *DCWidget) Tapped(event *fyne.PointEvent) {
	ti.dc_.Random(ti.complexity)
	fmt.Println(ti.dc_.GetExpression())
	ti.update()
}

////////////////////////////////

func UI() {

	myApp := app.New()
	win := myApp.NewWindow("Domain Coloring")
	win.Resize(fyne.NewSize(float32(w), float32(h)))

	canvas := win.Canvas()
	win.SetPadded(false)

	tapImage := New_DC_Widget(w, h, complexity, win)

	canvas.SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyEscape:
			win.Close()
			return

		case fyne.KeySpace: // ramdom dc
			tapImage.dc_.Random(tapImage.complexity)
			fmt.Println(tapImage.dc_.GetExpression())			

		case fyne.KeyPlus:
			tapImage.complexity++

		case fyne.KeyMinus:
			tapImage.complexity--
			if tapImage.complexity < 2 {
				tapImage.complexity = 2
			}
		case fyne.KeyS: // save to next dc#.png file
			tapImage.save_last()

		default:
			log.Printf("Typed key: %s", key.Name)
			return
		}

		tapImage.update()
	})

	win.SetContent(tapImage)
	win.ShowAndRun()
}
