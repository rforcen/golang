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
	"math"
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

	// waterman
	radius float64
	faces [][]int
	vertexes []*Point3d

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
		radius:     10,
		faces:      [][]int{},
		vertexes:   []*Point3d{},
		geoUpdate:  true,
		listID:     1,
		Lap:        0,
	}

	ui.update = func() {
		t0 := time.Now()

		ui.faces, ui.vertexes = WatermanPolyhedron(ui.radius)
		
		ui.geoUpdate = true	
		ui.window.SetTitle(fmt.Sprintf("Waterman Polyhedron, Radius: %.0f, Faces: %d, Vertexes: %d, lap: %.0f ms", ui.radius, len(ui.faces), len(ui.vertexes), float64(time.Since(t0).Milliseconds())))
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

			for _, face := range ui.faces { // faced poly
				C.glBegin(C.GL_POLYGON)
				for _, iv := range face {
					c := ui.vertexes[iv]
					
					C.glColor3f(C.float(rand.Float32()), C.float(rand.Float32()), C.float(rand.Float32()))
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
		case glfw.KeyPageDown:
			ui.radius -= 10
			if ui.radius < 10 {
				ui.radius = 10
			}
		case glfw.KeyPageUp:
			ui.radius += 10
		case glfw.KeySpace:
			ui.radius = math.Floor(rand.Float64() * 10000)
		case glfw.KeyEnter:
			ui.radius = 10000
		// quit
		case glfw.KeyEscape: // exit
			w.SetShouldClose(true)
			return
		}

		ui.Lap = time.Since(t0)
		ui.update()
	}
}
