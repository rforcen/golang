package algsurf

import (
	"runtime"
	"sync"

	"github.com/chewxy/math32"
)

// simple usage
func ParamFuncCoords(res int, fromU, toU, fromV, toV float32, paramFunc func(u, v float32) Point3d) (scale float32, coords []Point3d, normals []Point3d, textures []Point2d) {
	ps := ParametricSurface{res:res, ParametricFunc: paramFunc, scaled:false, fscale:1, fromU:fromU, toU:toU, fromV:fromV, toV:toV, coords: make([]Point3d, res*res*4), normals: make([]Point3d, res*res), textures: make([]Point2d, res*res*4), difU: math32.Abs(fromU - toU), difV: math32.Abs(fromV - toV), MaxVal: -math32.MaxFloat32, MinVal: math32.MaxFloat32}
	ps.calcCoordsMT()
	return ps.fscale, ps.coords, ps.normals, ps.textures
}

const (
	pi    = math32.Pi
	twoPi = math32.Pi * 2
)

type Point3d struct {
	x float32
	y float32
	z float32
}

func (p *Point3d) Scale(scale float32) *Point3d {
	p.x *= scale
	p.y *= scale
	p.z *= scale
	return p
}

type Point2d struct {
	x float32
	y float32
}

type ParametricSurface struct {
	res int

	fromU, toU, fromV, toV, difU, difV float32
	MaxVal, MinVal, Dif                float32

	scaled bool

	p               Point3d
	fscale          float32
	coords, normals []Point3d
	textures        []Point2d

	ParametricFunc func(u, v float32) Point3d
}

// helpers
func ternary(cnd bool, a, b float32) float32 { // ? 'c lang.' operator  cond ? a : b
	if cnd {
		return a
	}
	return b
}

// calc normal of ui.coords[i:i+4]
func Cross(a, b, c Point3d) Point3d {
	return Point3d{
		x: a.y*(b.z-c.z) + b.y*(c.z-a.z) + c.y*(a.z-b.z),
		y: a.z*(b.x-c.x) + b.z*(c.x-a.x) + c.z*(a.x-b.x),
		z: a.x*(b.y-c.y) + b.x*(c.y-a.y) + c.x*(a.y-b.y),
	}
}
func Normalize(p Point3d) Point3d {
	l := math32.Sqrt(p.x*p.x + p.y*p.y + p.z*p.z)
	if l == 0 {
		l = 1
	}
	return Point3d{
		x: p.x / l,
		y: p.y / l,
		z: p.z / l,
	}
}

func max(x, y float32) float32 { return ternary(x > y, x, y) }
func sqr(x float32) float32    { return x * x }
func cube(x float32) float32   { return x * x * x }
func sqr3(x float32) float32   { return x * x * x }
func sqr4(x float32) float32   { return x * x * x * x }
func sqr5(x float32) float32   { return x * x * x * x * x }


func (ps *ParametricSurface) setResol(resol int)            { ps.res = resol }
func (ps *ParametricSurface) rad2Deg(rad float32) float32   { return rad * 180.0 / math32.Pi }
func (ps *ParametricSurface) scaleU(val float32) float32    { return val*ps.difU + ps.fromU }
func (ps *ParametricSurface) scaleV(val float32) float32    { return val*ps.difV + ps.fromV }
func (ps *ParametricSurface) scale01(val float32) float32   { return val / ps.Dif }               // keep center
func (ps *ParametricSurface) scale0to1(val float32) float32 { return (val - ps.MinVal) / ps.Dif } // scale 0..1
func (ps *ParametricSurface) scale() {
	ps.p.x = ps.scale01(ps.p.x)
	ps.p.y = ps.scale01(ps.p.y)
	ps.p.z = ps.scale01(ps.p.z)
}

func (ps *ParametricSurface) Eval(u, v float32) Point3d {
	return ps.ParametricFunc(ps.scaleU(u), ps.scaleV(v))
}

func (ps *ParametricSurface) minMaxP() { // update min/max in p
	ps.MaxVal = math32.Max(ps.MaxVal, math32.Max(ps.p.x, math32.Max(ps.p.y, ps.p.z)))
	ps.MinVal = math32.Min(ps.MinVal, math32.Min(ps.p.x, math32.Min(ps.p.y, ps.p.z)))
}

