package sh

import "github.com/chewxy/math32"

// Coord
type Coord struct {
	X float32
	Y float32
	Z float32
}

func (c *Coord) Add(other *Coord) *Coord {
	return &Coord{c.X + other.X, c.Y + other.Y, c.Z + other.Z}
}

func (c *Coord) Sub(other *Coord) *Coord {
	return &Coord{c.X - other.X, c.Y - other.Y, c.Z - other.Z}
}

func (c *Coord) Mul(other *Coord) *Coord {
	return &Coord{c.X * other.X, c.Y * other.Y, c.Z * other.Z}
}
func (c *Coord) MulScalar(other float32) *Coord {
	return &Coord{c.X * other, c.Y * other, c.Z * other}
}

func (c *Coord) Div(other *Coord) *Coord {
	return &Coord{c.X / other.X, c.Y / other.Y, c.Z / other.Z}
}

func (c *Coord) Dot(other *Coord) float32 {
	return c.X*other.X + c.Y*other.Y + c.Z*other.Z
}

func (c *Coord) Cross(other *Coord) *Coord {
	return &Coord{c.Y*other.Z - c.Z*other.Y, c.Z*other.X - c.X*other.Z, c.X*other.Y - c.Y*other.X}
}

func (c *Coord) Norm() float32 {
	return math32.Sqrt(c.X*c.X + c.Y*c.Y + c.Z*c.Z)
}

func (c *Coord) Normalize() *Coord {
	return c.Div(&Coord{c.Norm(), c.Norm(), c.Norm()})
}

func Normal(v0 *Coord, v1 *Coord, v2 *Coord) *Coord {
	return v1.Sub(v0).Cross(v2.Sub(v0)).Normalize()
}

func NewCoord(x float32, y float32, z float32) Coord {
	return Coord{X: x, Y: y, Z: z}
}
