package poly

import (
	"fmt"
	"hash/fnv"
	"sort"
	"unsafe"
)

// Int4
type Int4 struct {
	i0, i1, i2, i3 uint32
}

func Int4_1(i0 uint32) Int4 {
	return Int4{i0 + 1, 0, 0, 0}
}

func Int4_2(i0, i1 uint32) Int4 {
	return Int4{i0 + 1, i1 + 1, 0, 0}
}

func Int4_3(i0, i1, i2 uint32) Int4 {
	return Int4{i0 + 1, i1 + 1, i2 + 1, 0}
}

func Int4_4(i0, i1, i2, i3 uint32) Int4 {
	return Int4{i0 + 1, i1 + 1, i2 + 1, i3 + 1}
}

func I4_min(i1 uint32, i2 uint32) Int4 {
	if i1 < i2 {
		return Int4_2(i1, i2)
	}
	return Int4_2(i2, i1)
}

func I4_min3(i uint32, v1 uint32, v2 uint32) Int4 {
	if v1 < v2 {
		return Int4_3(i, v1, v2)
	}
	return Int4_3(i, v2, v1)
}

func ToInt(s string) uint32 {
	szi := unsafe.Sizeof(uint32(0))
	if len(s) > int(szi) {
		s = s[0:szi]
	}
	for i := len(s); i < int(szi); i++ {
		s += "_"
	}
	return *(*uint32)(unsafe.Pointer(unsafe.StringData(s)))
}

// uint32 slice, Define a custom type for a slice of uint32

// Implement the sort.Interface for Uint32Slice
func (s Face) Len() int {
	return len(s)
}

func (s Face) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s Face) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Int4 hash optimal implementation

func (i Int4) Hash() uint64 {
	const prime1 = uint64(0x9e3779b97f4a7c15)
	const prime2 = uint64(0xbf58476d1ce4e5b9)
	const fnv_offset_basis = uint64(0xcbf29ce484222325)
	const final_mixer = uint64(0xff51afd7ed558ccd)

	h := fnv_offset_basis
	h = (h ^ uint64(i.i0)) * prime1
	h = (h ^ uint64(i.i1)) * prime2
	h = (h ^ uint64(i.i2)) * prime1
	h = (h ^ uint64(i.i3)) * prime2

	return (h^(h>>33))*final_mixer ^ ((h ^ (h >> 33)) * final_mixer >> 33)
}

// VertexIndex
type VertexIndex struct {
	Index  uint32
	Vertex Vertex
}

// Flag

type Flag struct {
	Vertexes  []Vertex
	Faces     Faces
	Fcs       [][]Int4
	Faceindex uint32
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
		f.AddVertex(Int4_1(uint32(i)), v)
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

	i := uint32(0)
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
		face_tmp := make(Face, 0, max_iters)

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

func (f *Flag) ProcessFcs() { // append faces
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
			if iv > uint32(len(f.Vertexes)) {
				f.Valid = false
				return
			}
		}
	}
}

func (f *Flag) UniqueFaces() { // remove dupes from f.faces in sorted comparison

	hash_key := func(face Face) uint64 { // create a sorted Face key uin64
		Hash := func(face Face) uint64 {
			h := fnv.New64a()
			h.Write(unsafe.Slice((*byte)(unsafe.Pointer(&face[0])), len(face)*4))
			return h.Sum64()
		}

		face_sorted := make(Face, len(face)) // make a copy, sort and return key string
		copy(face_sorted, face)
		sort.Sort(face_sorted)
		return Hash(face_sorted)
	}

	mfs := make(map[uint64]Face) // map of sorted faces

	for _, face := range f.Faces { // sort faces to map
		key := hash_key(face) // key: string printout of sorted face

		if _, ok := mfs[key]; !ok { // unique ?
			mfs[key] = face // yes, add original
		} // else is dupe -> discard
	}

	if len(mfs) != len(f.Faces) { // update required?
		f.Faces = make(Faces, len(mfs))
		i := uint32(0) // traverse map of unique sorted faces
		for _, face := range mfs {
			f.Faces[i] = face // copy original face
			i++
		}
	}
}

func (f *Flag) ToPoly() bool {
	f.ReindexVertexes()

	f.Faces = make(Faces, 0, len(f.Fcs)+len(f.M_map)) // make an educated guess

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
	return rp.Optimize().ScaleUnit()
}
