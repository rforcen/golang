package qh

/*
#cgo LDFLAGS: -lGL
#cgo CFLAGS: -I.

#include "render.h"
//#include "render.c"
*/
import "C"

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/chewxy/math32"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// UI holds the state for our user interface and camera controls.
type UI struct {
	window *glfw.Window

	// Mouse state
	firstMouse bool
	lastX      float32
	lastY      float32
	zoom       float32

	// waterman
	radius   float32
	faces    [][]int
	vertexes []*Point3d
	colors   []*Point3d

	update func()
	Lap    time.Duration

	geoUpdate bool
	listID    int
}

const (
	winW = 1024 * 2
	winH = 1024 * 2
)

func Do_UI() {
	// Must be called from main goroutine
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	// Create a window
	window, err := glfw.CreateWindow(winW, winH, "Polyhedron transformations", nil, nil)
	if err != nil {
		log.Fatalln("failed to create window:", err)
	}
	defer window.Destroy()

	window.MakeContextCurrent()

	// Initialize the UI struct
	ui := &UI{
		window:     window,
		firstMouse: true,
		lastX:      float32(winW / 2), // Center of the window
		lastY:      float32(winH / 2),
		zoom:       -4.0,
		radius:     math32.Floor(rand.Float32() * 10000),
		faces:      [][]int{},
		vertexes:   []*Point3d{},
		geoUpdate:  true,
		listID:     1,
		Lap:        0,
	}

	ui.update = func() {
		t0 := time.Now()

		ui.faces, ui.vertexes = WatermanPolyhedron(ui.radius)
		ui.calcColors()

		ui.geoUpdate = true
		ui.window.SetTitle(fmt.Sprintf("Waterman Polyhedron, Radius: %.0f, Faces: %d, Vertexes: %d, lap: %.0f ms", ui.radius, len(ui.faces), len(ui.vertexes), float32(time.Since(t0).Milliseconds())))
	}

	// Set input mode and callbacks
	// window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled) // Hide and lock the cursor
	window.SetCursorPosCallback(ui.mouseCallback)
	window.SetKeyCallback(ui.keyCallback)
	window.SetScrollCallback(func(window *glfw.Window, xoffset, yoffset float64) {
		ui.zoom += float32(yoffset)
	})

	ui.update()
	C.setPrefs()

	// Main loop
	for !window.ShouldClose() {
		w, h := window.GetSize()
		C.setGeo(C.int(ui.lastX), C.int(ui.lastY), C.float(ui.zoom), C.int(w), C.int(h))

		if ui.geoUpdate {

			ui.geoUpdate = false // compile render

			C.glDeleteLists(C.GLuint(ui.listID), 1)
			C.glNewList(C.GLuint(ui.listID), C.GL_COMPILE)

			for iface, face := range ui.faces { // faced poly
				C.glBegin(C.GL_POLYGON)
				color := ui.colors[iface]
				C.glColor3f(C.float(color.x), C.float(color.y), C.float(color.z))

				for _, iv := range face {
					c := ui.vertexes[iv]
					C.glVertex3f(C.float(c.x), C.float(c.y), C.float(c.z))
				}
				C.glEnd()
			}

			if len(ui.vertexes) < 5000 { // line poly
				C.glColor3f(0, 0, 0)
				for _, face := range ui.faces {
					C.glBegin(C.GL_LINE_LOOP)
					for _, iv := range face {
						c := ui.vertexes[iv]
						C.glVertex3f(C.float(c.x), C.float(c.y), C.float(c.z))
					}
					C.glEnd()
				}
			}

			C.glEndList()
		} else {
			C.glCallList(C.GLuint(ui.listID)) // draw the list
		}

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

// mouseCallback updates the yaw and pitch based on mouse movement.
func (ui *UI) mouseCallback(w *glfw.Window, xpos, ypos float64) {
	if ui.firstMouse {
		ui.lastX = float32(xpos)
		ui.lastY = float32(ypos)
		ui.firstMouse = false
	}

	xoffset := float32(xpos) - ui.lastX
	yoffset := float32(ypos) - ui.lastY // Reversed since Y-coordinates go from bottom to top
	ui.lastX = float32(xpos)
	ui.lastY = float32(ypos)

	sensitivity := 0.1
	xoffset *= float32(sensitivity)
	yoffset *= float32(sensitivity)

}

// keyCallback updates boolean flags based on keyboard input.
func (ui *UI) keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	t0 := time.Now()

	if action == glfw.Press {
		switch key {
		case glfw.KeyUp:
			ui.radius += 1
		case glfw.KeyDown:
			ui.radius -= 1
			if ui.radius < 4 {
				ui.radius = 4
			}
		case glfw.KeyPageUp:
			ui.radius -= 10
			if ui.radius < 10 {
				ui.radius = 10
			}
		case glfw.KeyPageDown:
			ui.radius += 10
		case glfw.KeySpace:
			ui.radius = math32.Floor(rand.Float32() * 10000)
		case glfw.KeyEnter:
			ui.radius = 10000
		case glfw.Key1:
			ui.radius = 10
		case glfw.Key2:
			ui.radius = 100
		case glfw.Key3:
			ui.radius = 1000
		case glfw.Key4:
			ui.radius = 10000
		case glfw.Key5:
			ui.radius = 20000
		case glfw.Key6:
			ui.radius = 25000
		case glfw.Key7:
			ui.radius = 30000
		case glfw.Key8:
			ui.radius = 35000
		case glfw.Key9:
			ui.radius = 40000
		case glfw.Key0:
			ui.radius = 45000

		// quit
		case glfw.KeyEscape: // exit
			w.SetShouldClose(true)
			return
		}

		ui.Lap = time.Since(t0)
		ui.update()
	}
}

func (ui *UI) calcColors() {

	sigDigits := func(f float32, n int) int { // n significant digits, loop solution is much faster than math one
		if f == 0 || n < 1 {
			return 0
		}

		powerOf10, powerOf100 := float32(1), float32(10)
		switch n {
		case 1:
		case 2:
			powerOf10, powerOf100 = float32(10), float32(100)
		case 3:
			powerOf10, powerOf100 = float32(100), float32(1000)
		case 4:
			powerOf10, powerOf100 = float32(1000), float32(10000)
		case 5:
			powerOf10, powerOf100 = float32(10000), float32(100000)
		default:
			powerOf10 = math32.Pow10(n - 1)
			powerOf100 = powerOf10 * 10
		}

		if f >= powerOf100 {
			for f >= powerOf100 {
				f /= 10
			}
		} else if f < powerOf10 {
			for f < powerOf10 {
				f *= 10
			}
		}

		return int(f)
	}

	ui.colors = make([]*Point3d, len(ui.faces))
	color_dict := map[int]*Point3d{} // color dictionary

	for iface, face := range ui.faces {
		// normal
		normal := normal(ui.vertexes[face[0]], ui.vertexes[face[1]], ui.vertexes[face[2]])

		// area
		vsum, vt := zeroPoint3d(), zeroPoint3d()
		v1, v2 := ui.vertexes[face[len(face)-2]], ui.vertexes[face[len(face)-1]]

		for _, v := range face {
			vt.cross(v1, v2)
			vsum.accumulate(vt)
			v1, v2 = v2, ui.vertexes[v]
		}
		area := math32.Abs(normal.dot(vsum)) / 2

		// color map -> color
		sf := sigDigits(area, 3)
		if _, ok := color_dict[sf]; !ok { // new random color to sf
			color_dict[sf] = newPoint3d(rand.Float32(), rand.Float32(), rand.Float32())
		}

		ui.colors[iface] = color_dict[sf]
	}
}
