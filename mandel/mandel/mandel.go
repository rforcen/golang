// mandelbrot fractals

package mandel

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"runtime"
	"sync"
	"time"
)

var fire_pallete_256 = []uint32{0, 0, 4, 12, 16, 24, 32, 36, 44, 48, 56, 64, 68, 76, 80, 88, 96,
	100, 108, 116, 120, 128, 132, 140, 148, 152, 160, 164, 172, 180, 184, 192, 200, 1224, 3272,
	4300, 6348, 7376, 9424, 10448, 12500, 14548, 15576, 17624, 18648, 20700, 21724, 23776, 25824,
	26848, 28900, 29924, 31976, 33000, 35048, 36076, 38124, 40176, 41200, 43248, 44276, 46324,
	47352, 49400, 51452, 313596, 837884, 1363196, 1887484, 2412796, 2937084, 3461372, 3986684,
	4510972, 5036284, 5560572, 6084860, 6610172, 7134460, 7659772, 8184060, 8708348, 9233660, 9757948,
	10283260, 10807548, 11331836, 11857148, 12381436, 12906748, 13431036, 13955324, 14480636,
	15004924, 15530236, 16054524, 16579836, 16317692, 16055548, 15793404, 15269116, 15006972,
	14744828, 14220540, 13958396, 13696252, 13171964, 12909820, 12647676, 12123388, 11861244,
	11599100, 11074812, 10812668, 10550524, 10288380, 9764092, 9501948, 9239804, 8715516, 8453372,
	8191228, 7666940, 7404796, 7142652, 6618364, 6356220, 6094076, 5569788, 5307644, 5045500, 4783356,
	4259068, 3996924, 3734780, 3210492, 2948348, 2686204, 2161916, 1899772, 1637628, 1113340, 851196,
	589052, 64764, 63740, 62716, 61692, 59644, 58620, 57596, 55548, 54524, 53500, 51452, 50428,
	49404, 47356, 46332, 45308, 43260, 42236, 41212, 40188, 38140, 37116, 36092, 34044, 33020,
	31996, 29948, 28924, 27900, 25852, 24828, 23804, 21756, 20732, 19708, 18684, 16636, 15612,
	14588, 12540, 11516, 10492, 8444, 7420, 6396, 4348, 3324, 2300, 252, 248, 244, 240, 236, 232,
	228, 224, 220, 216, 212, 208, 204, 200, 196, 192, 188, 184, 180, 176, 172, 168, 164, 160, 156,
	152, 148, 144, 140, 136, 132, 128, 124, 120, 116, 112, 108, 104, 100, 96, 92, 88, 84, 80, 76,
	72, 68, 64, 60, 56, 52, 48, 44, 40, 36, 32, 28, 24, 20, 16, 12, 8, 0, 0}

type Mandel struct {
	w, h   int
	Iters  int
	size   int
	Center complex128 // f64 + f64 i
	Range complex128

	cr    complex128
	rir   float64
	scale float64
	image []uint32
	Lap   float64
}

func NewMandel(w, h, iters int, center complex128, range_ complex128) Mandel {
	return Mandel{w: w, h: h, Iters: iters, size: w * h, Center: center, Range: range_, cr: complex(real(range_), real(range_)), rir: imag(range_) - real(range_), scale: 0.8 * float64(w) / float64(h)}
}

func (m *Mandel) Update() {
	m.cr = complex(real(m.Range), real(m.Range))
	m.rir = imag(m.Range) - real(m.Range)
	m.scale = 0.8 * float64(m.w) / float64(m.h)
}

func (m *Mandel) genPixel(index int) {
	doScale := func(iw, jh int) complex128 {
		c00 := m.cr + complex(m.rir*float64(iw)/float64(m.w), m.rir*float64(jh)/float64(m.h))
		return complex(real(c00)*m.scale-real(m.Center), imag(c00)*m.scale-imag(m.Center))
	}

	c0 := doScale(index%m.w, index/m.w)

	z := c0
	i := 0

	dist := func(z complex128) float64 {
		return real(z)*real(z) + imag(z)*imag(z)
	}

	for i < m.Iters && dist(z) < 4.0 {
		z = z*z + c0
		i++
	}

	if i != m.Iters {
		m.image[index] = 0xff000000 | fire_pallete_256[(i<<2)%len(fire_pallete_256)]
	} else {
		m.image[index] = 0xff000000
	}
}

func (m *Mandel) GenImageSt() {
	m.image = make([]uint32, m.size)

	t0 := time.Now()

	for index := 0; index < m.size; index++ {
		m.genPixel(index)
	}
	m.Lap = float64(time.Since(t0).Milliseconds())
}

func (m *Mandel) GenImage() {
	m.image = make([]uint32, m.size)

	numCores := runtime.NumCPU()
	itemsPerCore := m.size / numCores

	t0 := time.Now()

	var wg sync.WaitGroup
	wg.Add(numCores)

	for th := range numCores {
		go func(th int) {
			defer wg.Done()
			for index := th * itemsPerCore; index < min((th+1)*itemsPerCore, m.size); index++ {
				m.genPixel(index)
			}
		}(th)
	}
	wg.Wait()

	m.Lap = float64(time.Since(t0).Milliseconds())
}

func (m *Mandel) GenerateImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, m.w, m.h))

	for i, argbPixel := range m.image {
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

func (m Mandel) WriteImage(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close() // Ensure the file is closed when the function exits

	byteOrder := binary.LittleEndian // Or binary.BigEndian

	err = binary.Write(file, byteOrder, m.image)
	if err != nil {
		fmt.Println("Error writing data:", err)
		return
	}
}

func (m Mandel) WritePng(filename string) error {
	img := m.GenerateImage()

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

func Mandelbrot(w, h, iters int, center, range_ complex128) image.Image {
	m := NewMandel(w, h, iters, center, range_)

	m.GenImage()
	return m.GenerateImage()
}

// x, y in w,h generate a new center, range_ and update mandel
func (m *Mandel) Recalculate(x, y float64) (complex128, complex128) {
	w, h := float64(m.w), float64(m.h)
	dist := float64(w / 2)
	rx := dist / w
	ry := dist / h
	ratio := math.Abs(real(m.Range))

	m.Center += complex(ratio*(w/2-x)/w, ratio*(h/2-y)/h)
	m.Range = complex(real(m.Range)*rx, imag(m.Range)*ry)

	m.Update()
	return m.Center, m.Range
}


