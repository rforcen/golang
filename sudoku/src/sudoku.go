package sudoku

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"

	// "runtime"
	"sync"
	// "time"
)

// globals
var __globalFinished bool = false
var mutex sync.Mutex

// level
type Level int

const (
	lvl_dont_touch Level = iota
	lvl_very_easy
	lvl_easy
	lvl_medium
	lvl_difficult
	lvl_master
)
const (
	emptyCell = 0
)

func PrevLevel(l Level) Level {
	if l == lvl_very_easy {
		return l
	} else {
		return l - 1
	}
}

func NextLevel(l Level) Level {
	if l == lvl_master {
		return l
	} else {
		return l + 1
	}
}

var invalid_coord = Coord{-1, -1}

type Coord struct {
	row int
	col int
}

type Sudoku struct {
	n               int
	szBox           int
	board           [][]int8
	solution        [][]int8
	lookUp          [][][]Coord
	bs              Bitset
	found_solutions int
	max_solutions   int
}

func NewSudoku(n int) *Sudoku { // size of box 3,4,5...
	nxn := n * n
	s := &Sudoku{
		n:               nxn,
		szBox:           n,
		board:           make([][]int8, nxn), // 2d board & solution
		solution:        make([][]int8, nxn),
		bs:              newBitset(nxn),
		found_solutions: 0,
		max_solutions:   1,
	}
	for i := range nxn {
		s.board[i] = make([]int8, nxn)
		s.solution[i] = make([]int8, nxn)
	}
	s.genLookUp()

	return s
}

func (s *Sudoku) ResetBoard() {
	for i := range s.board {
		for j := range s.board[i] {
			s.board[i][j] = emptyCell
		}
	}
}

func (s *Sudoku) getSymbol(row, col int) string {
	const str = " 123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	b := s.board[row][col]
	if b < 0 || b >= int8(len(str)) {
		return ""
	}
	return string(str[b])
}

func (s *Sudoku) genLookUp() { // generate UPPER lookUp vector per cell including box
	n := s.n

	s.lookUp = make([][][]Coord, n) // init lookUp
	for i := range s.lookUp {
		s.lookUp[i] = make([][]Coord, n)
	}

	for row := range n {
		for col := range n {
			curr_coord := Coord{row, col}
			lkp := &s.lookUp[row][col]

			for r := range n {
				*lkp = append(*lkp, Coord{r, col}) // upper rows
			}
			for c := range n {
				*lkp = append(*lkp, Coord{row, c}) // & cols ONLY not lower ones
			}

			rb := s.szBox * (row / s.szBox) // box
			cb := s.szBox * (col / s.szBox)

			for r := range s.szBox {
				for c := range s.szBox {
					coord := Coord{r + rb, c + cb}

					if coord != curr_coord {
						*lkp = append(*lkp, coord)
					}
				}
			}

			unique := func(lkp *[]Coord) []Coord { //
				res := []Coord{}
				contains := func(v Coord) bool {
					for _, vc := range res {
						if vc == v {
							return true
						}
					}
					return false
				}
				for _, v := range *lkp {
					if !contains(v) {
						res = append(res, v)
					}
				}
				return res
			}

			*lkp = unique(lkp)
		}
	}
}

func (s *Sudoku) copyBoard2Sol() {
	for i := range s.solution {
		copy(s.solution[i], s.board[i])
	}
}

func (s *Sudoku) copySol2Board() {
	for i := range s.board {
		copy(s.board[i], s.solution[i])
	}
}

func (s Sudoku) find1stEmpty() Coord {
	for r := range s.n {
		for c := range s.n {
			if s.board[r][c] == emptyCell {
				return Coord{r, c}
			}
		}
	}
	return invalid_coord
}

func (s *Sudoku) saveSolution() {
	s.found_solutions++
	// s.print()
	s.copyBoard2Sol()
}

func (s *Sudoku) move(coord Coord, val int) {
	s.board[coord.row][coord.col] = int8(val)
}

func (s *Sudoku) genMoves(coord Coord) Bitset {
	s.bs.clear()

	for _, l := range s.lookUp[coord.row][coord.col] {
		v := s.board[l.row][l.col] // range 1..n
		if v != emptyCell {
			s.bs.set(int(v - 1))
		}
	}

	return s.bs
}

func (s *Sudoku) clone() *Sudoku {
	sdkCloned := &Sudoku{
		n:               s.n,
		szBox:           s.szBox,
		board:           make([][]int8, s.n), // new board instance
		solution:        s.solution,          // all clones share solution
		bs:              s.bs,
		found_solutions: s.found_solutions,
		max_solutions:   s.max_solutions,
		lookUp:          s.lookUp, // share lookup
	}
	for i := range s.n { // new board copy of s.board
		sdkCloned.board[i] = make([]int8, s.n)
		copy(sdkCloned.board[i], s.board[i])
	}
	return sdkCloned
}

