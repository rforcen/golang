// voronoi tiles

package voronoi

import (
	"image"
	"math"
	"math/rand/v2"
	"runtime"
	"sync"
)

const pointSize = 3

type Point struct {
	X, Y int
	Color uint32
}

type Voronoi struct {
	w, h, size int
	points []Point
	image []uint32
	single_thread bool
}

func NewVoronoi(w int, h int, n int, single_thread bool) *Voronoi {
	v :=  Voronoi{
		w:             w,
		h:             h,
		size:          w * h,
		points:        createRandomPoints(w, h, n),
		image:         make([]uint32, w * h),
		single_thread: single_thread,
	}

	if single_thread {
		v.genImageST()
	} else {
		v.genImageMT()
	}	
	return &v
}


func (v *Voronoi) genPixel(index int) uint32 {
	dist_sq := func(i int, j int, p Point) int {
		xd := i - p.X
		yd := j - p.Y
		return xd * xd + yd * yd
	}
	
	i := index % v.w
	j := index / v.w

	min_dist := math.MaxInt
	color := uint32(0xff000000)

	for _, p := range v.points {
		dist := dist_sq(i, j, p)
		if dist < pointSize {
			color = 0xff000000
			break
		}
		if dist < min_dist {
			min_dist = dist
			color = p.Color
		}
	}
	return color
}


func (v *Voronoi) genImageST() {
	for i := range v.image {
		v.image[i] = v.genPixel(i)
	}
}

func (v *Voronoi) genImageMT() {
	numCores := runtime.NumCPU()
	itemsPerCore := v.size / numCores

	var wg sync.WaitGroup
	wg.Add(numCores)

	for th := range numCores {
		go func() {
			for index := th * itemsPerCore; index < min((th+1)*itemsPerCore, v.size); index++ {
				v.image[index] = v.genPixel(index)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func createRandomPoints(w int, h int, n int) []Point {
	points := []Point{}
	for range n {
		points = append(points, Point{
			X:     rand.IntN(w),
			Y:     rand.IntN(h),
			Color: uint32(0xff000000 | rand.IntN(0x00ffffff)),
		})
	}
	return points	
}

func (v *Voronoi) GenerateImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, v.w, v.h))

	for i, argbPixel := range v.image {
		// The offset into the img.Pix slice is directly tied to the index 'i'
		// since each pixel takes 4 bytes (R, G, B, A).
		offset := i * 4

		// ABGR to RGBA conversion
		img.Pix[offset+2] = uint8((argbPixel >> 16) & 0xFF) // Red - swap w/Blue
		img.Pix[offset+1] = uint8((argbPixel >> 8) & 0xFF)  // Green
		img.Pix[offset+0] = uint8(argbPixel & 0xFF)         // Blue
		img.Pix[offset+3] = uint8((argbPixel >> 24) & 0xFF) // Alpha
	}

	return img
}
