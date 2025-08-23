package poly

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"slices"

	"github.com/chewxy/math32"
)

type RawVertexes [][]float32
type Vertexes []Vertex
type Face []uint32
type Faces []Face

type Polyhedron struct {
	VertexRaw RawVertexes
	Name      string
	Vertexes  Vertexes
	Faces     Faces
	Normals   Vertexes
	Colors    Vertexes
	Centers   Vertexes
	Areas     []float32
}

func (p *Polyhedron) ToVertexes() *Polyhedron {
	p.Vertexes = make(Vertexes, len(p.VertexRaw))
	for i, v := range p.VertexRaw {
		p.Vertexes[i] = Vertex{v[0], v[1], v[2]}
	}
	return p
}

func NewPolyhedron(p *Polyhedron) *Polyhedron { // to use in preset Polyhedrons

	res := &Polyhedron{
		Name:      p.Name,
		VertexRaw: p.VertexRaw, // use its vertexes
	}
	res.Faces = make(Faces, len(p.Faces))
	for i, face := range p.Faces {
		res.Faces[i] = make(Face, len(face))
		copy(res.Faces[i], face)
	}

	res.ToVertexes()
	res.CalcNormals()
	res.CalcAreas()
	res.CalcColors()
	res.CalcCenters()
	res.ScaleUnit()

	return res
}

// normals
func (p *Polyhedron) CalcNormals() Vertexes {
	p.Normals = make(Vertexes, len(p.Faces))

	for i, face := range p.Faces {
		v0, v1, v2 := p.Vertexes[face[0]], p.Vertexes[face[1]], p.Vertexes[face[2]]

		p.Normals[i] = *Normal(&v0, &v1, &v2)
	}
	return p.Normals
}

// areas (requires normals)
func (p *Polyhedron) CalcAreas() []float32 {
	p.Areas = make([]float32, len(p.Faces))

	for iface, face := range p.Faces {
		vsum := Vertex{0, 0, 0}
		fl := len(face)
		v1, v2 := p.Vertexes[face[fl-2]], p.Vertexes[face[fl-1]]

		for _, v := range face {
			vsum = *vsum.Add(v1.Cross(&v2))
			v1, v2 = v2, p.Vertexes[v]
		}
		p.Areas[iface] = math32.Abs(p.Normals[iface].Dot(&vsum)) / 2
	}
	return p.Areas
}

// colors (must have areas)
func (p *Polyhedron) CalcColors() Vertexes {
	sigfigs := func(f float32, n int) int {
		if f == 0 {
			return 0
		}
		mantissa := f / math32.Pow10(int(math32.Floor(math32.Log10(f))))
		return int(math32.Floor(mantissa * math32.Pow10(n-1)))
	}

	p.Colors = make(Vertexes, len(p.Faces)) // assign p.colors

	color_dict := map[int]Vertex{} // color dictionary
	for iface, a := range p.Areas {
		sf := sigfigs(a, 2)
		if _, ok := color_dict[sf]; !ok { // new color to sf
			color_dict[sf] = *NewVertex(rand.Float32(), rand.Float32(), rand.Float32())
		}
		p.Colors[iface] = color_dict[sf]
	}
	return p.Colors
}

// centers
func (p *Polyhedron) CalcCenters() Vertexes {
	p.Centers = make(Vertexes, len(p.Faces))

	for iface, face := range p.Faces {
		fcenter := Vertex{0, 0, 0}
		for _, v := range face {
			fcenter = *fcenter.Add(&p.Vertexes[v])
		}
		p.Centers[iface] = *fcenter.Mulc(1.0 / float32(len(face)))
	}
	return p.Centers
}

// avg normals
func (p *Polyhedron) AvgNormals() Vertexes {
	avgNorm := make(Vertexes, len(p.Faces))

	for iface, face := range p.Faces {
		fl := len(face)
		var normal_v Vertex
		var v1 = p.Vertexes[face[fl-2]]
		var v2 = p.Vertexes[face[fl-1]]
		for _, v := range face {
			v3 := p.Vertexes[v]
			normal_v = *normal_v.Add(Normal(&v1, &v2, &v3))
			v1, v2 = v2, v3
		}
		avgNorm[iface] = *normal_v.Unit()
	}
	return avgNorm
}

