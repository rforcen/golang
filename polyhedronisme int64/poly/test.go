package poly

import (
	"fmt"
)

func TestVertex() {
	v := NewVertex(1, 2, 3).Unit()
	fmt.Println(*v)
	v0 := NewVertex(1, 2, 3)
	fmt.Println(*v0)
	fmt.Println(v.Dot(v0))
	fmt.Println(v.Cross(v0))
	fmt.Println(v.Norm())
	fmt.Println(v.NormSquared())
	fmt.Println(v.MaxAbs())

	fmt.Println(Normal(NewVertex(1, 2, 3), NewVertex(2, 1, 3), NewVertex(3, 2, 1)))
}

func TestPoly() {
	fmt.Println(Tetrahedron)
	fmt.Println(Cube)
	fmt.Println(Icosahedron)
	fmt.Println(Octahedron)
	fmt.Println(Dodecahedron)

	Tetrahedron.ToVertexes()
	fmt.Println(Tetrahedron)

	p := NewPolyhedron(Cube).Normalize().Recalc()

	fmt.Println("Vertexes: ", p.Vertexes)
	fmt.Println("Faces   : ", p.Faces)
	fmt.Println("Normals : ", p.Normals)
	fmt.Println("Areas   : ", p.Areas)
	fmt.Println("Colors  : ", p.Colors)
	fmt.Println("Centers : ", p.Centers)

	p.WriteObj()

}

func TestJohnson() {
	for _, p := range Johnson {
		p.ToVertexes().Normalize().Recalc()
		fmt.Printf("%4s %5d %5d %t\n", p.Name, len(p.Vertexes), len(p.Faces), p.Check())
	}
}

func TestFlag() {
	f := NewFlag()

	p := Cube
	p.ToVertexes()

	f.SetVertexes(p.Vertexes)

	f.AddFace(Int4_1(0), Int4_1(1), Int4_1(2))
	f.AddFace(Int4_1(0), Int4_1(2), Int4_1(3))
	f.AddFace(Int4_1(1), Int4_1(2), Int4_1(3))
	f.AddFace(Int4_1(1), Int4_1(3), Int4_1(4))

	f.ReindexVertexes()

	fmt.Println("Vertexes: ", f.Vertexes)
	fmt.Println("Facemap : ", f.Facemap)
	fmt.Println("M_map   : ", f.M_map)
	fmt.Println("Fcs     : ", f.Fcs)

	f.ProcessM_map()
	fmt.Println("Valid   : ", f.Valid)
	fmt.Println("Faces   : ", f.Faces)

	f.AddFaceVect([]Int4{Int4_1(0), Int4_1(1), Int4_1(2)})
	f.ProcessFcs()
	fmt.Println("Valid   : ", f.Valid)
	fmt.Println("Faces   : ", f.Faces)

	f.Faces = [][]int{{1, 2, 3},
		{4, 5, 6},
		{1, 2, 3},
		{7, 8, 9},
		{8, 7, 9},
		{9, 8, 7},
		{4, 5, 6},
		{2, 1, 3},
		{3, 2, 1}}
	f.UniqueFaces()
	fmt.Println("Faces   : ")
	for _, face := range f.Faces {
		fmt.Println(face, "-"+fmt.Sprint(face)+"-")
	}

	f.ToPoly()
	// fmt.Println(f)
}

func TestDodeca() {
	p := NewPolyhedron(Dodecahedron)
	
	fmt.Println("Vertexes: ", p.Vertexes)
	fmt.Println("Faces   : ", p.Faces)
	fmt.Println("Normals : ", p.Normals)
	fmt.Println("Areas   : ", p.Areas)
	fmt.Println("Colors  : ", p.Colors)
	fmt.Println("Centers : ", p.Centers)
}

func TestTransforms() {
	fmt.Println()

	fmt.Printf("%x\n", ToInt("0"))
	fmt.Printf("%x\n", ToInt("1"))

	pp := NewPolyhedron(Cube)
	pp = pp.Kiss_n(0, 0.1)
	pp.WriteObj()
}
