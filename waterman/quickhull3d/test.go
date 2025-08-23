package qh

import (
	"fmt"
	"math/rand"
)

func TestVector3d() {
	v := newVector3d(1, 2, 3)
	v.add(newVector3d(4, 5, 6))
	v.sub(newVector3d(1, 1, 1))
	v.scale(2)
	v.normalize()
	v.setZero()
	v.cross(newVector3d(1, 2, 3), newVector3d(4, 5, 6))
	v.setRandom(0, 1)
	fmt.Println(v.String())
}

func TestFace() {

	n := 10000

	vertexes := make([]*Vertex, 0, n)
	indices := make([]int, 0, n)

	for i := 0; i < n; i++ {
		v0 := NewVertex(rand.Float64(), rand.Float64(), rand.Float64(), i)
		vertexes = append(vertexes, v0)
		indices = append(indices, i)
	}
	face := create(vertexes, indices)
	fmt.Printf("face with %d vertexes: %v\n", face.numVertices(), face)
}

func TestWaterman() {
	coords := watermanPolyhedron(3000)

	pts := make([]*Point3d, 0, len(coords)/3)
	for i := 0; i < len(coords); i += 3 {
		pts = append(pts, newPoint3d(coords[i], coords[i+1], coords[i+2]))
	}
	fmt.Println(pts[0], pts[len(pts)-1])

	vtx := make([]*Vertex, 0, len(pts))
	idxs := make([]int, 0, len(pts))

	for i := range len(pts) {
		vtx = append(vtx, NewVertexFromPoint(pts[i], i))
		idxs = append(idxs, i)
	}
	face := create(vtx, idxs)

	fmt.Println(*face)
}

func TestLenWaterman() {
	for i := 10; i < 100; i++ {
		rad := float64(i)
		coords := watermanPolyhedron(rad)
		fmt.Printf("#coords: %d, rad: %.0f\n", len(coords), rad)
	}
}

func TestQuickHull3d() {
	coords := watermanPolyhedron(102)
	qh := NewQuickHull3D(coords)
	coords = nil

	for _, i := range []int{0, len(qh.pointBuffer) / 2, len(qh.pointBuffer) - 1} {
		p := qh.pointBuffer[i]
		fmt.Printf("%6d: %v %d %v %v %v\n", i, p.pnt, p.index, p.face, p.next, p.prev)
	}

	for i := 10; i < 100; i++ {
		rad := float64(i)
		coords := watermanPolyhedron(rad)
		qh := NewQuickHull3D(coords)
		fmt.Printf("#coords: %d, rad: %.0f faces: %d | vertices: %d \n", len(coords), rad, qh.GetNumFaces(), qh.GetNumVertices())
	}
}
