package queens

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	padding         = float32(0)
	boardSize       = float32(800 * 2)
	nQueensAutoCalc = 28
)

var nQueens, nSol, currSol int = 20, 1, 0
var q *Queens = nil
var board []int8
var t0 time.Time = time.Now()
var recalc bool = true
var lap float64 = 0.
var mt bool = false

func reCalc(win fyne.Window, ns int) {
	if recalc {
		q = NewQueens(nQueens)
		t0 = time.Now()
		nSol = q.FindFirst(ns, mt)

		if currSol >= len(q.solutions) {
			currSol = len(q.solutions) - 1
		}
		board = q.solutions[currSol]

		currSol = 0
		lap = float64(time.Since(t0)) / 1e6
	} else {
		lap = 0.
	}

	win.SetTitle(fmt.Sprintf("n-queens [%d], (%d/%d), mt:%v lap: %.0f ms", nQueens, currSol+1, len(q.solutions), mt, lap))
}

func UI() {
	myApp := app.New()

	win := myApp.NewWindow(fmt.Sprintf("N-Queens %d", nQueens))
	reCalc(win, 1)

	reDraw := func() {
		boardWid := NewBoardWidget(board)
		boardWid.Resize(win.Canvas().Size())
		win.SetContent(boardWid)
	}

	win.Resize(fyne.NewSize(boardSize, boardSize))
	win.Canvas().SetOnTypedKey(func(keyEvent *fyne.KeyEvent) {

		board = q.solutions[currSol]
		recalc = false

		switch keyEvent.Name {
		case fyne.KeyEscape:
			win.Close()
			return

		case fyne.KeyPlus: // n queens +-
			nQueens++
			recalc = nQueens < nQueensAutoCalc

		case fyne.KeyMinus:
			if nQueens > 4 {
				nQueens--
			}
			recalc = nQueens < nQueensAutoCalc
		case fyne.KeyPageUp:
			nQueens += 10
			recalc = nQueens < nQueensAutoCalc

		case fyne.KeyPageDown:
			nQueens -= 10
			if nQueens < 4 {
				nQueens = 4
			}
			recalc = nQueens < nQueensAutoCalc

		case fyne.KeyLeft:
			if currSol > 0 {
				currSol--
			}
			board = q.solutions[currSol]

		case fyne.KeyRight:
			if currSol < len(q.solutions)-1 {
				currSol++
			}
			board = q.solutions[currSol]

		case fyne.KeyUp:
			nSol++
			recalc = true

		case fyne.KeyDown:
			if nSol > 1 {
				nSol--
			}
			recalc = true

		case fyne.KeySpace:
			recalc = true

		case fyne.KeyP:
			fmt.Println(q.n, q.solutions[currSol])

		case fyne.KeyM:
			mt = !mt
			recalc = true

		case fyne.KeyA:
			currSol = 0
			nSol = 0
			board = q.solutions[currSol]
			recalc = true

		// transformations
		case fyne.KeyR:
			board = q.Rotate90(board)
		case fyne.KeyH:
			board = q.MirrorH(board)
		case fyne.KeyV:
			board = q.MirrorV(board)
		case fyne.KeyT:
			board = q.TranslateV(board)
		case fyne.KeyY:
			board = q.TranslateH(board)

		case fyne.KeyZ:
			q.solutions = q.AllTransformations(board)
			board = q.solutions[currSol]
			currSol = 0

		default:
			board = q.solutions[currSol]
		}

		reCalc(win, nSol)
		reDraw()
	})

	reDraw()
	win.ShowAndRun()
}
