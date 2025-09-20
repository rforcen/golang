package polygonizer

import "github.com/chewxy/math32"

type ImplFunc struct {
	Function func(float32, float32, float32) float32
	Name     string
	Bounds   int
	Size     float32
}

var ImplicitFunctions = []ImplFunc{
	{Sphere, "Sphere",
	 60, 0.06}, {Blob, "Blob", 60, 0.06}, {NordstarndWeird, "NordstarndWeird", 60, 0.06}, {DecoCube, "DecoCube", 80, 0.025}, {Cassini, "Cassini", 60, 0.06},
	{Orth, "Orth", 113, 0.016}, {Orth3, "Orth3", 60, 0.06}, {Pretzel, "Pretzel", 135, 0.023}, {Tooth, "Tooth", 60, 0.06}, {Pilz, "Pilz", 94, 0.018},
	{Bretzel, "Bretzel", 70, 0.0152}, {BarthDecic, "BarthDecic", 60, 0.06}, {Clebsch0, "Clebsch0", 60, 0.06}, {Clebsch, "Clebsch", 60, 0.06}, {Chubs, "Chubs", 91, 0.0381},
	{Chair, "Chair", 60, 0.13}, {Roman, "Roman", 60, 0.06}, {TangleCube, "TangleCube", 60, 0.06}, {Goursat, "Goursat", 60, 0.06}, {Sinxyz, "Sinxyz", 60, 0.06}}

// helpers

func sqr(x float32) float32  { return x * x }
func cube(x float32) float32 { return x * x * x }
func pow3(x float32) float32 { return x * x * x }
func pow4(x float32) float32 { return x * x * x * x }
func sphere(x, y, z float32) float32 {
	rsq := x*x + y*y + z*z
	if rsq < 0.00001 {
		rsq = 0.00001
	}
	return 1.0 / rsq
}

// Implicit Funcs

func NordstarndWeird(x, y, z float32) float32 {
	return 25*
		(x*x*x*(y+z)+y*y*y*(x+z)+z*z*z*(x+y)) +
		50*(x*x*y*y+x*x*z*z+y*y*z*z) -
		125*(x*x*y*z+y*y*x*z+z*z*x*y) +
		60*x*y*z - 4*(x*y+y*z+z*x)
}

func DecoCube(x, y, z float32) float32 {
	a := float32(0.95)
	b := float32(0.01)
	return (sqr(x*x+y*y-a*a)+sqr(z*z-1))*
		(sqr(y*y+z*z-a*a)+sqr(x*x-1))*
		(sqr(z*z+x*x-a*a)+sqr(y*y-1)) -
		b
}

func Cassini(x, y, z float32) float32 {
	a := float32(0.3)
	return (sqr((x-a))+z*z)*(sqr((x+a))+z*z) -
		pow4(y) // ( (x-a)^2 + y^2) ((x+a)^2 + y^2) = z^4 a = 0.5
}

func Orth(x, y, z float32) float32 {
	a := float32(0.06)
	b := float32(2.0)
	return (sqr(x*x+y*y-1)+z*z)*(sqr(y*y+z*z-1)+x*x)*
		(sqr(z*z+x*x-1)+y*y) -
		a*a*(1+b*(x*x+y*y+z*z))
}

func Orthogonal(x, y, z float32) float32 {
	// auto (a,b) = (0.06, 2)
	return Orth(x, y, z)
}

func Orth3(x, y, z float32) float32 {
	return 4.0 - Orth(x+0.5, y-0.5, z-0.5) -
		Orth(x-0.5, y+0.5, z-0.5) - Orth(x-0.5, y-0.5, z+0.5)
}

func Pretzel(x, y, z float32) float32 {
	aa := float32(1.6)
	return sqr(((x-1)*(x-1)+y*y-aa*aa)*
		((x+1)*(x+1)+y*y-aa*aa)) +
		z*z*10 - 1
}

func Tooth(x, y, z float32) float32 {
	return pow4(x) + pow4(y) + pow4(z) - sqr(x) - sqr(y) - sqr(z)
}

