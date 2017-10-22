// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"runtime"
	"time"

	gl32 "github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/tbogdala/fizzle/editor"
)

var (
	windowTitle  = "Fizzle Game Editor"
	windowWidth  = 1280
	windowHeight = 720
	exitChan     chan bool
	bgColor      [4]float32
)

// lock our main goroutine down for gl/glfw
func init() {
	runtime.LockOSThread()
}

func main() {
	// create the main window
	win, err := initGlfw()
	if err != nil {
		fmt.Println(err)
		return
	}

	// construct a new editor state
	levelEd, err := editor.NewState(win)
	if err != nil {
		fmt.Println(err)
		return
	}
	levelEd.SetMode(editor.ModeComponent)

	// render loop time
	exitChan := make(chan bool, 1)
	fpsTicker := time.NewTicker(time.Second / 30)
	for {
		select {
		case <-exitChan:
			fpsTicker.Stop()
			nk.NkPlatformShutdown()
			glfw.Terminate()
			return
		case <-fpsTicker.C:
			if win.ShouldClose() {
				exitChan <- true
				continue
			}
			glfw.PollEvents()

			// clear the screen
			width, height := win.GetSize()
			gl.Viewport(0, 0, int32(width), int32(height))
			gl.Clear(gl.COLOR_BUFFER_BIT)
			gl.ClearColor(bgColor[0], bgColor[1], bgColor[2], bgColor[3])

			// draw the editor interface
			levelEd.Render()

			win.SwapBuffers()
		}
	}
}

// initGlfw initializes the GLFW window and OpenGL context
func initGlfw() (*glfw.Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	win, err := glfw.CreateWindow(windowWidth, windowHeight, windowTitle, nil, nil)
	if err != nil {
		return nil, err
	}
	win.SetSizeCallback(onWindowResize)
	win.MakeContextCurrent()
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("Failed to initialize OpenGL 3.3: %v", err)
	}

	// make sure we initialize 3.2 for nuklear
	if err := gl32.Init(); err != nil {
		return nil, fmt.Errorf("Failed to initialize OpenGL 3.2: %v", err)
	}

	return win, nil
}

func onWindowResize(w *glfw.Window, width int, height int) {
	//	renderer.ChangeResolution(int32(width), int32(height))
}
