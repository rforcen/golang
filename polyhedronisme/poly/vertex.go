package poly

import "github.com/chewxy/math32"

type Vertex struct {
	X float32
	Y float32
	Z float32
}

func NewVertex(x, y, z float32) *Vertex {
	return &Vertex{
		X: x,
		Y: y,
		Z: z,
	}
}

func (v *Vertex) Add(other *Vertex) *Vertex {
	return &Vertex{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

func (v *Vertex) Sub(other *Vertex) *Vertex {
	return &Vertex{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

func (v *Vertex) Mulc(s float32) *Vertex {
	return &Vertex{
		X: v.X * s,
		Y: v.Y * s,
		Z: v.Z * s,
	}
}

func (v *Vertex) Length() float32 {
	return math32.Sqrt(v.NormSquared())
}

func (v *Vertex) Unit() *Vertex {
	l := v.Length()
	if l == 0 {
		return v
	}
	return v.Mulc(1 / l)
}

func (v *Vertex) Cross(other *Vertex) *Vertex {
	return &Vertex{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

func (v *Vertex) Dot(other *Vertex) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v *Vertex) Norm() float32 {
	return math32.Sqrt(v.NormSquared())
}

func (v *Vertex) NormSquared() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v *Vertex) MaxAbs() float32 {
	return math32.Max(math32.Max(math32.Abs(v.X), math32.Abs(v.Y)), math32.Abs(v.Z))
}

func Normal(v0 *Vertex, v1 *Vertex, v2 *Vertex) *Vertex {
	return v1.Sub(v0).Cross(v2.Sub(v0)).Unit()
}

// Vertex helpers
func Tween(vec1 *Vertex, vec2 *Vertex, t float32) *Vertex {
	return vec1.Mulc(1 - t).Add(vec2.Mulc(t))
}

func OneThird(vec1 *Vertex, vec2 *Vertex) *Vertex {
	return Tween(vec1, vec2, 1.0/3)
}