func Pilz(x, y, z float32) float32 {
	a := float32(0.05)
	b := float32(-0.1)
	return sqr(sqr(x*x+y*y-1)+sqr(z-0.5))*
		(sqr(y*y/a*a+sqr(z+b)-1.0)+x*x) -
		a*(1.0+a*sqr(z-0.5))
}

func Bretzel(x, y, z float32) float32 {
	a := float32(0.003)
	b := float32(0.7)

	return sqr(x*x*(1-x*x)-y*y) + 0.5*z*z -
		a*(1+b*(x*x+y*y+z*z))
}

func BarthDecic(x, y, z float32) float32 {
	GR := float32(1.6180339887) // Golden ratio
	GR2 := GR * GR
	GR4 := GR2 * GR2
	w := float32(0.3)

	return 8*(x*x-GR4*y*y)*(y*y-GR4*z*z)*
		(z*z-GR4*x*x)*
		(x*x*x*x+y*y*y*y+z*z*z*z-
			2*x*x*y*y-2*x*x*z*z-2*y*y*z*z) +
		(3+5*GR)*sqr((x*x+y*y+z*z-w*w))*
			sqr((x*x+y*y+z*z-(2-GR)*w*w))*w*w
}

func Clebsch0(x, y, z float32) float32 {
	return 81*(cube(x)+cube(y)+cube(z)) -
		189*(sqr(x)*y+sqr(x)*z+sqr(y)*x+sqr(y)*z+sqr(z)*x+
			sqr(z)*y) +
		54*(x*y*z) + 126*(x*y+x*z+y*z) -
		9*(sqr(x)+sqr(y)+sqr(z)) - 9*(x+y+z) + 1
}

func Clebsch(x, y, z float32) float32 {
	return 16*cube(x) + 16*cube(y) - 31*cube(z) + 24*sqr(x)*z -
		48*sqr(x)*y - 48*x*sqr(y) + 24*sqr(y)*z -
		54*math32.Sqrt(3.0)*sqr(z) - 72*z
}

func Chubs(x, y, z float32) float32 {
	return pow4(x) + pow4(y) + pow4(z) - sqr(x) - sqr(y) - sqr(z) +
		0.5 // x^4 + y^4 + z^4 - x^2 - y^2 - z^2 + 0.5 = 0
}
func Chair(x, y, z float32) float32 {
	k := float32(5.0)
	a := float32(0.95)
	b := float32(0.8)
	return sqr(sqr(x)+sqr(y)+sqr(z)-a*sqr(k)) -
		b*((sqr((z-k))-2*sqr(x))*(sqr((z+k))-2*sqr(y)))
	// (x^2+y^2+z^2-a*k^2)^2-b*((z-k)^2-2*x^2)*((z+k)^2-2*y^2)=0,
	// with k=5, a=0.95 and b=0.8.
}
func Roman(x, y, z float32) float32 {
	r := float32(2.0)
	return sqr(x)*sqr(y) + sqr(y)*sqr(z) + sqr(z)*sqr(x) - r*x*y*z
}

func Sinxyz(x, y, z float32) float32 { return math32.Sin(x) * math32.Sin(y) * math32.Sin(z) }

func F001(x, y, z float32) float32 { return pow3(x) + pow3(y) + pow4(z) - 10 } // x^3 + y^3 + z^4 -10 = 0

func TangleCube(x, y, z float32) float32 {
	return pow4(x) - 5*sqr(x) + pow4(y) - 5*sqr(y) + pow4(z) - 5*sqr(z) +
		11.8
}

func Goursat(x, y, z float32) float32 { // (x^4 + y^4 + z^4) + a * (x^2 + y^2
	// + z^2)^2 + b * (x^2 + y^2 + z^2) +
	// c = 0
	a := float32(0.0)
	b := float32(0.0)
	c := float32(-1.0)
	return pow4(x) + pow4(y) + pow4(z) + a*sqr(sqr(x)+sqr(y)+sqr(z)) +
		b*(sqr(x)+sqr(y)+sqr(z)) + c
}

func Blob(x, y, z float32) float32 {
	return 4 - sphere(x+0.5, y-0.5, z-0.5) -
		sphere(x-0.5, y+0.5, z-0.5) - sphere(x-0.5, y-0.5, z+0.5)
}

func Sphere(x, y, z float32) float32 { return sphere(x, y, z) - 1 }
