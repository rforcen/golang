package polygonizer

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

	update func()
	Lap    time.Duration

	vertexes  []VERTEX
	triangles []TRIANGLE
	msg       string

	implFuncNo int
	bounds     int
	size       float32

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
		geoUpdate:  true,
		listID:     1,

		implFuncNo: 1,
		bounds:     60,
		size:       0.06,
	}
	ui.update = func() {
		t0 := time.Now()
		ui.vertexes, ui.triangles, ui.msg = Polygonize(ImplicitFunctions[ui.implFuncNo].Function, ui.size, ui.bounds)
		ui.Lap = time.Since(t0)

		ui.window.SetTitle(fmt.Sprintf("Polygonizer, %s size:%f, bounds:%d, %d vertices, %d triangles, %s lap: %s", ImplicitFunctions[ui.implFuncNo].Name, ui.size, ui.bounds, len(ui.vertexes), len(ui.triangles), ui.msg, ui.Lap))
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
	// C.setPrefs()
	C.sceneInit()

	// Main loop
	for !window.ShouldClose() {
		w, h := window.GetSize()
		C.setGeo(C.int(ui.lastX), C.int(ui.lastY), C.float(ui.zoom), C.int(w), C.int(h))

		if ui.geoUpdate {
			ui.geoUpdate = false // compile render

			C.glDeleteLists(C.GLuint(ui.listID), 1) // delete & create list
			C.glNewList(C.GLuint(ui.listID), C.GL_COMPILE)

			C.glColor3f(C.float(0.5), C.float(0.5), C.float(0)) // golden object

			for _, triangle := range ui.triangles { // draw trigs

				C.glBegin(C.GL_TRIANGLES)
				for _, iv := range []int{triangle.i1, triangle.i2, triangle.i3} {
					c := ui.vertexes[iv]
					C.glNormal3f(C.float(c.normal.x), C.float(c.normal.y), C.float(c.normal.z))
					C.glVertex3f(C.float(c.position.x), C.float(c.position.y), C.float(c.position.z))
				}
				C.glEnd()

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
	if action == glfw.Press {
		switch key {
		case glfw.KeyUp:
			ui.bounds++
		case glfw.KeyDown:
			ui.bounds--
			if ui.bounds < 1 {
				ui.bounds = 1
			}
		case glfw.KeyLeft:
			ui.size *= 0.9
		case glfw.KeyRight:
			ui.size *= 1.1

		case glfw.KeyPageUp:
			ui.implFuncNo++
			if ui.implFuncNo >= len(ImplicitFunctions) {
				ui.implFuncNo = 0
			}
			ui.size = ImplicitFunctions[ui.implFuncNo].Size
			ui.bounds = ImplicitFunctions[ui.implFuncNo].Bounds
		case glfw.KeyPageDown:
			ui.implFuncNo--
			if ui.implFuncNo < 0 {
				ui.implFuncNo = len(ImplicitFunctions) - 1
			}
			ui.size = ImplicitFunctions[ui.implFuncNo].Size
			ui.bounds = ImplicitFunctions[ui.implFuncNo].Bounds

		case glfw.KeySpace:
			ui.implFuncNo = 0
			ui.size = ImplicitFunctions[ui.implFuncNo].Size
			ui.bounds = ImplicitFunctions[ui.implFuncNo].Bounds

		// quit
		case glfw.KeyEscape: // exit
			w.SetShouldClose(true)
			return
		}

		ui.update()
	}
}
