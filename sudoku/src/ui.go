package sudoku

import (
	"fmt"
	"image/png"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	padding   = float32(0)
	boardSize = float32(800 * 2)
)

var nSudoku = 3
var sudoku *Sudoku
var t0 time.Time = time.Now()
var lap float64 = 0.
var mt bool = false
var level Level = lvl_medium

func reCalc() {
	sudoku = NewSudoku(nSudoku)
	t0 = time.Now()
	if nSudoku < 5 {
		if mt {
			sudoku.SolveMT()
		} else {
			sudoku.Solve()
		}
	}
	sudoku.shuffleBoard()

	lap = float64(time.Since(t0)) / 1e6
}

func UI() {
	myApp := app.New()

	win := myApp.NewWindow(fmt.Sprintf("Sudoku %d", nSudoku))
	reCalc()

	reDraw := func() {
		boardWid := NewBoardWidget(sudoku)
		boardWid.Resize(win.Canvas().Size())
		win.SetContent(boardWid)

		win.SetTitle(fmt.Sprintf("sudoku [%d], lap: %.0f ms | %v | [%v], MT:%v", nSudoku, lap, level, sudoku.IsValid(), mt))
	}

	win.Resize(fyne.NewSize(boardSize, boardSize))

	win.Canvas().SetOnTypedKey(func(keyEvent *fyne.KeyEvent) {

		needRecalc := false

		switch keyEvent.Name {
		case fyne.KeyEscape:
			win.Close()
			return

		case fyne.KeyPlus: // n queens +-
			nSudoku++
			needRecalc = true

		case fyne.KeyMinus:
			if nSudoku > 2 {
				nSudoku--
				needRecalc = true
			}

		case fyne.KeyLeft:
			level = PrevLevel(level)

		case fyne.KeyRight:
			level = NextLevel(level)

		case fyne.KeySpace:
			if mt {
				sudoku.GenProblemMT(level)
			} else {
				sudoku.GenProblem(level)
			}

		case fyne.KeyEnter, fyne.KeyReturn:
			t0 = time.Now()
			sudoku.Solve()
			lap = float64(time.Since(t0)) / 1e6

		case fyne.KeyM:
			mt = !mt
			needRecalc = true

		case fyne.KeyP:
			sudoku.print()

		case fyne.KeyC:
			capturedImage := win.Canvas().Capture()

			fileExists := func(filename string) bool {
				info, err := os.Stat(filename)
				if os.IsNotExist(err) {
					return false
				}
				return !info.IsDir()
			}

			for nf := 0; ; nf++ { // find 1st sudoku_#.png file
				if !fileExists(fmt.Sprintf("sudoku_%d.png", nf)) {
					file, _ := os.Create(fmt.Sprintf("sudoku_%d.png", nf))
					defer file.Close()

					png.Encode(file, capturedImage)
					break
				}
			}

		default:
			return
		}

		if needRecalc {
			reCalc()
		}
		reDraw()
	})

	reDraw()
	win.ShowAndRun()
}
