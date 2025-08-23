package qh

import "math"

func watermanPolyhedron(radius float64) []float64 {
	estimateCoords := func(rad int) int {		// Linear formula: #coords ≈ 67.9 * rad - 442
		result := 67.9*float64(rad) - 442.0
		return int(math.Round(result))
	}
	a, b, c := 0.0, 0.0, 0.0 // center

	coords := make([]float64, 0, estimateCoords(int(radius))+100)

	s := radius
	radius2 := radius // * radius
	xra := math.Ceil(a - s)
	xrb := math.Floor(a + s)
	zra, zrb := 0.0, 0.0

	for x := xra; x <= xrb; x++ {
		R := radius2 - (x-a)*(x-a)
		if R < 0 {
			continue
		}
		s = math.Sqrt(R)
		yra := math.Ceil(b - s)
		yrb := math.Floor(b + s)
		for y := yra; y <= yrb; y++ {
			Ry := R - (y-b)*(y-b)
			if Ry < 0 {
				continue
			}
			if Ry == 0 && c == math.Floor(c) { //case Ry=0
				if math.Mod(x+y+c, 2) != 0 {
					continue
				} else {
					zra = c
					zrb = c
				}
			} else { // case Ry > 0
				s = math.Sqrt(Ry)
				zra = math.Ceil(c - s)
				zrb = math.Floor(c + s)
				if math.Mod(x+y, 2) == 0 { // (x+y)mod2=0
					if math.Mod(zra, 2) != 0 {
						if zra <= c {
							zra = zra + 1
						} else {
							zra = zra - 1
						}
					}
				} else { // (x+y) mod 2 <> 0
					if math.Mod(zra, 2) == 0 {
						if zra <= c {
							zra = zra + 1
						} else {
							zra = zra - 1
						}
					}
				}
			}

			for z := zra; z <= zrb; z += 2 { // save vertex x,y,z
				coords = append(coords, x)
				coords = append(coords, y)
				coords = append(coords, z)
			}
		}
	}

	// scale coords, by max absolute value
	max := 0.0
	for i := 0; i < len(coords); i++ {
		max=math.Max(max, math.Abs(coords[i]))
	}
	if max != 0 {
		for i := 0; i < len(coords); i++ {
			coords[i] /= max
		}
	}
	return coords
}
