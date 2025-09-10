package voronoi

import (
	"fmt"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	winSize = float32(1024)
)

var vrn *Voronoi = nil
var t0 time.Time = time.Now()
var lap float64 = 0.
var single_thread bool = false

func reCalc(win fyne.Window) {
	w, h := int(win.Canvas().Size().Width), int(win.Canvas().Size().Height)
	t0 = time.Now()

	vrn = NewVoronoi(w, h, w/2, single_thread)

	lap = float64(time.Since(t0)) / 1e6

	thText := func() string {
		if single_thread {
			return "ST"
		}
		return fmt.Sprintf("%d MT", runtime.NumCPU())
	}
	win.SetTitle(fmt.Sprintf("Voronoi %d x %d, %s, lap: %.0f ms", vrn.w, vrn.h, thText(), lap))
}

func UI() {
	myApp := app.New()

	win := myApp.NewWindow("Voronoi")
	win.Resize(fyne.NewSize(winSize, winSize))

	reCalc(win)

	reDraw := func() {
		boardWid := NewBoardWidget(vrn)
		boardWid.Resize(win.Canvas().Size())
		win.SetContent(boardWid)
	}

	win.Canvas().SetOnTypedKey(func(keyEvent *fyne.KeyEvent) {
		switch keyEvent.Name {
		case fyne.KeyEscape:
			win.Close()
			return
		case fyne.KeyM:
			single_thread = !single_thread
		}

		reCalc(win)
		reDraw()
	})

	reDraw()
	win.ShowAndRun()
}
