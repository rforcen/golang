package queens

import (
	"fmt"
	"time"
	"math/big"
)

func Test_traverse_all_boards() { // using big.Int
	nq := 8
	q := NewQueens(nq)

	nSols := big.NewInt(0)

	for i := big.NewInt(0); i.Cmp(q.nComb) != 0; i.Add(i, big.NewInt(1)) {
		q.set_n_board(i)
		if q.isValid() {
			nSols.Add(nSols, big.NewInt(1))
			fi := float64(i.Int64())
			ft := float64(q.nComb.Int64())
			fmt.Printf("%.0f %8s: %v \n", 100.*fi/ft, nSols.String(), q.board)
		}
	}
	fmt.Println(q.nComb.String())
}

func Test_Next() {
	nq := 12
	q := NewQueens(nq)

	percInit := 3

	initPos:=big.NewInt(int64(percInit)) // 4 * q.nComb / 100
	nc:=new(big.Int).Set(q.nComb)
	nc.Div(nc, big.NewInt(100))
	initPos.Mul(initPos, nc)
	fmt.Println("percInit:", percInit, "initial pos:",initPos.String(), "from", q.nComb.String())

	q.set_n_board(initPos)
	i := 1
	nSol := 0

	for q.Next() {
		i++
		if q.isValid() {
			nSol++
			fmt.Printf("%.0f %8d: %v \n", 100.*float64(i)/float64(q.nComb.Int64()), nSol, q.board)
			break
		}
	}
}


func TestScan() {
	nq := 31
	q := NewQueens(nq)
	t0:=time.Now()
	q.FindFirst(1, false)
	fmt.Printf("nQueens: %d solutions: %v evals: %v time: %v\n", nq, q.solutions, q.countEvals, time.Since(t0))
}