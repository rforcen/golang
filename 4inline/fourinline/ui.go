package fourinline

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
)

const (
	padding   = float32(0)
	boardSize = float32(800 * 2)
)

var fil *Fourinline
var t0 time.Time = time.Now()
var lap float64 = 0.
var mt bool = false
var level Index = 9
var boardWin *BoardWidget
var win fyne.Window
var endOfGame bool = false

func reDraw() {
	if fil.humanWins() {
		win.SetTitle("You win")
		endOfGame = true
	} else if fil.computerWins() {
		win.SetTitle("I win")
		endOfGame = true
	} else if fil.board.isDraw() {
		win.SetTitle("Draw")
		endOfGame = true
	} else {
		win.SetTitle(fmt.Sprintf("Fourinline, level: %v | [%v], MT:%v, lap:%.2f", level, fil.board.eval2String(fil.evaluate(Human)), mt, lap))
	}	
	
	boardWin.Refresh()	
}

func setBoard() {
	t0 = time.Now()
	fil = newFourinline()
	endOfGame = false
	fil.play(level)
	lap = float64(time.Since(t0)) / 1e6

	boardWin = NewBoardWidget(fil)	
	boardWin.Resize(win.Canvas().Size())	
	boardWin.Refresh()
	win.SetContent(boardWin)	

	win.Resize(fyne.NewSize(boardSize, boardSize))

	reDraw()
	
	boardWin.SetMouseCallBack(func(event *desktop.MouseEvent) {
		if endOfGame {
			return
		}
		if event.Button == desktop.MouseButtonPrimary {	
			col:=int(event.Position.X)
			if fil.board.moveCheck(Index(col), Human) {
				t0 = time.Now()
				fil.play(level)
				lap = float64(time.Since(t0)) / 1e6

				reDraw()
			}
		}
	})
}

func UI() {
	

	myApp := app.New()
	win = myApp.NewWindow("Fourinline")

	setBoard()

	win.Canvas().SetOnTypedKey(func(keyEvent *fyne.KeyEvent) {

		switch keyEvent.Name {
		case fyne.KeyEscape:
			win.Close()
			return

		case fyne.KeyPlus: // level +-
			level++

		case fyne.KeyMinus:
			if level > 1 {
				level--
			}

		case fyne.KeyM:
			mt = !mt

		case fyne.KeySpace:
			
			setBoard()

		case fyne.Key0, fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4, fyne.Key5, fyne.Key6:
			key2int := func(key fyne.KeyName) Index {
				return Index(key[0] - '0')
			}
			if !endOfGame {
				if fil.board.moveCheck(key2int(keyEvent.Name), Human) {
					t0 = time.Now()
					fil.play(level)
					lap = float64(time.Since(t0)) / 1e6
				}
			}
			
		default:
			return
		}

		
		reDraw()
	})

	reDraw()
	win.ShowAndRun()
}
