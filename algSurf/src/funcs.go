package algsurf

import "github.com/chewxy/math32"

// surface implementation

type ParamDef struct {
	name                   string
	fromU, toU, fromV, toV float32
	paramFunc              func(u, v float32) Point3d
}

var ParamDefs = []ParamDef{
	{"cap", 0, pi, 0, pi, func(u, v float32) Point3d {
		return Point3d{
			x: 0.5 * math32.Cos(u) * math32.Sin(2*v),
			y: 0.5 * math32.Sin(u) * math32.Sin(2*v),
			z: 0.5 * (math32.Pow(math32.Cos(v), 2) - math32.Pow(math32.Cos(u), 2)*math32.Pow(math32.Sin(v), 2)),
		}
	}},
	{"boy", 0, pi, 0, pi, func(u, v float32) Point3d {
		dv := (2 - math32.Sqrt(2)*math32.Sin(3*u)*math32.Sin(2*v))
		d1 := math32.Cos(u) * math32.Sin(2*v)
		d2 := math32.Sqrt(2) * math32.Pow(math32.Cos(v), 2)

		return Point3d{
			x: (d2*math32.Cos(2*u) + d1) / dv,
			y: (d2*math32.Sin(2*u) + d1) / dv,
			z: (3 * math32.Pow(math32.Cos(v), 2)) / dv,
		}
	}},
	{"roman", 0, 1, 0, twoPi, func(u, v float32) Point3d {
		r2 := u * u
		rq := math32.Sqrt(1 - r2)
		st := math32.Sin(v)
		ct := math32.Cos(v)

		return Point3d{
			x: r2 * st * ct,
			y: u * st * rq,
			z: u * ct * rq,
		}
	}},
	{"sea shell", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		N := float32(5.6) // number of turns
		H := float32(3.5) // height
		P := float32(2)   // power
		L := float32(4)   // Controls spike length
		K := float32(9)

		W := func(u float32) float32 { return math32.Pow(u/(2*pi), P) }

		return Point3d{
			x: W(u) * math32.Cos(N*u) * (1 + math32.Cos(v)),
			y: W(u) * math32.Sin(N*u) * (1 + math32.Cos(v)),
			z: W(u)*(math32.Sin(v)+math32.Pow(math32.Sin(v/2), K)*L) + H*math32.Pow(u/(2*pi), P+1),
		}
	}},
	{"tudor rose", 0, pi, 0, pi, func(u, v float32) Point3d {
		R := func(u, v float32) float32 {
			return math32.Cos(v) * math32.Cos(v) * math32.Max(math32.Abs(math32.Sin(4*u)), 0.9-0.2*math32.Abs(math32.Cos(8*u)))
		}
		return Point3d{
			x: R(u, v) * math32.Cos(u) * math32.Cos(v),
			y: R(u, v) * math32.Sin(u) * math32.Cos(v),
			z: R(u, v) * math32.Sin(v) * 0.5,
		}
	}},
	{"breather", -20, 20, 20, 80, func(u, v float32) Point3d {
		aa := float32(0.45) // Values from 0.4 to 0.6 produce sensible results
		w1 := 1 - aa*aa
		w := math32.Sqrt(w1)

		d := func(u, v float32) float32 {
			return aa * (math32.Pow((w*math32.Cosh(aa*u)), 2) + math32.Pow((aa*math32.Sin(w*v)), 2))
		}
		return Point3d{
			x: -u + (2 * w1 * math32.Cosh(aa*u) * math32.Sinh(aa*u) / d(u, v)),
			y: 2 * w * math32.Cosh(aa*u) * (-(w * math32.Cos(v) * math32.Cos(w*v)) -
				(math32.Sin(v) * math32.Sin(w*v))) / d(u, v),
			z: 2 * w * math32.Cosh(aa*u) * (-(w * math32.Sin(v) * math32.Cos(w*v)) +
				(math32.Cos(v) * math32.Sin(w*v))) / d(u, v),
		}
	}},
	{"klein bottle", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		t := float32(4.5)
		tmp := (4 + 2*math32.Cos(u)*math32.Cos(t*v) - math32.Sin(2*u)*math32.Sin(t*v))
		return Point3d{
			x: math32.Sin(v) * tmp,
			y: math32.Cos(v) * tmp,
			z: 2*math32.Cos(u)*math32.Sin(t*v) + math32.Sin(2*u)*math32.Cos(t*v),
		}
	}},
	{"klein bottle 0", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		triCond := func(b bool, fa, fb float32) float32 {
			if b {
				return fa
			}
			return fb
		}
		return Point3d{
			x: triCond(0 <= u && u < pi, 6*math32.Cos(u)*(1+math32.Sin(u))+4*(1-0.5*math32.Cos(u))*math32.Cos(u)*math32.Cos(v), 6*math32.Cos(u)*(1+math32.Sin(u))+4*(1-0.5*math32.Cos(u))*math32.Cos(v+pi)),
			y: triCond(0 <= u && u < pi, 16*math32.Sin(u)+4*(1-0.5*math32.Cos(u))*math32.Sin(u)*math32.Cos(v), 16*math32.Sin(u)),
			z: 4 * (1 - 0.5*math32.Cos(u)) * math32.Sin(v),
		}
	}},
	{"bour", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		return Point3d{
			x: u*math32.Cos(v) - 0.5*u*u*math32.Cos(2*v),
			y: -u*math32.Sin(v) - 0.5*u*u*math32.Sin(2*v),
			z: 4 / 3 * math32.Pow(u, 1.5) * math32.Cos(1.5*v),
		}
	}},
	{"dini", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		var psi float32 = 0.3 // aa;
		if psi < 0.001 {
			psi = 0.001
		}
		if psi > 0.999 {
			psi = 0.999
		}
		psi = psi * pi
		sinpsi := math32.Sin(psi)
		cospsi := math32.Cos(psi)
		g := (u - cospsi*v) / sinpsi
		s := math32.Exp(g)
		r := (2 * sinpsi) / (s + 1/s)
		t := r * (s - 1/s) * 0.5

		return Point3d{
			x: u - t,
			y: r * math32.Cos(v),
			z: r * math32.Sin(v),
		}
	}},
	{"enneper", -1, 1, -1, 1, func(u, v float32) Point3d {
		return Point3d{
			x: u - u*u*u/3 + u*v*v,
			y: v - v*v*v/3 + v*u*u,
			z: u*u - v*v,
		}
	}},
	{"scherk", 1, 30, 1, 30, func(u, v float32) Point3d {
		var aa float32 = 0.1
		v += 0.1
		return Point3d{
			x: u,
			y: v,
			z: (math32.Log(math32.Abs(math32.Cos(aa*v) / math32.Cos(aa*u)))) / aa,
		}
	}},
	{"conical spiral", 0, 1, -1, 1, func(u, v float32) Point3d {
		return Point3d{
			x: u * v * math32.Sin(15*v),
			y: v,
			z: u * v * math32.Cos(15*v),
		}
	}},
	{"bohemian dome", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		A, B, C := float32(0.5), float32(1.5), float32(1)
		return Point3d{
			x: A * math32.Cos(u),
			y: B*math32.Cos(v) + A*math32.Sin(u),
			z: C * math32.Sin(v),
		}
	}},
	{"astrodial ellipse", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		A, B, C := float32(1), float32(1), float32(1)
		return Point3d{
			x: math32.Pow(A*math32.Cos(u)*math32.Cos(v), 3),
			y: math32.Pow(B*math32.Sin(u)*math32.Cos(v), 3),
			z: math32.Pow(C*math32.Sin(v), 3),
		}
	}},
	{"apple", 0, twoPi, -pi, pi, func(u, v float32) Point3d {
		R1, R2 := float32(4), float32(3.8)
		return Point3d{
			x: math32.Cos(u)*(R1+R2*math32.Cos(v)) + math32.Pow((v/pi), 100),
			y: math32.Sin(u)*(R1+R2*math32.Cos(v)) + 0.25*math32.Cos(5*u),
			z: -2.3*math32.Log(1-v*0.3157) + 6*math32.Sin(v) + 2*math32.Cos(v),
		}
	}},
	{"ammonite", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		N := float32(5.6)   // number of turns
		F := float32(120.0) // wave frequency
		A := float32(0.2)   // wave amplitude
		W := func(u float32) float32 { return math32.Pow(u/(2*pi), 2.2) }

		return Point3d{
			x: W(u) * math32.Cos(N*u) * (2 + math32.Sin(v+math32.Cos(F*u)*A)),
			y: W(u) * math32.Sin(N*u) * (2 + math32.Sin(v+math32.Cos(F*u)*A)),
			z: W(u) * math32.Cos(v),
		}
	}},
	{"plucker comoid", -2, 2, -1, 1, func(u, v float32) Point3d {
		return Point3d{
			x: u * v,
			y: u * math32.Sqrt(1-math32.Pow(v, 2)),
			z: 1 - math32.Pow(v, 2),
		}
	}},
	{"cayley", 0, 3, 0, twoPi, func(u, v float32) Point3d {
		return Point3d{
			x: u*math32.Sin(v) - u*math32.Cos(v),
			y: math32.Pow(u, 2) * math32.Sin(v) * math32.Cos(v),
			z: math32.Pow(u, 3) * math32.Pow(math32.Sin(v), 2) * math32.Cos(v),
		}
	}},
	{"up down shell", -10, 10, -10, 10, func(u, v float32) Point3d {
		return Point3d{
			x: u * math32.Sin(v) * math32.Cos(v),
			y: u * math32.Cos(v) * math32.Cos(v),
			z: u * math32.Sin(v),
		}
	}},
	{"butterfly", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		t1 := (math32.Exp(math32.Cos(u)) - 2*math32.Cos(4*u) + math32.Pow(math32.Sin(u/12), 5)) * math32.Sin(v)
		return Point3d{
			x: math32.Sin(u) * t1,
			y: math32.Cos(u) * t1,
			z: math32.Sin(v),
		}
	}},
	{"rose", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		a := float32(1)
		n := float32(7)
		return Point3d{
			x: a * math32.Sin(n*u) * math32.Cos(u) * math32.Sin(v),
			y: a * math32.Sin(n*u) * math32.Sin(u) * math32.Sin(v),
			z: math32.Cos(v) / (n * 3),
		}
	}},
	{"kuen", -4, 4, -3.75, 3.75, func(u, v float32) Point3d {
		return Point3d{
			x: 2 * math32.Cosh(v) * (math32.Cos(u) + u*math32.Sin(u)) / (math32.Cosh(v)*math32.Cosh(v) + u*u),
			y: 2 * math32.Cosh(v) * (-u*math32.Cos(u) + math32.Sin(u)) /
				(math32.Cosh(v)*math32.Cosh(v) + u*u),
			z: v - (2*math32.Sinh(v)*math32.Cosh(v))/(math32.Cosh(v)*math32.Cosh(v)+u*u),
		}
	}},

	{"tanaka-0", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		a, b1, b2, c, w, h := float32(0), float32(4), float32(3), float32(4), float32(7), float32(4) // center hole size of a torus
		return Point3d{
			x: (a - math32.Cos(v) + w*math32.Sin(b1*u)) * math32.Cos(b2*u),
			y: (a - math32.Cos(v) + w*math32.Sin(b1*u)) * math32.Sin(b2*u),
			z: h*(w*math32.Sin(b1*u)+math32.Sin(v)) + c,
		}
	}},
	{"tanaka-1", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		a, b1, b2, c, w, h := float32(0), float32(4), float32(3), float32(0), float32(7), float32(4) // center hole size of a torus
		return Point3d{
			x: (a - math32.Cos(v) + w*math32.Sin(b1*u)) * math32.Cos(b2*u),
			y: (a - math32.Cos(v) + w*math32.Sin(b1*u)) * math32.Sin(b2*u),
			z: h*(w*math32.Sin(b1*u)+math32.Sin(v)) + c,
		}
	}},
	{"tanaka-2", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		a, b1, b2, c, w, h := float32(0), float32(3), float32(4), float32(8), float32(5), float32(2) // center hole size of a torus
		return Point3d{
			x: (a - math32.Cos(v) + w*math32.Sin(b1*u)) * math32.Cos(b2*u),
			y: (a - math32.Cos(v) + w*math32.Sin(b1*u)) * math32.Sin(b2*u),
			z: h*(w*math32.Sin(b1*u)+math32.Sin(v)) + c,
		}
	}},
	{"tanaka-3", 0, twoPi, 0, twoPi, func(u, v float32) Point3d {
		a, b1, b2, c, w, h := float32(14), float32(3), float32(1), float32(8), float32(5), float32(2) // center hole size of a torus
		return Point3d{
			x: (a - math32.Cos(v) + w*math32.Sin(b1*u)) * math32.Cos(b2*u),
			y: (a - math32.Cos(v) + w*math32.Sin(b1*u)) * math32.Sin(b2*u),
			z: h*(w*math32.Sin(b1*u)+math32.Sin(v)) + c,
		}
	}},
}
