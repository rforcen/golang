package polygonizer

/*
 * Adapted from
 *
 * C code from the article
 * "An Implicit Surface Polygonizer"
 * http::www.unchainedgeometry.com/jbloom/papers/Polygonizer.pdf
 * by Jules Bloomenthal, jules@bloomenthal.com
 * in "Graphics Gems IV", Academic Press, 1994
 */

import (
	"github.com/chewxy/math32"
	"math/rand/v2"
)

// main call
func Polygonize(function func(float32, float32, float32) float32, size float32, bounds int) (vertexes []VERTEX, triangles []TRIANGLE, err string) {
	return polygonize(function, size, bounds, 0, 0, 0, NOTET)
}

// TYPES
type POINT struct {
	x, y, z float32
}
type TEST struct { // test the function for a signed value
	p     *POINT  // location of test
	value float32 // function value at p
	ok    bool    // if value is of correct sign
}
type VERTEX struct { // surface vertex
	position, normal *POINT // position and surface normal
}

type TRIANGLE struct {
	i1, i2, i3 int
}

type CORNER struct { // corner of a cube
	i, j, k        int     // (i, j, k) is index within lattice
	x, y, z, value float32 // location and function value
}
type CUBE struct { // partitioning cell (cube)
	i, j, k int        // lattice location of cube
	corners [8]*CORNER // eight corners
}
type CUBES struct { // linked list of cubes acting as stack
	cube *CUBE  // a single cube
	next *CUBES // remaining elements
}
type CENTERLIST struct { // list of cube locations
	i, j, k int         // cube location
	next    *CENTERLIST // remaining elements
}
type CORNERLIST struct { // list of corners
	i, j, k int         // corner id
	value   float32     // corner value
	next    *CORNERLIST // remaining elements
}
type EDGELIST struct { // list of edges
	i1, j1, k1, i2, j2, k2 int       // edge corner ids
	vid                    int       // vertex id
	next                   *EDGELIST // remaining elements
}
type INTLIST struct { // list of integers
	i    int      // an integer
	next *INTLIST // remaining elements
}
type INTLISTS struct { // list of list of integers
	list *INTLIST  // a list of integers
	next *INTLISTS // remaining elements
}
type PROCESS struct { // parameters, function, storage
	// implicit surface function
	function func(float32, float32, float32) float32

	size, delta float32 // cube size, normal delta
	bounds      int     // cube range within lattice
	start       *POINT  // start point on surface
	cubes       *CUBES  // active cubes
	vertices    []VERTEX
	triangles   []TRIANGLE

	centers []*CENTERLIST // cube center hash table
	corners []*CORNERLIST // corner value hash table
	edges   []*EDGELIST   // edge and vertex id hash table
}

const RES = 10 // # converge iterations
var cubetable [256]*INTLISTS

const (
	NSAMPLES = 10 * 1000 // org. 10000
	HASHBIT  = 5
	HASHSIZE = 1 << (3 * HASHBIT) // hash table size (32768)
	MASK     = (1 << HASHBIT) - 1

	TET   = 0 // use tetrahedral decomposition
	NOTET = 1 // no tetrahedral decomposition
)
const (
	L   = 0  // left direction: = -x, -i
	R   = 1  // right direction: = +x, +i
	B   = 2  // bottom direction: -y, -j
	T   = 3  // top direction: = +y, +j
	N   = 4  // near direction: = -z, -k
	F   = 5  // far direction: = +z, +k
	LBN = 0  // left bottom near corner
	LBF = 1  // left bottom far corner
	LTN = 2  // left top near corner
	LTF = 3  // left top far corner
	RBN = 4  // right bottom near corner
	RBF = 5  // right bottom far corner
	RTN = 6  // right top near corner
	RTF = 7  // right top far corner
	LB  = 0  // left bottom edge
	LT  = 1  // left top edge
	LN  = 2  // left near edge
	LF  = 3  // left far edge
	RB  = 4  // right bottom edge
	RT  = 5  // right top edge
	RN  = 6  // right near edge
	RF  = 7  // right far edge
	BN  = 8  // bottom near edge
	BF  = 9  // bottom far edge
	TN  = 10 // top near edge
	TF  = 11 // top far edge
)

