package algsurf

import (
	"fmt"
	"time"
)

func TestAllFuncs() {
	res := 512*2
	fmt.Println("res:", res)

	for j := range 5 {
		st := 0.0
		for i, r := range ParamDefs {
			t0 := time.Now()
			_, coords, normals, textures := ParamFuncCoords(res, r.fromU, r.toU, r.fromV, r.toV, r.paramFunc)
			st += float64(time.Since(t0)) / 1e6
			if j == 0 {
				fmt.Printf("%3d %-20s %5.1f ms, ", i, r.name, float64(time.Since(t0)) / 1e6)
				fmt.Println(coords[:3], normals[:3], textures[:3])
			}
		}
		fmt.Printf("%d, %5.1f ms\n", j, st)
	}
}
