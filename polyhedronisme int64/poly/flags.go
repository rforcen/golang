package poly

import (
	"fmt"
	"sort"
	"unsafe"
)

// Int4
type Int4 struct {
	i0, i1, i2, i3 int
}

func Int4_1(i0 int) Int4 {
	return Int4{i0 + 1, 0, 0, 0}
}

func Int4_2(i0, i1 int) Int4 {
	return Int4{i0 + 1, i1 + 1, 0, 0}
}

func Int4_3(i0, i1, i2 int) Int4 {
	return Int4{i0 + 1, i1 + 1, i2 + 1, 0}
}

func Int4_4(i0, i1, i2, i3 int) Int4 {
	return Int4{i0 + 1, i1 + 1, i2 + 1, i3 + 1}
}

func I4_min(i1 int, i2 int) Int4 {
	if i1 < i2 {
		return Int4_2(i1, i2)
	}
	return Int4_2(i2, i1)
}

func I4_min3(i int, v1 int, v2 int) Int4 {
	if v1 < v2 {
		return Int4_3(i, v1, v2)
	}
	return Int4_3(i, v2, v1)
}

func ToInt(s string) int {
	szi := unsafe.Sizeof(int(0))
	if len(s) > int(szi) {
		s = s[0:szi]
	}
	for i := len(s); i < int(szi); i++ {
		s += "_"
	}
	return *(*int)(unsafe.Pointer(unsafe.StringData(s)))
}

// VertexIndex

type VertexIndex struct {
	Index  int
	Vertex Vertex
}

// Flag

type Flag struct {
	Vertexes  []Vertex
	Faces     [][]int
	Fcs       [][]Int4
	Faceindex int
	Valid     bool

	Facemap map[Int4]VertexIndex
	M_map   map[Int4]map[Int4]Int4
}

func NewFlag() *Flag {
	return &Flag{
		Facemap: make(map[Int4]VertexIndex),
		M_map:   make(map[Int4]map[Int4]Int4),
		Valid:   true,
	}
}
func (f *Flag) SetVertexes(vs Vertexes) {
	for i, v := range vs {
		f.AddVertex(Int4_1(i), v)
	}
}

func (f *Flag) AddVertex(ix Int4, vtx Vertex) {
	f.Facemap[ix] = VertexIndex{f.Faceindex, vtx}
	f.Faceindex++
}

func (f *Flag) AddFace(i0 Int4, i1 Int4, i2 Int4) {
	if _, ok := f.M_map[i0]; !ok {
		f.M_map[i0] = make(map[Int4]Int4)
	}
	f.M_map[i0][i1] = i2
}

func (f *Flag) AddFaceVect(v []Int4) {
	f.Fcs = append(f.Fcs, v)
}

func (f *Flag) ReindexVertexes() {
	f.Vertexes = make(Vertexes, 0, len(f.Facemap))

	i := 0
	for k, v := range f.Facemap {
		f.Vertexes = append(f.Vertexes, v.Vertex)
		f.Facemap[k] = VertexIndex{i, v.Vertex}
		i++
	}
}

const max_iters = 100

func (f *Flag) ProcessM_map() bool {
	for i, face := range f.M_map {

		var v Int4
		for _, iv := range face { // get 1st key
			v = iv
			break
		}
		v0 := v

		// traverse v to create a face
		face_tmp := make([]int, 0, max_iters)

		for range max_iters {
			face_tmp = append(face_tmp, f.Facemap[v].Index)
			v = f.M_map[i][v]

			if v == v0 { // found, closed loop
				break
			}
		}

		if v != v0 { // couldn't close loop -> invalid
			fmt.Printf("dead loop v:%v, v0:%v\n", v, v0)
			f.Faces = nil
			f.Valid = false
			return f.Valid
		}

		f.Faces = append(f.Faces, face_tmp)

	}

	return f.Valid
}

func (f *Flag) ProcessFcs() {	 // append faces
	for _, fc := range f.Fcs {
		face_tmp := make(Face, len(fc))
		for j, vix := range fc {
			face_tmp[j] = f.Facemap[vix].Index
		}
		f.Faces = append(f.Faces, face_tmp)
	}
}

func (f *Flag) Check() {
	for _, face := range f.Faces {
		if len(face) < 3 {
			f.Valid = false
			return
		}
		for _, iv := range face {
			if iv > len(f.Vertexes) {
				f.Valid = false
				return
			}
		}
	}
}

func (f *Flag) UniqueFaces() { // remove dupes from f.faces in sorted comparison

	sorted_key := func(face Face) string { // create a Face sorted key string
		face_sorted := make(Face, len(face)) // make a copy, sort and return key string
		copy(face_sorted, face)
		sort.Ints(face_sorted)
		return fmt.Sprint(face_sorted)
	}

	mfs := make(map[string]Face) // map of sorted faces

	for _, face := range f.Faces { // sort faces to map
		key := sorted_key(face) // key: string printout of sorted face

		if _, ok := mfs[key]; !ok { // unique ?
			mfs[key] = face // yes, add original
		} // else is dupe -> discard
	}

	if len(mfs) != len(f.Faces) { // update required?
		f.Faces = make(Faces, len(mfs))
		i := 0 // traverse map of unique sorted faces
		for _, face := range mfs {
			f.Faces[i] = face // copy original face
			i++
		}
	}
}

func (f *Flag) ToPoly() bool {
	f.ReindexVertexes()

	if f.ProcessM_map() {
		f.ProcessFcs()
		f.UniqueFaces() // remove dupes preserving face order
		f.Check()
	}
	return f.Valid
}

func (f *Flag) CreatePoly(tr string, p *Polyhedron) *Polyhedron {
	if !f.ToPoly() {
		return p
	}
	rp := Polyhedron{Name: tr + p.Name, Vertexes: f.Vertexes, Faces: f.Faces}
	return rp.Normalize().ScaleUnit()
}
