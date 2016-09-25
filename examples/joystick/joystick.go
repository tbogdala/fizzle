// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"os"
	"runtime"

	gl "github.com/go-gl/gl/v3.3-core/gl"
	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/fizzle/graphicsprovider/opengl"
	forward "github.com/tbogdala/fizzle/renderer/forward"

	gui "github.com/tbogdala/eweygewey"
	guiinput "github.com/tbogdala/eweygewey/glfwinput"
)

/*
  This example illustrates how to display a user interface using text widgets.

  Coincidently, this also can be used as a tool to see what axis/buttons do
  for a given joystick in GLFW.
*/

// GLFW event handling must run on the main OS thread.
func init() {
	runtime.LockOSThread()
}

const (
	width  = 1280
	height = 720

	fontScale    = 14
	fontGlyphs   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890. :[]{}\\|<>;\"'~`?/-+_=()*&^%$#@!"
	fontFilepath = "../assets/Oswald-Heavy.ttf"
)

// main is the entry point for the application.
func main() {
	// start off by initializing the GL and GLFW libraries and creating a window.
	// the default window size we use is 800x600
	mainWindow, gfx := initGraphics("Joystick Mapper", width, height)

	// set the callback function for key input
	mainWindow.SetKeyCallback(keyCallback)

	// create a new renderer
	renderer := forward.NewForwardRenderer(gfx)
	defer renderer.Destroy()

	// create and initialize the gui Manager
	uiman := gui.NewManager(gfx)
	err := uiman.Initialize(gui.VertShader330, gui.FragShader330, width, height, height)
	if err != nil {
		fmt.Printf("Failed to initialize the user interface!\n%v", err)
		os.Exit(1)
	}
	guiinput.SetInputHandlers(uiman, mainWindow)

	// load a font
	_, err = uiman.NewFont("Default", fontFilepath, fontScale, fontGlyphs)
	if err != nil {
		fmt.Printf("Failed to load the font file!\n%v", err)
		os.Exit(1)
	}

	// load the text shader
	textShader, err := forward.CreateColorTextShader()
	if err != nil {
		fmt.Printf("Failed to compile and link the text shader program!\n%v", err)
		os.Exit(1)
	}
	defer textShader.Destroy()

	// setup the camera to look at the cube
	camera := fizzle.NewYawPitchCamera(mgl.Vec3{0.0, 1.0, 5.0})
	camera.LookAtDirect(mgl.Vec3{0, 0, 0})

	// set some OpenGL flags
	gfx.BlendEquation(graphics.FUNC_ADD)
	gfx.BlendFunc(graphics.SRC_ALPHA, graphics.ONE_MINUS_SRC_ALPHA)
	gfx.Enable(graphics.BLEND)
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)

	// create a window to display the joystick properties
	joystickWindow := uiman.NewWindow("JoystickWnd", 0.1, 0.9, 0.8, 0.8, func(wnd *gui.Window) {
		// write out the name of the joystick
		if glfw.JoystickPresent(glfw.Joystick1) {
			wnd.Text(fmt.Sprintf("Joystick: %s", glfw.GetJoystickName(glfw.Joystick1)))
		} else {
			wnd.Text("Joystick1 Not Detected")
			return
		}

		// get the axes and button values
		axes := glfw.GetJoystickAxes(glfw.Joystick1)
		buttons := glfw.GetJoystickButtons(glfw.Joystick1)

		// create a section for the button mappings
		wnd.Separator()
		for i, b := range buttons {
			if i != 0 {
				wnd.StartRow()
			}
			wnd.Text(fmt.Sprintf("Button %d: %d", i, b))
		}

		// create a section for the axis mappings
		wnd.Separator()
		for i, f := range axes {
			if i != 0 {
				wnd.StartRow()
			}
			wnd.Text(fmt.Sprintf("Axis %d: %f", i, f))
		}
	})
	joystickWindow.Title = "Button Mappings"
	//joystickWindow.Style.WindowBgColor[3] = 1.0 // turn off transparent bg

	// loop until something told the mainWindow that it should close
	for !mainWindow.ShouldClose() {
		// clear the screen
		gl.Viewport(0, 0, int32(width), int32(height))
		gl.ClearColor(0.05, 0.05, 0.05, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// draw the user interface
		uiman.Construct(0)
		uiman.Draw()

		// draw the screen
		mainWindow.SwapBuffers()
		glfw.PollEvents()
	}
}

// initGraphics creates an OpenGL window and initializes the required graphics libraries.
// It will either succeed or panic.
func initGraphics(title string, w int, h int) (*glfw.Window, graphics.GraphicsProvider) {
	// GLFW must be initialized before it's called
	err := glfw.Init()
	if err != nil {
		panic("Can't init glfw! " + err.Error())
	}

	// request a OpenGL 3.3 core context
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// do the actual window creation
	mainWindow, err := glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		panic("Failed to create the main window! " + err.Error())
	}
	mainWindow.MakeContextCurrent()

	// initialize OpenGL
	gfx, err := opengl.InitOpenGL()
	if err != nil {
		panic("Failed to initialize OpenGL! " + err.Error())
	}
	fizzle.SetGraphics(gfx)

	return mainWindow, gfx
}

// keyCallback is set as a callback in main() and is used to close the window
// when the escape key is hit.
func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}
}
