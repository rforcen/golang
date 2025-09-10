package queens

import (
	"math/big"
	"runtime"
	"sync"
)

type Queens struct {
	board         []int8
	solutions     [][]int8
	n             int8
	nComb         *big.Int
	countEvals    uint64
	ld, rd, cl    []bool // left, right, columns set
	abort         bool
	stopSolutions int
	queens        []*Queens
}

func NewQueens(n int) *Queens {
	nComb := big.NewInt(0)
	nComb.Exp(big.NewInt(int64(n)), big.NewInt(int64(n)), nil)

	q := &Queens{
		board:         make([]int8, n),
		ld:            make([]bool, 2*n),
		rd:            make([]bool, 2*n),
		cl:            make([]bool, 2*n),
		solutions:     make([][]int8, 0),
		n:             int8(n),
		abort:         false,
		nComb:         nComb,
		countEvals:    0,
		stopSolutions: 0,
		queens:        nil,
	}
	return q
}

func (q *Queens) zeroBoard() {
	for i := int8(0); i < q.n; i++ {
		q.set(i, 0)
	}
}
func (q *Queens) Next() bool {
	for i := int8(0); i < q.n; {
		q.board[i]++
		if q.board[i] >= q.n {
			for j := i; j >= 0; j-- { // reset prev
				q.board[j] = 0
			}
			i++
		} else {
			return true
		}
	}
	return false
}
func (q *Queens) isValid() bool {
	absInt8 := func(x int8) int8 {
		if x < 0 {
			return -x
		}
		return x
	}
	for i := int8(0); i < q.n; i++ {
		for j := int8(i + 1); j < q.n; j++ {
			if q.board[i] == q.board[j] {
				return false
			}
			if i-q.board[i] == j-q.board[j] {
				return false
			}
			if absInt8(q.board[i]-q.board[j]) == absInt8(i-j) {
				return false
			}
		}
	}
	return true
}

