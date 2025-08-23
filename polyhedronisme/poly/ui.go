package poly

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

	// poly
	Poly  *Polyhedron
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
		Poly:      NewPolyhedron(Dodecahedron),
		geoUpdate:  true,
		listID:     1,
	}
	ui.update = func() {
		ui.Poly.Recalc()
		ui.window.SetTitle(fmt.Sprintf("Polyhedron transformations [%s], Faces: %d, Vertexes: %d, Colors: %d, lap: %.0f ms", ui.Poly.Name, len(ui.Poly.Faces), len(ui.Poly.Vertexes), len(*Unique(&ui.Poly.Colors)), float64(ui.Lap.Milliseconds())))
		ui.geoUpdate = true
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

			for iface, face := range ui.Poly.Faces { // faced poly
				C.glBegin(C.GL_POLYGON)
				for _, iv := range face {
					c := ui.Poly.Vertexes[iv]
					col := ui.Poly.Colors[iface]
					C.glColor3f(C.float(col.X), C.float(col.Y), C.float(col.Z))
					C.glVertex3f(C.float(c.X), C.float(c.Y), C.float(c.Z))
				}
				C.glEnd()
			}

			if len(ui.Poly.Vertexes) < 5000 { // line poly
				C.glColor3f(0, 0, 0)
				for _, face := range ui.Poly.Faces {
					C.glBegin(C.GL_LINE_LOOP)
					for _, iv := range face {
						c := ui.Poly.Vertexes[iv]
						C.glVertex3f(C.float(c.X), C.float(c.Y), C.float(c.Z))
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
		// plato solids
		case glfw.KeyC:
			ui.Poly = NewPolyhedron(Cube)
		case glfw.KeyT:
			ui.Poly = NewPolyhedron(Tetrahedron)
		case glfw.KeyO:
			ui.Poly = NewPolyhedron(Octahedron)
		case glfw.KeyD:
			ui.Poly = NewPolyhedron(Dodecahedron)
		case glfw.KeyI:
			ui.Poly = NewPolyhedron(Icosahedron)
		case glfw.KeyJ:
			ui.Poly = NewPolyhedron(Johnson[rand.Intn(len(Johnson))])

		// transformations
		case glfw.KeyK: // kiss
			ui.Poly = ui.Poly.Kiss_n(0, 0.05)

		case glfw.KeyA: // ambo
			ui.Poly = ui.Poly.Ambo()

		case glfw.KeyQ: // quinto
			ui.Poly = ui.Poly.Quinto()

		case glfw.KeyH: // hollow
			ui.Poly = ui.Poly.Hollow(0.2, 0.1)

		case glfw.KeyG: // gyro
			ui.Poly = ui.Poly.Gyro()

		case glfw.KeyP: // propellor
			ui.Poly = ui.Poly.Propellor()

		case glfw.KeyU: // dual
			ui.Poly = ui.Poly.Dual()

		case glfw.KeyF: // chamfer
			ui.Poly = ui.Poly.Chamfer(0.1)

		case glfw.KeyN: // inset
			ui.Poly = ui.Poly.Inset(0, 0.3, -0.1)

		// write obj
		case glfw.KeyW:
			ui.Poly.WriteObj()
		// space = random transform to a cube
		case glfw.KeySpace:
			ui.Poly = NewPolyhedron(Cube)

			trs := "kaqhgpdci"
			for range rand.Intn(10) + 1 {
				if len(ui.Poly.Vertexes) > 100000 { // max vertexes
					break
				}
				switch trs[rand.Intn(len(trs))] {
				case 'k':
					ui.Poly = ui.Poly.Kiss_n(0, 0.05)
				case 'a':
					ui.Poly = ui.Poly.Ambo()
				case 'q':
					ui.Poly = ui.Poly.Quinto()
				case 'h':
					ui.Poly = ui.Poly.Hollow(0.2, 0.1)
				case 'g':
					ui.Poly = ui.Poly.Gyro()
				case 'p':
					ui.Poly = ui.Poly.Propellor()
				case 'd':
					ui.Poly = ui.Poly.Dual()
				case 'c':
					ui.Poly = ui.Poly.Chamfer(0.1)
				case 'i':
					ui.Poly = ui.Poly.Inset(0, 0.3, -0.1)
				}
			}
		// quit
		case glfw.KeyEscape: // exit
			w.SetShouldClose(true)
			return
		}

		ui.Lap = time.Since(t0)
		ui.update()
	}
}
