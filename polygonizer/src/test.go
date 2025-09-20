package polygonizer

import (
	"fmt"
	"time"
)

func Test01() {
	fmt.Println("test01")
	p0 := POINT{1, 2, 3}
	fmt.Println(p0)

}

func Test02() {
	fmt.Println("test02")

	for range 1000 {
		t0 := time.Now()

		vertexes, triangles, msg := Polygonize(ImplicitFunctions[0].Function, 0.06, 60)

		fmt.Printf("polygonize, lap:%dms, msg:%s, %5d vertices, %5d triangles\n", time.Since(t0).Milliseconds(), msg, len(vertexes), len(triangles))

		// check triangles
		if len(triangles) > 0 && len(vertexes) > 0 {
			triangles = triangles[:10]
		} else {
			panic("no vertexes")
		}
		fmt.Println("check triangles", triangles)
		for _, t := range triangles {
			if t.i1 < 0 || t.i2 < 0 || t.i3 < 0 || t.i1 >= len(vertexes) || t.i2 >= len(vertexes) || t.i3 >= len(vertexes) {
				fmt.Println(t)
			}
		}
	}
}
