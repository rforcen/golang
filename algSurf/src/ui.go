package algsurf

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

	coords   []Point3d
	normals  []Point3d
	textures []Point2d

	paramFuncNo int
	scale       float32
	res         int

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
	window, err := glfw.CreateWindow(winW, winH, "Algebraic Surfaces", nil, nil)
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
		zoom:       -2,
		geoUpdate:  true,
		listID:     1,

		paramFuncNo: 4,
		res:         512 / 2,
	}
	ui.update = func() {
		t0 := time.Now()
		ui.scale, ui.coords, ui.normals, ui.textures = ParamFuncCoords(ui.res, ParamDefs[ui.paramFuncNo].fromU, ParamDefs[ui.paramFuncNo].toU, ParamDefs[ui.paramFuncNo].fromV, ParamDefs[ui.paramFuncNo].toV, ParamDefs[ui.paramFuncNo].paramFunc)
		ui.Lap = time.Since(t0)

		ui.window.SetTitle(fmt.Sprintf("Algebraic Surfaces, %s, res:%d, %d vertices, %d triangles, lap: %.1f ms", ParamDefs[ui.paramFuncNo].name, ui.res, len(ui.coords), len(ui.textures), float64(ui.Lap)/1e6))
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
			
			// render
			C.glBegin(C.GL_QUADS)
			for i := 0; i < len(ui.coords); i += 4 { // draw trigs

				C.glNormal3f(C.float(ui.normals[i/4].x), C.float(ui.normals[i/4].y), C.float(ui.normals[i/4].z))
				for j := range 4 {
					C.glVertex3f(C.float(ui.coords[i+j].x*ui.scale), C.float(ui.coords[i+j].y*ui.scale), C.float(ui.coords[i+j].z*ui.scale))
				}

			}
			C.glEnd()
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
			ui.res += 32
		case glfw.KeyDown:
			ui.res -= 32
			if ui.res < 32 {
				ui.res = 32
			}
		case glfw.KeyLeft:
			ui.paramFuncNo++
			if ui.paramFuncNo >= len(ParamDefs) {
				ui.paramFuncNo = 0
			}
		case glfw.KeyRight:
			ui.paramFuncNo--
			if ui.paramFuncNo < 0 {
				ui.paramFuncNo = len(ParamDefs) - 1
			}

		case glfw.KeySpace:
			ui.paramFuncNo = 4
			ui.res = 512 / 2

		// quit
		case glfw.KeyEscape: // exit
			w.SetShouldClose(true)
			return
		}

		ui.update()
	}
}
