package qh

import (
	"fmt"
	"math"
	"math/rand/v2"
)

const DOUBLE_PREC = 2.2204460492503131e-16

type Vector3d struct {
	x float64
	y float64
	z float64
}

type Point3d = Vector3d

func newVector3d(x, y, z float64) *Vector3d {
	return &Vector3d{x, y, z}
}
func newPoint3d(x, y, z float64) *Point3d {
	return &Point3d{x, y, z}
}

func (v *Vector3d) set(x, y, z float64) {
	v.x = x
	v.y = y
	v.z = z
}
func (v *Vector3d) get(i int) float64 {
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

func (v *Vector3d) add(v1 *Vector3d) {
	v.x += v1.x
	v.y += v1.y
	v.z += v1.z
}
func (v *Vector3d) sub(v1 *Vector3d) {
	v.x -= v1.x
	v.y -= v1.y
	v.z -= v1.z
}
func (v *Vector3d) sub2(v1 *Vector3d, v2 *Vector3d) {
	v.x = v1.x - v2.x
	v.y = v1.y - v2.y
	v.z = v1.z - v2.z
}

func (v *Vector3d) scale(s float64) {
	v.x *= s
	v.y *= s
	v.z *= s
}

func (v *Vector3d) norm() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

func (v *Vector3d) normSquared() float64 {
	return v.x*v.x + v.y*v.y + v.z*v.z
}

func (v *Vector3d) distance(v1 *Vector3d) float64 {
	return math.Sqrt((v.x-v1.x)*(v.x-v1.x) + (v.y-v1.y)*(v.y-v1.y) + (v.z-v1.z)*(v.z-v1.z))
}

func (v *Vector3d) distanceSquared(v1 *Vector3d) float64 {
	return (v.x-v1.x)*(v.x-v1.x) + (v.y-v1.y)*(v.y-v1.y) + (v.z-v1.z)*(v.z-v1.z)
}

func (v *Vector3d) dot(v1 *Vector3d) float64 {
	return v.x*v1.x + v.y*v1.y + v.z*v1.z
}

func (v *Vector3d) normalize() {
	lenSqr := v.x*v.x + v.y*v.y + v.z*v.z
	err := lenSqr - 1
	if err > (2*DOUBLE_PREC) || err < -(2*DOUBLE_PREC) {
		len := math.Sqrt(lenSqr)
		v.x /= len
		v.y /= len
		v.z /= len
	}
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

func (v *Vector3d) setRandom(lower, upper float64) {
	_range := upper - lower

	v.x = rand.Float64()*_range + lower
	v.y = rand.Float64()*_range + lower
	v.z = rand.Float64()*_range + lower
}

func newRand() *Vector3d {
	return newVector3d(rand.Float64(), rand.Float64(), rand.Float64())
}

func (v *Vector3d) String() string {
	return fmt.Sprintf("{%.3f %.3f %.3f}", v.x, v.y, v.z)
}
