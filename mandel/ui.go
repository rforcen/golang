package main

import (
	// "fmt"
	"fmt"
	"log"
	"math"
	"os"

	"mandel/mandel"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	n      = 1
	w      = 1024 * n
	h      = 1024 * n
	Center = complex(0.5, 0.0)
	Range  = complex(-2.0, 2.0)
	Iters  = 200

	offset = 0.03
)

// tap mandel widget
type MandelWidget struct {
	widget.BaseWidget // Embed BaseWidget to get default implementations for many widget methods
	img               *canvas.Image
	win               fyne.Window

	Mnd                mandel.Mandel
	Center, CenterInit complex128
	Range, RangeInit   complex128
	Iters, ItersInit   int
}

func NewTapImage(w, h, iters int, Center, Range complex128, win fyne.Window) *MandelWidget {
	mnd := mandel.NewMandel(w, h, iters, Center, Range) // create mandel & Image
	mnd.GenImage()

	ti := &MandelWidget{
		img:        canvas.NewImageFromImage(mnd.GenerateImage()),
		Mnd:        mnd,
		Center:     Center,
		Range:      Range,
		Iters:      iters,
		CenterInit: Center,
		RangeInit:  Range,
		ItersInit:  iters,
		win:        win,
	}

	ti.ExtendBaseWidget(ti)
	ti.img.FillMode = canvas.ImageFillOriginal

	return ti
}

func (ti *MandelWidget) update() {
	ti.Mnd.Iters, ti.Mnd.Center, ti.Mnd.Range = ti.Iters, ti.Center, ti.Range

	ti.Mnd.Update()
	ti.Mnd.GenImage()

	ti.img.Image = ti.Mnd.GenerateImage()
	ti.img.Refresh()

	ti.win.SetTitle(fmt.Sprintf("Mandelbrot fractal %v x %v, iters: %v, lap: %v ms, range: %.2g", w, h, ti.Iters, ti.Mnd.Lap, math.Abs(real(ti.Range))))
}

func (ti *MandelWidget) reset() {
	ti.Center, ti.Range, ti.Iters = ti.CenterInit, ti.RangeInit, ti.ItersInit

	ti.update()
}

func (ti *MandelWidget) save_last() {
	for fn := 0; ; fn++ {
		fname := fmt.Sprintf("mandel%v.png", fn)
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			ti.Mnd.WritePng(fname)
			break
		}
	}
}

func (ti *MandelWidget) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(ti.img) }

func (ti *MandelWidget) Tapped(event *fyne.PointEvent) {
	scale := ti.win.Canvas().Scale() // scale to tpi value

	x, y := scale * event.Position.X, scale * event.Position.Y

	ti.Center, ti.Range = ti.Mnd.Recalculate(float64(x), float64(y))
	ti.update()


	//log.Printf("win size: %v x %v", ti.win.Canvas().Size().Width, ti.win.Canvas().Size().Height)
}

func (ti *MandelWidget) TappedSecondary(event *fyne.PointEvent) {}
// func (ti *MandelWidget) Resize(size fyne.Size) {
	
// 	ti.BaseWidget.Resize(size)
// 	ti.Mnd = mandel.NewMandel(int(size.Width), int(size.Height), ti.Iters, ti.Center, ti.Range)
// 	ti.update()
// }


////////////////////////////////

func main() {

	myApp := app.New()
	win := myApp.NewWindow("Mandelbrot fractal")
	win.Resize(fyne.NewSize(float32(w), float32(h)))

	canvas := win.Canvas()
	// win.SetPadded(false)

	tapImage := NewTapImage(w, h, Iters, Center, Range, win)

	canvas.SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyEscape:
			win.Close()
			return

		case fyne.KeySpace:
			tapImage.reset()

		case fyne.KeyPlus:
			tapImage.Iters *= 2
		case fyne.KeyMinus:
			tapImage.Iters /= 2
			if tapImage.Iters < 2 {
				tapImage.Iters = 2
			}
		case fyne.KeyLeft:
			tapImage.Center -= complex(offset*math.Abs(real(tapImage.Range)), 0.0)
		case fyne.KeyRight:
			tapImage.Center += complex(offset*math.Abs(real(tapImage.Range)), 0.0)
		case fyne.KeyUp:
			tapImage.Center += complex(0.0, offset*math.Abs(imag(tapImage.Range)))
		case fyne.KeyDown:
			tapImage.Center -= complex(0.0, offset*math.Abs(imag(tapImage.Range)))

		case fyne.KeyPageUp:
			tapImage.Range = complex(real(tapImage.Range)*2, imag(tapImage.Range)*2)
		case fyne.KeyPageDown:
			tapImage.Range = complex(real(tapImage.Range)/2, imag(tapImage.Range)/2)

		case fyne.KeyS:
			tapImage.save_last()

		default:
			log.Printf("Typed key: %s", key.Name)
		}

		tapImage.update()
	})

	
	win.SetContent(tapImage)
	win.ShowAndRun()
}