func (s *Sudoku) scan(coord Coord) {
	if coord == invalid_coord {
		s.saveSolution()
	} else {
		if s.found_solutions < s.max_solutions {
			if s.board[coord.row][coord.col] == emptyCell { // skip non empty cells (solve process)

				bs := s.genMoves(coord)
				for mv := bs.next(); mv != -1; mv = bs.next() {
					s.move(coord, mv+1)

					s.scan(s.nextCoord(coord))

					s.move(coord, emptyCell) // b(r,c)=0
				}
			} else { // next cell
				s.scan(s.nextCoord(coord))
			}
		}
	}
}
func (s *Sudoku) scanMT(coord Coord, ngo, maxGo int) {
	if coord == invalid_coord {
		mutex.Lock()
		if !__globalFinished {
			__globalFinished = true
			s.saveSolution()
		}
		mutex.Unlock()

	} else {
		if s.found_solutions < s.max_solutions {
			if s.board[coord.row][coord.col] == emptyCell { // skip non empty cells (solve process)

				bs := s.genMoves(coord)
				for mv := bs.next(); mv != -1; mv = bs.next() {
					if __globalFinished {
						return
					}
					s.move(coord, mv+1)

					if ngo >= maxGo {
						s.scanMT(s.nextCoord(coord), ngo, maxGo)
					} else {
						go s.clone().scanMT(s.nextCoord(coord), ngo+1, maxGo)
					}

					s.move(coord, emptyCell) // b(r,c)=0

				}
			} else { // next cell
				if ngo >= maxGo {
					s.scanMT(s.nextCoord(coord), ngo, maxGo)
				} else {
					go s.clone().scanMT(s.nextCoord(coord), ngo+1, maxGo)
				}
			}
		}
	}
}

func (s Sudoku) IsValid() bool {
	cmpIntArray := func(a, b []int) int {
		for i := range a {
			if a[i] != b[i] {
				return a[i] - b[i]
			}
		}
		return 0
	}

	seq1n := []int{}
	for i := range s.n {
		seq1n = append(seq1n, i+1)
	}

	for rc := range 2 { // rows & cols
		for r := range s.n {
			vb := make([]int, s.n)

			for c := range s.n {
				b := int8(0)
				if rc == 0 {
					b = s.board[r][c]
				} else {
					b = s.board[c][r]
				}
				if b != 0 {
					vb[b-1] = int(b)
				} else {
					return false
				}
			}

			if cmpIntArray(vb, seq1n) != 0 {
				return false
			}
		}
	}

	for row := 0; row < s.n; row += s.szBox {
		for col := 0; col < s.n; col += s.szBox {
			vb := make([]int, s.n)
			for r := range s.szBox { // box
				for c := range s.szBox {
					b := s.board[r+row][c+col]
					if b != 0 {
						vb[b-1] = int(b)
					} else {
						return false
					}
				}
			}
			if cmpIntArray(vb, seq1n) != 0 {
				return false
			}
		}
	}
	return true
}

func (s *Sudoku) Solve() {
	__globalFinished = false

	coord := s.find1stEmpty()

	if coord != invalid_coord {
		s.found_solutions = 0
		s.scan(coord)
		s.copySol2Board()
	}
}
func (s *Sudoku) SolveMT() {
	__globalFinished = false

	coord := s.find1stEmpty()

	if coord != invalid_coord {
		s.found_solutions = 0
		s.scanMT(coord, 0, runtime.NumCPU())

		for !__globalFinished {
			time.Sleep(time.Second)
		}
		s.copySol2Board()
	}
}

// shuffle board

func (s *Sudoku) shuffleBoard() {
	// helpers
	makeSeq := func(n int) []int { // create a iota array
		seq := make([]int, n)
		for i := range n {
			seq[i] = i
		}
		return seq
	}
	shuffle := func(seq []int) {
		for i := range len(seq) {
			j := rand.IntN(i + 1)
			seq[i], seq[j] = seq[j], seq[i]
		}
	}
	swapCol := func(c0 int, c1 int) {
		for c := range s.n {
			s.board[c][c0], s.board[c][c1] = s.board[c][c1], s.board[c][c0]
		}
	}
	swapRow := func(r0 int, r1 int) {
		for r := range s.n {
			s.board[r0][r], s.board[r1][r] = s.board[r1][r], s.board[r0][r]
		}
	}
	//

	szb := s.szBox
	n := s.n

	s0 := makeSeq(szb)
	s1 := makeSeq(szb)

	for b := range n {
		r, c := b/szb, b%szb

		for range n*n + 1 {
			shuffle(s0)
			shuffle(s1)

			for i := range s0 {
				x0, x1 := s0[i], s1[i]
				rs, cs := r*szb, c*szb
				swapCol(szb-1+cs, cs)
				swapRow(szb-1+rs, rs)
				swapCol(x0+cs, x1+cs)
				swapRow(x0+rs, x1+rs)
			}
		}
	}
}

func (s *Sudoku) GenProblem(level Level) {
	// init board
	s.ResetBoard()
	s.Solve()

	s.shuffleBoard()

	// the higher the level the more empty cells 0:n/3, 1:n/2, 2:2*n/3
	for range s.n * s.n / (1 + int(lvl_master) - int(level)) {
		s.board[rand.IntN(s.n)][rand.IntN(s.n)] = emptyCell
	}
}
func (s *Sudoku) GenProblemMT(level Level) {
	// init board
	s.ResetBoard()
	s.SolveMT()

	s.shuffleBoard()

	// the higher the level the more empty cells 0:n/3, 1:n/2, 2:2*n/3
	for range s.n * s.n / (1 + int(lvl_master) - int(level)) {
		s.board[rand.IntN(s.n)][rand.IntN(s.n)] = emptyCell
	}
}

func (s *Sudoku) print() {
	drawLine := func() {
		for range s.n + s.szBox + 1 {
			fmt.Print("-")
		}
		fmt.Println()
	}

	drawLine()
	for r := range s.n {
		for c := range s.n {
			if c%s.szBox == 0 {
				fmt.Print("|")
			}
			fmt.Printf("%s", s.getSymbol(r, c))
		}
		fmt.Print("|")
		fmt.Println()
		if r%s.szBox == s.szBox-1 {
			drawLine()
		}
	}
}

func (s *Sudoku) nextCoord(coord Coord) Coord {
	if coord.col < s.n-1 {
		return Coord{coord.row, coord.col + 1}
	} else {
		if coord.row < s.n-1 {
			return Coord{coord.row + 1, 0}
		} else {
			return invalid_coord
		}
	}
}
