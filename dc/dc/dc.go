package dc

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"math"
	"math/cmplx"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	pi2 = math.Pi * 2
)

var Presets = []string{
	"acos((1+i)*log(sin(z^3-1)/z))",
	"(1+i)*log(sin(z^3-1)/z)",
	"(1+i)*sin(z)",
	"z + z^2/sin(z^4-1)",
	"log(sin(z))",
	"cos(z)/(sin(z^4-1)",
	"z^6-1",
	"(z^2-1) * (z-2-i)^2 / (z^2+2*i)",
	"sin(z)*c(1,2)",
	"sin(1/z)",
	"sin(z)*sin(1/z)",
	"1/sin(1/sin(z))",
	"z",
	"(z^2+1)/(z^2-1)",
	"(z^2+1)/z",
	"(z+3)*(z+1)^2",
	"(z/2)^2*(z+1-2i)*(z+2+2i)/z^3",
	"(z^2)-0.75-(0.2*i)",
}

type DC struct {
	z_comp     ZCompiler
	w          int
	h          int
	size       int
	expression string
	image      []uint32
	Lap        float64
}

func NewDC(w int, h int, expression string) DC {
	dc_ := DC{
		w:          w,
		h:          h,
		size:       w * h,
		expression: expression,
		z_comp:     NewCompiler(expression),
		image:      make([]uint32, w*h),
	}
	return dc_
}

func (dc_ *DC) genPixel(th, index_ int) {
	pow3 := func(x float64) float64 { return x * x * x }

	limit := math.Pi

	rmi, rma, imi, ima := -limit, limit, -limit, limit

	x, y := math.Mod(float64(index_), float64(dc_.w)), float64(index_)/float64(dc_.w)

	// map pixel to complex plane
	z := complex(float64(rmi+(rma-rmi)*x/float64(dc_.w)), float64(imi+(ima-imi)*y/float64(dc_.h)))

	// execute
	var result complex128
	result = dc_.z_comp.execute(z)

	// convert result to color
	hue, m := cmplx.Phase(result), cmplx.Abs(result)
	hue = math.Mod(math.Mod(hue, pi2)+pi2, pi2) / pi2

	ranges, rangee := 0.0, 1.0
	for m > rangee {
		ranges = rangee
		rangee *= math.E
	}

	k := (m - ranges) / (rangee - ranges)
	var kk float64
	if k < 0.5 {
		kk = k * 2
	} else {
		kk = 1 - (k-0.5)*2
	}

	sat := 0.4 + (1-pow3(1-kk))*0.6
	val := 0.6 + (1-pow3(1-(1-kk)))*0.4

	dc_.image[index_] = hsv_2_rgb(hue, sat, val)
}

func hsv_2_rgb(h float64, s float64, v float64) uint32 {
	r, g, b := 0.0, 0.0, 0.0

	if s == 0 {
		r, g, b = v, v, v
	} else {
		if h == 1 {
			h = 0
		}

		z := math.Floor(h * 6)
		i, f := int(z), h*6-z
		p, q, t := v*(1-s), v*(1-s*f), v*(1-s*(1-f))

		switch i {
		case 0:
			r, g, b = v, t, p
		case 1:
			r, g, b = q, v, p
		case 2:
			r, g, b = p, v, t
		case 3:
			r, g, b = p, q, v
		case 4:
			r, g, b = t, p, v
		case 5:
			r, g, b = v, p, q
		}
	}
	return 0xff000000 | uint32(r*255)<<16 | uint32(g*255)<<8 | uint32(b*255)
}

func (dc_ *DC) GenImageSt() {
	t0 := time.Now()
	for index_ := 0; index_ < dc_.size; index_++ {
		dc_.genPixel(0, index_)
	}
	dc_.Lap = float64(time.Since(t0).Milliseconds())
}

func (dc_ *DC) GenImageMt() {
	t0 := time.Now()

	numCores := runtime.NumCPU()
	itemsPerCore := dc_.size / numCores

	var wg sync.WaitGroup
	wg.Add(numCores)

	for th := range numCores {
		go func(th int) {
			defer wg.Done()
			for index := th * itemsPerCore; index < min((th+1)*itemsPerCore, dc_.size); index++ {
				dc_.genPixel(th, index)
			}
		}(th)
	}
	wg.Wait()

	dc_.Lap = float64(time.Since(t0).Milliseconds())
}

func (dc_ *DC) WriteImage(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close() // Ensure the file is closed when the function exits

	byteOrder := binary.LittleEndian // Or binary.BigEndian

	err = binary.Write(file, byteOrder, dc_.image)
	if err != nil {
		fmt.Println("Error writing data:", err)
		return
	}
}

func (dc_ *DC) GenerateImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, dc_.w, dc_.h))

	for i, argbPixel := range dc_.image {
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

func (dc_ *DC) WritePng(filename string) error {
	img := dc_.GenerateImage()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	return nil
}

func (dc_ *DC) Random(complexity int) {
	// dc_.z_comp = GenRandomExpression(complexity) // old school way
	dc_.z_comp = NewCompiler(GenRandom(complexity))
	dc_.GenImageMt()
}

func (dc_ *DC) GetExpression() string {
	return dc_.z_comp.expr
}