func (ps *ParametricSurface) initMinMax() {
	ps.MaxVal = -math32.MaxFloat32
	ps.MinVal = math32.MaxFloat32
}

func (ps *ParametricSurface) calcDif() {
	ps.Dif = math32.Abs(ps.MaxVal - ps.MinVal)
}

func (ps *ParametricSurface) addTextVertex(tx, ty float32) { // add textures coords and vertex
	ps.p = ps.Eval(tx, ty)
	ps.coords = append(ps.coords, *ps.p.Scale(ps.fscale))
	ps.textures = append(ps.textures, Point2d{tx, ty})
	

	ps.minMaxP()
}
func (ps *ParametricSurface) addTextVertexMT(index int, tx, ty float32) { // add textures coords and vertex
	ps.p = ps.Eval(tx, ty)
	ps.coords[index] = *ps.p.Scale(ps.fscale)
	ps.textures[index] = Point2d{tx, ty}	
}

func (ps *ParametricSurface) scaleCoords() {
	for _, p := range ps.coords {
		ps.p = p
		ps.minMaxP()
	}
	ps.calcDif()
	ps.fscale = ternary(ps.Dif == 0, 1, 1.0/ps.Dif) // autoscale
}

func (ps *ParametricSurface) calcCoords(resol int, fromU, toU, fromV, toV float32) (coords []Point3d, normals []Point3d, textures []Point2d) { // calc & load

	ps.coords = make([]Point3d, 0, resol*resol*4)
	ps.normals = make([]Point3d, 0, resol*resol*4)
	ps.textures = make([]Point2d, 0, resol*resol*4)

	ps.fromU = fromU
	ps.toU = toU
	ps.fromV = fromV
	ps.toV = toV // define limits
	ps.difU, ps.difV = math32.Abs(ps.fromU-ps.toU) , math32.Abs(ps.fromV-ps.toV) / 1e4

	ps.setResol(resol)
	ps.initMinMax()

	dr, dt := 1.0 / float32(ps.res), 1.0 / float32(ps.res)

	for i := range ps.res { // generated res*res*4 coords -> QUADS
		idr := float32(i) * dr
		for j := range ps.res {
			jdt := float32(j) * dt
			jdr := jdt
			{
				ps.addTextVertex(idr, jdr)
				ps.addTextVertex(idr+dr, jdt)
				ps.addTextVertex(idr+dr, jdt+dt)
				ps.addTextVertex(idr, jdt+dt)
			}
			lc:=len(ps.coords)
			ps.normals = append(ps.normals, Normalize(Cross(ps.coords[lc-4], ps.coords[lc-3], ps.coords[lc-2])))
		}
	}
	ps.calcDif()
	ps.fscale = ternary(ps.Dif == 0, 1, 1.0/ps.Dif) // autoscale

	return ps.coords, ps.normals, ps.textures
}
func (ps *ParametricSurface) calcCoordsMT() { // calc & load

	dr, dt := 1.0/float32(ps.res), 1.0/float32(ps.res)

	nth := runtime.NumCPU()
	wg := sync.WaitGroup{}
	wg.Add(nth)

	for th := range nth {
		go func(th int) {
			defer wg.Done()
			for i := th; i < ps.res; i += nth { // generated res*res*4 coords -> QUADS
				idr := float32(i) * dr
				for j := range ps.res {
					jdt := float32(j) * dt
					jdr := jdt
					{
						ps.addTextVertexMT(4*(i*ps.res+j), idr, jdr)
						ps.addTextVertexMT(4*(i*ps.res+j)+1, idr+dr, jdt)
						ps.addTextVertexMT(4*(i*ps.res+j)+2, idr+dr, jdt+dt)
						ps.addTextVertexMT(4*(i*ps.res+j)+3, idr, jdt+dt)

						ps.normals[i*ps.res+j] = Normalize(Cross(ps.Eval(idr, jdr), ps.Eval(idr+dr, jdt), ps.Eval(idr+dr, jdt+dt)))						
					}
				}
			}
		}(th)
	}
	wg.Wait()

	ps.scaleCoords()
}
