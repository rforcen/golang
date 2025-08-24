package qh

import (
	"fmt"
	"math/rand/v2"

	"github.com/chewxy/math32"
)

const FLOAT_PREC = float32(1.1920929e-7) // float32, machine epsilon

type Vector3d struct {
	x, y, z float32
}

type Point3d = Vector3d

func newVector3d(x, y, z float32) *Vector3d {
	return &Vector3d{x, y, z}
}
func newPoint3d(x, y, z float32) *Point3d {
	return &Point3d{x, y, z}
}
func zeroPoint3d() *Point3d {
	return &Point3d{0, 0, 0}
}
func (v *Vector3d) get(i int) float32 {
	switch i {
	case 0:
		return v.x
	case 1:
		return v.y
	case 2:
		return v.z
	default:
		panic("index out of range")
	}
}
func (v *Vector3d) setPoint(p *Point3d) {
	v.x = p.x
	v.y = p.y
	v.z = p.z
}

func (v *Vector3d) accumulate(v1 *Vector3d) {
	v.x += v1.x
	v.y += v1.y
	v.z += v1.z
}

func sub(v1, v2 *Vector3d) *Vector3d {
	return &Vector3d{v1.x - v2.x, v1.y - v2.y, v1.z - v2.z}
}

func (v *Vector3d) sub2(v1 *Vector3d, v2 *Vector3d) {
	v.x = v1.x - v2.x
	v.y = v1.y - v2.y
	v.z = v1.z - v2.z
}

func (v *Vector3d) scale(s float32) {
	v.x *= s
	v.y *= s
	v.z *= s
}

func (v *Vector3d) norm() float32 {
	return math32.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

func (v *Vector3d) normSquared() float32 {
	return v.x*v.x + v.y*v.y + v.z*v.z
}

func (v *Vector3d) dot(v1 *Vector3d) float32 {
	return v.x*v1.x + v.y*v1.y + v.z*v1.z
}

func normal(v0 *Vector3d, v1 *Vector3d, v2 *Vector3d) *Vector3d {
	return crossRet(sub(v1, v0), sub(v2, v0)).normalizeRet()
}

func (v *Vector3d) normalize() {
	lenSqr := v.x*v.x + v.y*v.y + v.z*v.z
	err := lenSqr - 1
	if err > (2*FLOAT_PREC) || err < -(2*FLOAT_PREC) {
		len := math32.Sqrt(lenSqr)
		v.x /= len
		v.y /= len
		v.z /= len
	}
}
func (v *Vector3d) normalizeRet() *Vector3d {
	lenSqr := v.x*v.x + v.y*v.y + v.z*v.z
	err := lenSqr - 1
	vt := *v
	if err > (2*FLOAT_PREC) || err < -(2*FLOAT_PREC) {
		len := math32.Sqrt(lenSqr)
		vt.x /= len
		vt.y /= len
		vt.z /= len
	}
	return &vt
}

func (v *Vector3d) setZero() {
	v.x = 0
	v.y = 0
	v.z = 0
}

func (v *Vector3d) cross(v1 *Vector3d, v2 *Vector3d) {
	v.x = v1.y*v2.z - v1.z*v2.y
	v.y = v1.z*v2.x - v1.x*v2.z
	v.z = v1.x*v2.y - v1.y*v2.x
}
func crossRet(v1 *Vector3d, v2 *Vector3d) *Vector3d {
	return &Vector3d{v1.y*v2.z - v1.z*v2.y, v1.z*v2.x - v1.x*v2.z, v1.x*v2.y - v1.y*v2.x}
}

func (v *Vector3d) setRandom(lower, upper float32) {
	_range := upper - lower

	v.x = rand.Float32()*_range + lower
	v.y = rand.Float32()*_range + lower
	v.z = rand.Float32()*_range + lower
}

func (v *Vector3d) String() string {
	return fmt.Sprintf("{%.3f %.3f %.3f}", v.x, v.y, v.z)
}
