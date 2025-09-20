package fourinline

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
)

var LEVEL Index = 9

const (
	N     = 7
	N_COL = N
	N_ROW = N - 1

	EVAL_DRAW = -2 // worst than 0 (play) and better than MIN_EVAL (loose)
	MAX_EVAL  = math.MaxInt8
	MIN_EVAL  = -math.MaxInt8
)

type Chip byte
type Index int8

const (
	Empty Chip = iota
	Human
	Machine
)

type Board struct {
	board    [N_ROW][N_COL]Chip
	cols_sum [N_COL]Index
}

type Move struct {
	col  Index
	res  Index
	chip Chip
}

type EvalMove struct {
	eval Index
	move Move
}

type Coord struct {
	row, col Index
}

type Fourinline struct {
	board      Board
	bestMove  Move
	winCoords [][4]Coord
}

// Board
func NewBoard() Board {
	return Board{
		board:    [N_ROW][N_COL]Chip{},
		cols_sum: [N_COL]Index{},
	}
}

func (b *Board) genMoves() []Index {
	var result []Index

	for i := range N_COL {
		p := Index(i + N_COL/2)
		if p >= N_COL {
			p -= N_COL
		}
		if b.cols_sum[p] < N_ROW {
			result = append(result, p)
		}
	}
	return result
}

func (b *Board) isDraw() bool {
	return len(b.genMoves()) == 0
}

func (b *Board) move(col Index, chip Chip) {
	b.board[b.cols_sum[col]][col] = chip
	b.cols_sum[col]++
}

func (b *Board) moveCheck(col Index, chip Chip) bool {
	result := b.cols_sum[col] < N_ROW && col < N_COL
	if result {
		b.board[b.cols_sum[col]][col] = chip
		b.cols_sum[col]++
	}
	return result
}

func (b *Board) take(col Index) {
	if b.cols_sum[col] > 0 {
		b.cols_sum[col]--
		b.board[b.cols_sum[col]][col] = Empty
	}
}

func (b *Board) print() {
	b2char := func(chip Chip) byte {
		switch chip {
		case Human:
			return 'X'
		case Machine:
			return 'O'
		default:
			return ' '
		}
	}
	fmt.Println("-------------")
	for r := range N_ROW {
		for c := range N_COL {
			fmt.Printf("%c ", b2char(b.board[N_ROW-1-r][c]))
		}
		fmt.Println()
	}
	fmt.Println("-------------")
	fmt.Println("0 1 2 3 4 5 6")
}

func (b *Board) eval2String(ev Index) string {
	switch ev {
	case MAX_EVAL:
		return "i'll win"
	case MIN_EVAL:
		return "you can win"
	default:
		return "play"
	}
}

// move

func newMove() Move {
	return Move{
		col:  0,
		res:  MIN_EVAL,
		chip: Empty,
	}
}

func (mv *Move) setIfBetter(col Index, res Index, chip Chip) {
	if res > mv.res {
		mv.col = col
		mv.res = res
		mv.chip = chip
	}
}

func (mv *Move) set(col Index, res Index, chip Chip) {
	mv.col = col
	mv.res = res
	mv.chip = chip
}

func (mv *Move) clear() {
	mv.col = 0
	mv.res = MIN_EVAL
	mv.chip = Empty
}

// 4inline

func (fil *Fourinline) computerWins() bool { return fil.evaluate(Machine) == MAX_EVAL }
func (fil *Fourinline) humanWins() bool    { return fil.evaluate(Human) == MAX_EVAL }

func (fil *Fourinline) evaluate(chip Chip) Index {
	if fil.isWinner(chip) {
		return MAX_EVAL
	} else {
		return 0
	}
}

func newFourinline() *Fourinline {
	return &Fourinline{
		board:      NewBoard(),
		bestMove:  newMove(),
		winCoords: findAllWinningCoords(),
	}
}

func (fil *Fourinline) move(mv *Move) {
	if mv.chip == Empty { // human wins condition... -> correct move
		moves := fil.board.genMoves()
		if len(moves) != 0 {
			mv.set(moves[0], 0, Machine)
		}
	}

	fil.board.moveCheck(mv.col, mv.chip)
}

func (fil *Fourinline) play(level Index) Index { // single thread
	fil.bestMove.clear()
	result := fil.alphaBeta(level, level, MIN_EVAL, MAX_EVAL, Human)
	fil.move(&fil.bestMove)
	return result
}

func (fil *Fourinline) playEval(level Index, em *EvalMove) Index { // single thread evaluate & best_move
	fil.bestMove.clear()
	result := fil.alphaBeta(level, level, MIN_EVAL, MAX_EVAL, Human)
	fil.move(&fil.bestMove)
	return result
}