func (p *Polyhedron) ScaleUnit() *Polyhedron {
	mx := float32(math32.MaxFloat32)

	for _, v := range p.Vertexes { // find max abs component of any vertex
		mx = math32.Min(mx, v.MaxAbs())
	}
	if mx != 0 { // scale all vertexes
		for _, v := range p.Vertexes {
			v = *v.Mulc(1.0 / mx)
		}
	}

	return p
}

func (p *Polyhedron) Recalc() *Polyhedron {
	if !p.Check() {
		return p
	}
	p.CalcNormals()
	p.CalcAreas()
	p.CalcColors()
	p.CalcCenters()

	return p
}

func (p *Polyhedron) Clear() *Polyhedron {
	p.Normals = nil
	p.Colors = nil
	p.Centers = nil
	p.Areas = nil
	return p
}

func (p *Polyhedron) Optimize() *Polyhedron { // remove orphan vertexes
	if !p.Check() {
		return p
	}

	MaxFaceIndex := func() uint32 {
		MaxInt := func(a, b uint32) uint32 {
			if a > b {
				return a
			}
			return b
		}
		max := p.Faces[0][0]
		for _, face := range p.Faces {
			for _, ix := range face {
				max = MaxInt(max, ix)
			}
		}
		return max
	}

	old_new := make([]uint32, MaxFaceIndex()+1) // as index range 0..vindex-1
	for i := range old_new {
		old_new[i] = math.MaxUint32
	}

	var nvdx uint32 = 0
	used_vtx := make(Vertexes, 0, len(p.Vertexes))

	for _, face := range p.Faces {
		for _, ix := range face {
			if old_new[ix] == math.MaxUint32 {
				old_new[ix] = nvdx
				used_vtx = append(used_vtx, p.Vertexes[ix])
				nvdx++
			}
		}
	}

	for ix := range p.Faces { // assign faces
		for i := range p.Faces[ix] {
			p.Faces[ix][i] = old_new[p.Faces[ix][i]]
		}
	}
	p.Vertexes = used_vtx

	p.Clear()
	return p.ScaleUnit()
}

func (p *Polyhedron) WriteObj() {
	file, _ := os.Create(p.Name + ".obj")
	defer file.Close()

	file.WriteString("#polyHedronisme http://levskaya.github.com/polyhedronisme\n")

	file.WriteString("group " + p.Name + "\n")
	file.WriteString("#vertices\n")

	for iv, v := range p.Vertexes {
		findIndex := func(iv uint32) int { // find a vertex index in Faces
			for iface, face := range p.Faces {
				if slices.Contains(face, iv) {
					return iface
				}
			}
			return -1
		}

		color := p.Colors[findIndex(uint32(iv))]
		fmt.Fprintf(file, "v %.2f %.2f %.2f %.2f %.2f %.2f\n", v.X, v.Y, v.Z, color.X, color.Y, color.Z)
	}

	// tesselate faces
	trigs := make(Faces, 0, len(p.Faces)*3)
	for _, face := range p.Faces {
		trigs = append(trigs, Tesselate(face)...)
	}

	file.WriteString("#face defs \n")
	for i, face := range trigs {
		fmt.Fprintf(file, "f ")
		for _, v := range face {
			fmt.Fprintf(file, "%d//%d ", v+1, i+1)
		}
		fmt.Fprint(file, "\n")
	}

}

func (p *Polyhedron) Check() bool {
	for _, face := range p.Faces {
		if len(face) < 3 {
			return false
		}
		for _, v := range face {
			if v >= uint32(len(p.Vertexes)) {
				return false
			}
		}
	}
	return true
}

func Unique(vx *Vertexes) *Vertexes {
	res := make(Vertexes, 0)
	contains := func(vx *Vertexes, v_ Vertex) bool {
		for _, v := range *vx {
			if v == v_ {
				return true
			}
		}
		return false
	}
	for _, v := range *vx {
		if !contains(&res, v) {
			res = append(res, v)
		}
	}
	return &res
}

func Tesselate(face Face) Faces {
	pivot := face[0]
	res := Faces{}

	for i := 1; i < len(face)-1; i++ {
		res = append(res, Face{pivot, face[i], face[i+1]})
	}
	return res
}
