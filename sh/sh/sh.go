package sh

import (
	"bufio"
	"fmt"
	"github.com/chewxy/math32"
	"os"
	"runtime"
	"sync"
)

// Location, mesh item
type Location struct {
	Coord
	Normal Coord
	Color  Coord
	UV     Coord
}

// SH, spherical harmonics
type SH struct {
	Mesh         []Location
	Faces        [][]int
	Res          int
	ColorMap     int
	Size         int
	Code         int
	M            []float32
	Du           float32
	Dv           float32
	Du10         float32
	Dv10         float32
	Dx           float32
	MaxVal       float32
	SingleThread bool
}

func NewSH(res int, color_map int, code int) *SH {
	du := 2 * math32.Pi / float32(res)
	dv := math32.Pi / float32(res)

	return &SH{
		Res:          res,
		ColorMap:     color_map,
		Code:         code,
		Size:         res * res,
		M:            ToFloatv(code),
		Du:           du,
		Dv:           dv,
		Du10:         du / 10.0,
		Dv10:         dv / 10.0,
		Dx:           1.0 / float32(res),
		MaxVal:       -1.0,
		SingleThread: false,
	}
}

func (sh *SH) calcCoord(theta float32, phi float32) Coord {
	sin_phi := math32.Sin(phi)

	r := math32.Pow(math32.Sin(sh.M[0]*phi), sh.M[1])
	r += math32.Pow(math32.Cos(sh.M[2]*phi), sh.M[3])
	r += math32.Pow(math32.Sin(sh.M[4]*theta), sh.M[5])
	r += math32.Pow(math32.Cos(sh.M[6]*theta), sh.M[7])

	return Coord{X: r * sin_phi * math32.Cos(theta), Y: r * math32.Cos(phi), Z: r * sin_phi * math32.Sin(theta)}
}

func (sh *SH) calcLocation(i int, j int) Location {
	u := sh.Du * float32(i)
	v := sh.Dv * float32(j)

	idx := float32(i) * sh.Dx
	jdx := float32(j) * sh.Dx

	coord := sh.calcCoord(u, v)
	crd_up := sh.calcCoord(u+sh.Du10, v)
	crd_right := sh.calcCoord(u, v+sh.Dv10)

	sh.MaxVal = math32.Max(sh.MaxVal, math32.Max(math32.Abs(coord.X), math32.Max(math32.Abs(coord.Y), math32.Abs(coord.Z)))) //  semaphored?

	return Location{
		Coord:  coord,
		Normal: *Normal(&coord, &crd_up, &crd_right),
		Color:  ColorMap(u, 0, math32.Pi*2, sh.ColorMap),
		UV:     Coord{X: idx, Y: jdx, Z: 0.0},
	}
}

func (sh *SH) ScaleCoords() {
	if sh.MaxVal != 0.0 {
		for i := 0; i < sh.Size; i++ {
			sh.Mesh[i].Coord = *sh.Mesh[i].Coord.Div(&Coord{sh.MaxVal, sh.MaxVal, sh.MaxVal})
		}
	}
}

func (sh *SH) GenerateFaces() {
	n := sh.Res
	sh.Faces = make([][]int, 0)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-1; j++ {
			sh.Faces = append(sh.Faces, []int{i*n + j, (i+1)*n + j, (i+1)*n + j + 1, i*n + j + 1})
		}
		sh.Faces = append(sh.Faces, []int{i*n + (n - 1), (i+1)*n + (n - 1), (i + 1) * n, i * n})
	}
}

func (sh *SH) CalcMesh() {
	sh.Mesh = make([]Location, sh.Size)

	for i := 0; i < sh.Size; i++ {
		sh.Mesh[i] = sh.calcLocation(i%sh.Res, i/sh.Res)
	}

	sh.ScaleCoords()
	sh.GenerateFaces()
}
func (sh *SH) CalcMeshMt() {
	sh.Mesh = make([]Location, sh.Size)

	numCores := runtime.NumCPU()
	itemsPerCore := sh.Size / numCores

	var wg sync.WaitGroup
	wg.Add(numCores)

	for th := range numCores {
		go func(th int) {
			defer wg.Done()
			for index := th * itemsPerCore; index < min((th+1)*itemsPerCore, sh.Size); index++ {
				sh.Mesh[index] = sh.calcLocation(index%sh.Res, index/sh.Res)
			}
		}(th)
	}
	wg.Wait()

	sh.ScaleCoords()
	sh.GenerateFaces()
}

func (sh *SH) WriteObj(fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	fmt.Fprintln(writer, "# Spherical Harmonics, code:", sh.Code, "res:", sh.Res, "color_map:", sh.ColorMap)
	for _, loc := range sh.Mesh {
		fmt.Fprintln(writer, "v", fmt.Sprintf("%.3f %.3f %.3f %.3f %.3f %.3f", loc.Coord.X, loc.Coord.Y, loc.Coord.Z, loc.Color.X, loc.Color.Y, loc.Color.Z))
	}
	for _, face := range sh.Faces {
		fmt.Fprintln(writer, "f", face[0]+1, face[1]+1, face[2]+1, face[3]+1)
	}
	writer.Flush()
}