func (fil *Fourinline) alphaBeta(level Index, max_level Index, alpha Index, beta Index, who Chip) Index {
	var eval Index
	result := Index(0)

	if level == 0 {
		result = fil.evaluate(who) // eval. terminal node
	} else {
		moves := fil.board.genMoves()

		if len(moves) > 0 {

			switch who {
			case Human: // test all Machine moves
				for _, mv := range moves {
					fil.board.move(mv, Machine)

					if fil.computerWins() {
						eval = MAX_EVAL
						if level == max_level {
							fil.bestMove.set(mv, MAX_EVAL, Machine)
						}
					} else {
						eval = fil.alphaBeta(level-1, max_level, alpha, beta, Machine)
					}

					if eval > alpha {
						alpha = eval
						if level == max_level {
							fil.bestMove.setIfBetter(mv, alpha, Machine)
						}
					}
					fil.board.take(mv)

					if beta <= alpha {
						break
					} // beta prune
				}

				result = alpha

			case Machine: // test all human moves
				for _, mv := range moves {
					fil.board.move(mv, Human)

					if fil.humanWins() {
						eval = -MAX_EVAL
						alpha = -MAX_EVAL
						if level == max_level {
							fil.bestMove.set(mv, MIN_EVAL, Machine)
						}
					} else {
						eval = fil.alphaBeta(level-1, max_level, alpha, beta, Human)
					}

					if eval < beta {
						beta = eval
						if level == max_level {
							fil.bestMove.set(mv, beta, Machine)
						}
					}

					fil.board.take(mv)

					if beta <= alpha {
						break
					} // alpha prune
				}
				result = beta

			case Empty:
			}
			
		} else {
			result = EVAL_DRAW
		}
	}
	return result
}

func (fil *Fourinline) isWinner(chip Chip) bool {
	result := false

	for _, wcs := range fil.winCoords {
		var is_win = true
		for _, c4 := range wcs {
			if fil.board.board[c4.row][c4.col] != chip {
				is_win = false
				break
			}
		}
		if is_win {
			result = true
			break
		}
	}
	return result
}

func findAllWinningCoords() [][4]Coord {

	var result [][4]Coord
	for r := range N_ROW { // rows
		for c := range N_COL - 4 + 1 {
			v := [4]Coord{}
			for i := range 4 {
				v[i] = Coord{Index(r), Index(c + i)}
			}
			result = append(result, v)
		}
	}

	for c := range N_COL { // cols
		for r := range N_ROW - 4 + 1 {
			v := [4]Coord{}
			for i := range 4 {
				v[i] = Coord{Index(r + i), Index(c)}
			}
			result = append(result, v)
		}
	}

	// diag-right & left
	for _, r := range []Index{2, 1, 0} {
		for _, cr := range [][]Index{{0, 1, 2, 3}, {0, 1, 2, 3, 4}, {0, 1, 2, 3, 4, 5}} {
			for _, c := range cr {
				cpr := [4]Coord{}
				cpl := [4]Coord{}
				np := 0

				for p := range Index(4) {
					if r+p >= N_ROW || c+p >= N_COL {
						break
					}
					cpr[p] = Coord{Index(r + p), Index(c + p)}
					cpl[p] = Coord{Index(r + p), Index((N_COL - 1) - (c + p))}
					np++
				}

				if np == 4 {
					result = append(result, cpr)
					result = append(result, cpl)
				}
			}
		}
	}

	return result
}

// cli play
func Test01() {
	b := NewBoard()
	chip := Machine

	for {
		moves := b.genMoves()
		if len(moves) != 0 {
			b.move(moves[rand.Intn(len(moves))], chip)
			if chip == Human {
				chip = Machine
			} else {
				chip = Human
			}
		} else {
			break
		}
		fmt.Println()
		b.print()
	}
}

func Test02() {
	fil:=newFourinline()

	for {
		fil.play(LEVEL)
		fil.board.print()
		fmt.Println()

		endCondition := func() bool {
			if fil.humanWins(){
				fmt.Println("You win")
				return true
			}
			if fil.computerWins(){
				fmt.Println("I win")
				return true
			}		
			if fil.board.isDraw(){
				fmt.Println("Draw")
				return true
			}
			return false
		}

		if endCondition(){
			return
		}
		
		reader := bufio.NewReader(os.Stdin)
		char,_ := reader.ReadByte()
		switch char {
		case 'q':		return
		case '0','1','2','3','4','5','6':	if fil.board.moveCheck(Index(char-'0'), Human){
				fil.board.print()
				if endCondition(){
					return
				}
			}
		}
	}
}