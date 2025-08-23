package sh

/*
#cgo LDFLAGS: -lGL
#cgo CFLAGS: -I.

#include "render.h"
*/
import "C"

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"
	"unsafe"

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

	// sh
	res           int
	colorMap      int
	code          int
	sh_           *SH
	update        func()
	multiThreaded bool
	listID        int
	geoUpdate     bool
}

const (
	winW = 1024 * 2
	winH = 1024 * 2
)


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
		case glfw.KeySpace, glfw.KeyR:
			ui.code = RandCode()

		case glfw.KeyUp: // res +-
			ui.res *= 2

		case glfw.KeyDown:
			ui.res /= 2

		case glfw.KeyLeft: // colormap +-
			ui.colorMap = (ui.colorMap + 1) % N_COLOR_MAPS

		case glfw.KeyRight:
			ui.colorMap = (ui.colorMap + N_COLOR_MAPS - 1) % N_COLOR_MAPS

		case glfw.KeyM: // multi thread
			ui.multiThreaded = !ui.multiThreaded

		case glfw.KeyT: // test worky code
			ui.code = FindCode(88888888)

		case glfw.KeyW: // write obj
			ui.sh_.WriteObj(fmt.Sprintf("sh_%d_%d_%d.obj", ui.res, ui.colorMap, SH_codes[ui.code]))
			return

		case glfw.KeyEscape: // exit
			w.SetShouldClose(true)
			return
		}

		ui.sh_ = NewSH(ui.res, ui.colorMap, ui.code)
		ui.update()
	}
}

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
	window, err := glfw.CreateWindow(winW, winH, "Spherical Harmonics", nil, nil)
	if err != nil {
		log.Fatalln("failed to create window:", err)
	}
	defer window.Destroy()

	window.MakeContextCurrent()

	// Initialize the UI struct
	ui := &UI{
		window:        window,
		firstMouse:    true,
		lastX:         float32(winW / 2), // Center of the window
		lastY:         float32(winH / 2),
		zoom:          -4.0,
		res:           512,
		colorMap:      rand.Intn(25),
		code:          RandCode(),
		sh_:           NewSH(512, rand.Intn(25), RandCode()),
		multiThreaded: true,
		listID:        -1,
		geoUpdate:     true,
	}
	ui.update = func() {
		t0 := time.Now()
		var mt string
		if ui.multiThreaded {
			ui.sh_.CalcMeshMt()
			mt = "MT"
		} else {
			ui.sh_.CalcMesh()
			mt = "ST"
		}
		ui.geoUpdate = true
		ui.window.SetTitle(fmt.Sprintf("Spherical Harmonics [%s], code: %d, res: %d, color_map: %d, lap: %.2f ms", mt, ui.code, ui.res, ui.colorMap, time.Since(t0).Seconds()*1e3))
	}

	// Set input mode and callbacks
	// window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled) // Hide and lock the cursor
	window.SetCursorPosCallback(ui.mouseCallback)
	window.SetKeyCallback(ui.keyCallback)
	window.SetScrollCallback(func(window *glfw.Window, xoffset, yoffset float64) {
		ui.zoom += float32(yoffset)
	})

	ui.update()

	C.sceneInit()	

	// Main loop
	for !window.ShouldClose() {
		w, h := window.GetSize()
		C.setGeo(C.int(ui.lastX), C.int(ui.lastY), C.float(ui.zoom), C.int(w), C.int(h))
		if ui.geoUpdate {
			ui.geoUpdate = false

			C.glDeleteLists(C.GLuint(ui.listID), 1)
			C.glNewList(C.GLuint(ui.listID), C.GL_COMPILE)

			C.drawMesh((*C.CLocation)(unsafe.Pointer(&ui.sh_.Mesh[0])), C.int(ui.sh_.Res))
			C.glEndList()
		} else {
			C.glCallList(C.GLuint(ui.listID)) // draw the list
		}

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