// convert a lexicographic n value to a board
func (q *Queens) set_n_board(np *big.Int) {
	n := new(big.Int).Set(np)
	q.zeroBoard()

	for i := int8(0); i < q.n; i++ {
		nm := new(big.Int).Set(n)
		q.board[i] = int8(nm.Mod(nm, big.NewInt(int64(q.n))).Int64())
		n.Div(n, big.NewInt(int64(q.n)))
		if n.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
}

func (q *Queens) clone() *Queens {
	qr := *q

	qr.board = make([]int8, q.n)
	qr.ld = make([]bool, 2*q.n)
	qr.rd = make([]bool, 2*q.n)
	qr.cl = make([]bool, 2*q.n)
	qr.solutions = make([][]int8, len(q.solutions))
	qr.queens = make([]*Queens, len(q.queens))

	copy(qr.queens, q.queens)
	copy(qr.board, q.board)
	copy(qr.ld, q.ld)
	copy(qr.rd, q.rd)
	copy(qr.cl, q.cl)
	copy(qr.solutions, q.solutions)

	return &qr
}

// find solution
func (q *Queens) isValidPosition(col int8, i int8) bool {
	return !q.ld[i-col+q.n-1] && !q.rd[i+col] && !q.cl[i]
}
func (q *Queens) set(col int8, i int8) { // move
	q.board[col] = i
	q.ld[i-col+q.n-1], q.rd[i+col], q.cl[i] = true, true, true
}
func (q *Queens) reset(col int8, i int8) { // unmove
	q.board[col] = 0
	q.ld[i-col+q.n-1], q.rd[i+col], q.cl[i] = false, false, false
}

func (q *Queens) saveSolution() {
	if !q.abort && q.isValid() {
		b := make([]int8, q.n) // make a copy and add to solutions
		copy(b, q.board)
		q.solutions = append(q.solutions, b)

		if q.queens != nil {
			sumSols := 0
			for _, qs := range q.queens {
				sumSols += len(qs.solutions)
			}
			if q.stopSolutions != 0 && sumSols >= q.stopSolutions {
				// stop all
				q.abort = true
				for i := range q.queens {
					q.queens[i].abort = true
				}
			}
		} else {
			if q.stopSolutions != 0 && len(q.solutions) >= q.stopSolutions {
				q.abort = true
			}
		}
	}
}

func (q *Queens) scan(col int8) {
	if !q.abort {
		if col >= q.n {
			q.saveSolution()
		} else {
			for i := range q.n {
				if q.isValidPosition(col, i) {
					q.set(col, i)   // move
					q.scan(col + 1) // scan col+1
					q.reset(col, i) // unmove
				}
			}
			q.countEvals += uint64(q.n)
		}
	}
}

func (q *Queens) FindFirst(n_sols int, mt bool) int {
	if mt {
		return q.FindFirstMT(n_sols)
	} else {
		return q.FindFirstST(n_sols)
	}
}
func (q *Queens) FindFirstST(n_sols int) int {
	if n_sols == 0 {
		q.stopSolutions = 0
	} else {
		q.stopSolutions = n_sols
	}

	if q.n > 10 {
		q.set(0, q.n/2) // a solution seems to be closer to n/2
		q.set(1, (q.n/2+2)%q.n)
		q.scan(2)
	} else {
		q.scan(0)
	}

	return len(q.solutions)
}

func (q *Queens) FindFirstMT(n_sols int) int {
	if n_sols == 0 {
		q.stopSolutions = 0
	} else {
		q.stopSolutions = n_sols
	}

	minInt := func(a, b int8) int8 {
		if a < b {
			return a
		}
		return b
	}
	numCores := minInt(q.n, int8(runtime.NumCPU()))

	q.queens = make([]*Queens, numCores) // create queens array
	for i := range int8(numCores) {
		q.queens[i] = NewQueens(int(q.n))
		q.queens[i].stopSolutions = q.stopSolutions

		q.queens[i].set(0, q.n/2) // set cols 0,1 -> scan(2)
		q.queens[i].set(1, (q.n/2+i+2)%q.n)

		q.queens[i].queens = q.queens // all queens should have acces to the same set
	}

	// launch
	var wg sync.WaitGroup
	wg.Add(int(numCores))

	for th := range numCores { // scan for each thread from 2
		go func() {
			q.queens[th].scan(2)
			wg.Done()
		}()
	}

	wg.Wait()

	q.solutions = [][]int8{}
	for i := range q.queens {
		q.solutions = append(q.solutions, q.queens[i].solutions...)
	}

	return len(q.solutions)
}

func cloneBoard(board []int8) []int8 {
	cloned := make([]int8, len(board))
	copy(cloned, board)
	return cloned
}

// transformations
func (q *Queens) Rotate90(board []int8) []int8 {
	rotQueens := make([]int8, q.n)
	for i := range q.n {
		rotQueens[i] = 0
		for j := range q.n {
			if board[j] == i {
				rotQueens[i] = q.n - j - 1
				break
			}
		}
	}
	return rotQueens
}

func (q *Queens) MirrorH(board []int8) []int8 {
	mirrorQueens := make([]int8, q.n)
	for i := range q.n {
		mirrorQueens[i] = q.n - 1 - board[i]
	}
	return mirrorQueens
}
func (q *Queens) MirrorV(board []int8) []int8 {
	mirrorQueens := make([]int8, q.n)
	for i := range q.n / 2 {
		mirrorQueens[i], mirrorQueens[q.n-1-i] = board[q.n-1-i], board[i]
	}
	return mirrorQueens
}
func (q *Queens) TranslateV(board []int8) []int8 {
	translatedQueens := make([]int8, q.n)
	for i := range q.n {
		translatedQueens[i] = (board[i] + 1) % q.n
	}
	return translatedQueens
}
func (q *Queens) TranslateH(board []int8) []int8 {
	v := make([]int8, q.n)
	for i := range q.n - 1 {
		v[i+1] = board[i]
	}
	v[0] = board[q.n-1]
	return v
}

func (q *Queens) AllTransformations(board []int8) [][]int8 {
	sMap := map[uint64][]int8{} // hash(board) -> board
	hash := func(s []int8) uint64 {	// hash function
		const (
			fnv_offset_basis = uint64(0xcbf29ce484222325)
			final_mixer      = uint64(0xff51afd7ed558ccd)
		)
		primes:=[]uint64{uint64(0x9e3779b97f4a7c15),uint64(0xbf58476d1ce4e5b9)}
		h := fnv_offset_basis

		for i := range s {
			h = (h ^ uint64(s[i])) * primes[i&1]
		}
		return (h ^ (h >> 33)) * final_mixer ^ ((h ^ (h >> 33)) * final_mixer >> 33)

	}

	buildSols := func() [][]int8 { // build solution array from map
		sols := make([][]int8, 0, len(sMap))
		for _, v := range sMap {
			sols = append(sols, v)
		}
		return sols
	}

	b := cloneBoard(board) // working board

	for range 2 {
		for range 2 {
			for range 4 {
				for range q.n {
					for range q.n {
						b = q.TranslateV(b)
						sMap[hash(b)] = b
					}
					b = q.TranslateH(b)
					sMap[hash(b)] = b
				}
				b = q.Rotate90(b)
				sMap[hash(b)] = b
			}
			b = q.MirrorH(b)
			sMap[hash(b)] = b
		}
		b = q.MirrorV(b)
		sMap[hash(b)] = b
	}

	return buildSols()
}
