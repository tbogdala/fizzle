// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	gl32 "github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/fizzle/editor"
	"github.com/tbogdala/fizzle/graphicsprovider/opengl"
	"github.com/tbogdala/fizzle/renderer"
	"github.com/tbogdala/fizzle/renderer/forward"
)

var (
	windowTitle  = "Fizzle Game Editor"
	windowWidth  = 1280
	windowHeight = 720
	exitChan     chan bool
	bgColor      [4]float32
	render       renderer.Renderer
)

// command line flags
var (
	versionString = "v0.1.0 DEVELOPMENT"
	appFlags      = kingpin.New("gameeditor", "game editor for the Fizzle graphics library.")
	flagComponent = appFlags.Flag("component", "edit the component JSON file specified").Default("").String()
)

// lock our main goroutine down for gl/glfw
func init() {
	runtime.LockOSThread()
}

func main() {
	appFlags.Version(versionString)
	appFlags.Parse(os.Args[1:])

	// create the main window
	win, err := initGlfw()
	if err != nil {
		fmt.Println(err)
		return
	}

	// create a renderer
	gfx := fizzle.GetGraphics()
	windowWidth, windowHeight := win.GetSize()
	render = forward.NewForwardRenderer(gfx)
	render.ChangeResolution(int32(windowWidth), int32(windowHeight))
	defer render.Destroy()

	// construct a new editor state
	levelEd, err := editor.NewState(win, render)
	if err != nil {
		fmt.Println(err)
		return
	}

	// if a component file was specified, switch editing modes
	// and load up the component
	if *flagComponent != "" {
		levelEd.SetMode(editor.ModeComponent)
		err = levelEd.LoadComponentFile(*flagComponent)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

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

	// initialize OpenGL
	gfx, err := opengl.InitOpenGL()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenGL 3.3: %v", err)
	}
	fizzle.SetGraphics(gfx)

	// make sure we initialize 3.2 for nuklear
	if err := gl32.Init(); err != nil {
		return nil, fmt.Errorf("Failed to initialize OpenGL 3.2: %v", err)
	}

	return win, nil
}

func onWindowResize(w *glfw.Window, width int, height int) {
	render.ChangeResolution(int32(width), int32(height))
}
