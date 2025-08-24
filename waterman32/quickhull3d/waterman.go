package qh

import "github.com/chewxy/math32"

func watermanCoords(radius float32) []float32 {
	estimateCoords := func(rad int) int { // estimate # coords, Linear formula: #coords ≈ 67.9 * rad - 442
		return int(math32.Abs(math32.Round(67.9*float32(rad) - 442.0)))
	}
	zf32 := float32(0)
	a, b, c := zf32, zf32, zf32 // center

	coords := make([]float32, 0, estimateCoords(int(radius))+100)

	s := radius
	
	xra, xrb := math32.Ceil(a-s), math32.Floor(a+s)
	zra, zrb := zf32, zf32

	for x := xra; x <= xrb; x++ {
		R := radius - (x-a)*(x-a)
		if R < 0 {
			continue
		}
		s = math32.Sqrt(R)
		yra, yrb := math32.Ceil(b-s), math32.Floor(b+s)
		for y := yra; y <= yrb; y++ {
			Ry := R - (y-b)*(y-b)
			if Ry < 0 {
				continue
			}
			if Ry == 0 && c == math32.Floor(c) { //case Ry=0
				if math32.Mod(x+y+c, 2) != 0 {
					continue
				} else {
					zra = c
					zrb = c
				}
			} else { // case Ry > 0
				s = math32.Sqrt(Ry)
				zra = math32.Ceil(c - s)
				zrb = math32.Floor(c + s)
				if math32.Mod(x+y, 2) == 0 { // (x+y)mod2=0
					if math32.Mod(zra, 2) != 0 {
						if zra <= c {
							zra = zra + 1
						} else {
							zra = zra - 1
						}
					}
				} else { // (x+y) mod 2 <> 0
					if math32.Mod(zra, 2) == 0 {
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
	max := zf32
	for i := 0; i < len(coords); i++ {
		max = math32.Max(max, math32.Abs(coords[i]))
	}
	if max != 0 {
		for i := 0; i < len(coords); i++ {
			coords[i] /= max
		}
	}
	return coords
}