// edge: LB, LT, LN, LF, RB, RT, RN, RF, BN, BF, TN, TF
var corner1 = []int{LBN, LTN, LBN, LBF, RBN, RTN, RBN, RBF, LBN, LBF, LTN, LTF}
var corner2 = []int{LBF, LTF, LTN, LTF, RBF, RTF, RTN, RTF, RBN, RBF, RTN, RTF}
var leftface = []int{B, L, L, F, R, T, N, R, N, B, T, F}

// face on left when going corner1 to corner2
var rightface = []int{L, T, N, L, B, R, R, F, B, F, N, T}

// helpers

func HASH(i, j, k int) int { return ((i & MASK) << HASHBIT) | ((j & MASK) << HASHBIT) | (k & MASK) }
func BIT(i, bit int) int   { return (i >> bit) & 1 }
func FLIP(i, bit int) int  { return i ^ 1<<bit }

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func ternary(cnd bool, a, b int) int { // ? 'c lang.' operator  cond ? a : b
	if cnd {
		return a
	}
	return b
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func rnd_05() float32 { // random -0.5 .. +0.5
	return float32(rand.Float64() - 0.5)
}

// add triangle to process
func (p *PROCESS) addTrig(a, b, c int) bool {
	p.triangles = append(p.triangles, TRIANGLE{a, b, c})
	return true
}

// docube: triangulate the cube directly, without decomposition
func (p *PROCESS) docube(cube *CUBE) bool {
	index := 0
	for i := range 8 {
		if cube.corners[i].value > 0.0 {
			index += (1 << i)
		}
	}
	for polys := cubetable[index]; polys != nil; polys = polys.next {
		a, b, count := -1, -1, 0
		for edges := polys.list; edges != nil; edges = edges.next {
			c1 := cube.corners[corner1[edges.i]]
			c2 := cube.corners[corner2[edges.i]]
			c := p.vertid(c1, c2)
			count++
			if count > 2 {
				p.addTrig(a, b, c)
			}
			if count < 3 {
				a = b
			}
			b = c
		}
	}
	return true
}

// getedge: return vertex id for edge; return -1 if not set

func (p *PROCESS) getedge(c1, c2 CORNER) int {
	if c1.i > c2.i || (c1.i == c2.i && (c1.j > c2.j || (c1.j == c2.j && c1.k > c2.k))) {
		c1, c2 = c2, c1
	}
	index := HASH(c1.i, c1.j, c1.k) + HASH(c2.i, c2.j, c2.k)
	q := p.edges[index]
	for ; q != nil; q = q.next {
		if q.i1 == c1.i && q.j1 == c1.j && q.k1 == c1.k &&
			q.i2 == c2.i && q.j2 == c2.j && q.k2 == c2.k {
			return q.vid
		}
	}
	return -1
}

// setedge: set vertex id for edge

func (p *PROCESS) setedge(c1, c2 CORNER, vid int) {
	if c1.i > c2.i || (c1.i == c2.i && (c1.j > c2.j || (c1.j == c2.j && c1.k > c2.k))) {
		c1, c2 = c2, c1
	}
	index := HASH(c1.i, c1.j, c1.k) + HASH(c2.i, c2.j, c2.k)
	p.edges[index] = &EDGELIST{c1.i, c1.j, c1.k, c2.i, c2.j, c2.k, vid, p.edges[index]}
}

func (p *PROCESS) vnormal(vertex *VERTEX) {
	f := p.function(vertex.position.x, vertex.position.y, vertex.position.z)

	x := p.function(vertex.position.x+p.delta, vertex.position.y, vertex.position.z) - f
	y := p.function(vertex.position.x, vertex.position.y+p.delta, vertex.position.z) - f
	z := p.function(vertex.position.x, vertex.position.y, vertex.position.z+p.delta) - f

	f = math32.Sqrt(x*x + y*y + z*z)

	if f != 0.0 {
		x /= f
		y /= f
		z /= f
	}
	vertex.normal = &POINT{x, y, z}
}

// vertid: return index for vertex on edge:
// c1->value and c2->value are presumed of different sign
// return saved index if any; else compute vertex and save

func (p *PROCESS) vertid(c1, c2 *CORNER) int {
	vtx := VERTEX{position: &POINT{}, normal: &POINT{}}

	vid := p.getedge(*c1, *c2)
	if vid != -1 {
		return vid // previously computed
	}
	p.converge(&POINT{c1.x, c1.y, c1.z}, &POINT{c2.x, c2.y, c2.z}, c1.value, vtx.position) // position
	p.vnormal(&vtx)                                                                        // normal

	p.vertices = append(p.vertices, vtx) // add new vertex

	vid = len(p.vertices) - 1
	p.setedge(*c1, *c2, vid)
	return vid
}

func (p *PROCESS) dotet(cube *CUBE, c1, c2, c3, c4 int) bool {
	a, b, c, d := cube.corners[c1], cube.corners[c2], cube.corners[c3], cube.corners[c4]

	index := 0
	e1, e2, e3, e4, e5, e6 := 0, 0, 0, 0, 0, 0

	apos := a.value > 0.0
	if apos {
		index += 8
	}
	bpos := b.value > 0.0
	if bpos {
		index += 4
	}
	cpos := c.value > 0.0
	if cpos {
		index += 2
	}
	dpos := d.value > 0.0
	if dpos {
		index += 1
	}
	// index is now 4-bit number representing one of the 16 possible cases
	if apos != bpos {
		e1 = p.vertid(a, b)
	}
	if apos != cpos {
		e2 = p.vertid(a, c)
	}
	if apos != dpos {
		e3 = p.vertid(a, d)
	}
	if bpos != cpos {
		e4 = p.vertid(b, c)
	}
	if bpos != dpos {
		e5 = p.vertid(b, d)
	}
	if cpos != dpos {
		e6 = p.vertid(c, d)
	}

	// 14 productive tetrahedral cases (0000 and 1111 do not yield polygons
	switch index {
	case 1:
		return p.addTrig(e5, e6, e3)
	case 2:
		return p.addTrig(e2, e6, e4)
	case 3:
		return p.addTrig(e3, e5, e4) && p.addTrig(e3, e4, e2)
	case 4:
		return p.addTrig(e1, e4, e5)
	case 5:
		return p.addTrig(e3, e1, e4) && p.addTrig(e3, e4, e6)
	case 6:
		return p.addTrig(e1, e2, e6) && p.addTrig(e1, e6, e5)
	case 7:
		return p.addTrig(e1, e2, e3)
	case 8:
		return p.addTrig(e1, e3, e2)
	case 9:
		return p.addTrig(e1, e5, e6) && p.addTrig(e1, e6, e2)
	case 10:
		return p.addTrig(e1, e3, e6) && p.addTrig(e1, e6, e4)
	case 11:
		return p.addTrig(e1, e5, e4)
	case 12:
		return p.addTrig(e3, e2, e4) && p.addTrig(e3, e4, e5)
	case 13:
		return p.addTrig(e6, e2, e4)
	case 14:
		return p.addTrig(e5, e3, e6)
	}
	return true
}

// setcenter: set (i,j,k) entry of table[]
// return 1 if already set; otherwise, set and return 0

func (p *PROCESS) setcenter(i, j, k int) bool {
	index := HASH(i, j, k)
	q := p.centers[index]

	for l := q; l != nil; l = l.next {
		if l.i == i && l.j == j && l.k == k {
			return true
		}
	}
	p.centers[index] = &CENTERLIST{i: i, j: j, k: k, next: q}

	return false
}

// setcorner: return corner with the given lattice location
// set (and cache) its function value

func (p *PROCESS) setcorner(i, j, k int) *CORNER {
	// for speed, do corner value caching here

	c := &CORNER{i: i, j: j, k: k, x: p.start.x + (float32(i)-.5)*p.size, y: p.start.y + (float32(j)-.5)*p.size, z: p.start.z + (float32(k)-.5)*p.size}
	index := HASH(i, j, k)
	l := p.corners[index]

	for ; l != nil; l = l.next {
		if l.i == i && l.j == j && l.k == k {
			c.value = l.value
			return c
		}
	}
	l = &CORNERLIST{i: i, j: j, k: k, value: p.function(c.x, c.y, c.z), next: p.corners[index]}
	c.value = l.value
	l.next = p.corners[index]
	p.corners[index] = l

	return c
}

func (p *PROCESS) find(sign int, x, y, z float32) TEST {
	psize := p.size

	for range NSAMPLES {
		rx, ry, rz := x+psize*rnd_05(), y+psize*rnd_05(), z+psize*rnd_05()
		value := p.function(rx, ry, rz)

		if sign == bool2int(value > 0.0) {
			return TEST{p: &POINT{rx, ry, rz}, value: value, ok: true}
		}
		psize = psize * 1.0005 // slowly expand search outwards
	}
	return TEST{ok: false}
}

func polygonize(
	function func(float32, float32, float32) float32,
	size float32, bounds int, x, y, z float32, mode int) (Vertexes []VERTEX, Triangles []TRIANGLE, msg string) {

	p := &PROCESS{function: function, size: size, bounds: bounds, delta: float32(size) / float32(RES*RES),
		start:    &POINT{},
		vertices: make([]VERTEX, 0, NSAMPLES), triangles: make([]TRIANGLE, 0, NSAMPLES),
		corners: make([]*CORNERLIST, HASHSIZE), centers: make([]*CENTERLIST, HASHSIZE), edges: make([]*EDGELIST, HASHSIZE*2)}

	pos := make([]int, 8)

	for i := range 256 {
		cubetable[i] = nil

		done := make([]int, 12)
		for c := range pos {
			pos[c] = BIT(i, c)
		}

		for e := range done {
			if done[e] == 0 && (pos[corner1[e]] != pos[corner2[e]]) {
				var ints *INTLIST = nil
				start, edge := e, e
				face := ternary(pos[corner1[e]] != 0, rightface[e], leftface[e])
				for {
					nextcwedge := func(edge, face int) int {
						switch edge {
						case LB:
							return ternary(face == L, LF, BN)
						case LT:
							return ternary(face == L, LN, TF)
						case LN:
							return ternary(face == L, LB, TN)
						case LF:
							return ternary(face == L, LT, BF)
						case RB:
							return ternary(face == R, RN, BF)
						case RT:
							return ternary(face == R, RF, TN)
						case RN:
							return ternary(face == R, RT, BN)
						case RF:
							return ternary(face == R, RB, TF)
						case BN:
							return ternary(face == B, RB, LN)
						case BF:
							return ternary(face == B, LB, RF)
						case TN:
							return ternary(face == T, LT, RN)
						case TF:
							return ternary(face == T, RT, LF)
						default:
							return -1 // should never reach here
						}
					}

					edge = nextcwedge(edge, face)
					done[edge] = 1
					if pos[corner1[edge]] != pos[corner2[edge]] {
						ints = &INTLIST{i: edge, next: ints}
						if edge == start {
							break
						}
						face = ternary(face == leftface[edge], rightface[edge], leftface[edge])
					}
				}
				cubetable[i] = &INTLISTS{list: ints, next: cubetable[i]}
			}
		}
	}
	// find point on surface, beginning search at (x, y, z):
	in := p.find(1, x, y, z)
	out := p.find(0, x, y, z)

	if !in.ok || !out.ok {
		return nil, nil, "can't find starting point"
	}
	p.converge(in.p, out.p, in.value, p.start)

	// push initial cube on stack:
	p.cubes = &CUBES{cube: &CUBE{}, next: nil}

	// set corners of initial cube:
	for n := range 8 {
		p.cubes.cube.corners[n] = p.setcorner(BIT(n, 2), BIT(n, 1), BIT(n, 0))
	}

	p.setcenter(0, 0, 0)

	for p.cubes != nil { // process active cubes till none left
		c := p.cubes.cube

		noabort := false
		if mode == TET { // either decompose into tetrahedra and polygonize:
			noabort = p.dotet(c, LBN, LTN, RBN, LBF) &&
				p.dotet(c, RTN, LTN, LBF, RBN) &&
				p.dotet(c, RTN, LTN, LTF, LBF) &&
				p.dotet(c, RTN, RBN, LBF, RBF) &&
				p.dotet(c, RTN, LBF, LTF, RBF) &&
				p.dotet(c, RTN, LTF, RTF, RBF)
		} else { // or polygonize the cube directly:
			noabort = p.docube(c)
		}
		if !noabort {
			return nil, nil, "aborted"
		}

		// pop current cube from stack
		p.cubes = p.cubes.next
		// test six face directions, maybe add to stack:
		p.testface(c.i-1, c.j, c.k, c, L, LBN, LBF, LTN, LTF)
		p.testface(c.i+1, c.j, c.k, c, R, RBN, RBF, RTN, RTF)
		p.testface(c.i, c.j-1, c.k, c, B, LBN, LBF, RBN, RBF)
		p.testface(c.i, c.j+1, c.k, c, T, LTN, LTF, RTN, RTF)
		p.testface(c.i, c.j, c.k-1, c, N, LBN, LTN, RBN, RTN)
		p.testface(c.i, c.j, c.k+1, c, F, LBF, LTF, RBF, RTF)
	}

	return scaleVertices(p.vertices), p.triangles, ""
}

func (p *PROCESS) converge(p1, p2 *POINT, v float32, pnt *POINT) {

	var pos, neg POINT

	if v < 0 {
		pos = *p2
		neg = *p1
	} else {
		pos = *p1
		neg = *p2
	}
	for range RES {
		*pnt = POINT{x: 0.5 * (pos.x + neg.x), y: 0.5 * (pos.y + neg.y), z: 0.5 * (pos.z + neg.z)}

		if p.function(pnt.x, pnt.y, pnt.z) > 0.0 {
			pos = *pnt
		} else {
			neg = *pnt
		}
	}
}

func (p *PROCESS) testface(i, j, k int, old *CUBE, face int, c1, c2, c3, c4 int) {
	facebit := [6]int{2, 2, 1, 1, 0, 0}
	pos := old.corners[c1].value > 0.0
	bit := facebit[face]

	// test if no surface crossing, cube out of bounds, or already visited:
	if (old.corners[c2].value > 0) == pos &&
		(old.corners[c3].value > 0) == pos &&
		(old.corners[c4].value > 0) == pos {
		return
	}
	if absInt(i) > p.bounds || absInt(j) > p.bounds || absInt(k) > p.bounds {
		return
	}
	if p.setcenter(i, j, k) {
		return
	}

	// create newList cube:
	newList := CUBE{i: i, j: j, k: k}
	newList.corners[FLIP(c1, bit)] = old.corners[c1]
	newList.corners[FLIP(c2, bit)] = old.corners[c2]
	newList.corners[FLIP(c3, bit)] = old.corners[c3]
	newList.corners[FLIP(c4, bit)] = old.corners[c4]

	for n := range 8 {
		if newList.corners[n] == nil {
			newList.corners[n] = p.setcorner(i+BIT(n, 2), j+BIT(n, 1), k+BIT(n, 0))
		}
	}

	//add cube to top of stack:
	p.cubes = &CUBES{cube: &newList, next: p.cubes}
}

func scaleVertices(vertices []VERTEX) []VERTEX {
	maxv := -float32(math32.MaxFloat32)
	for i := range vertices {
		maxv = math32.Max(maxv, math32.Abs(vertices[i].position.x))
		maxv = math32.Max(maxv, math32.Abs(vertices[i].position.y))
		maxv = math32.Max(maxv, math32.Abs(vertices[i].position.z))
	}
	if maxv != 0 {
		for i := range vertices {
			vertices[i].position.x /= maxv
			vertices[i].position.y /= maxv
			vertices[i].position.z /= maxv
		}
	}
	return vertices
}
