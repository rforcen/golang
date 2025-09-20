package sudoku

type Bitset struct {
	bits uint64
	n    int
	len  int
	ix   int
}

func newBitset(n int) Bitset {
	if n > 64 {
		panic("max bitset size is 64")
	}

	return Bitset{
		bits: 0,
		n:    n,
		len:  0,
		ix:   0,
	}
}

func (bs *Bitset) clear() {
	bs.bits = 0
	bs.len = 0
	bs.ix = 0
}

func (bs *Bitset) set(n int) {
	bs.len++
	bs.bits |= 1 << n
}

func (bs *Bitset) get(n int) bool {
	return bs.bits&(1<<n) == 0 // 0 is empty
}

func (bs *Bitset) next() int {
	for {
		if bs.get(bs.ix) {
			bs.ix++
			if bs.ix <= bs.n {
				return bs.ix - 1
			} else {
				return -1
			}
		}
		bs.ix++
		if bs.ix > bs.n {
			break
		}
	}
	return -1
}
