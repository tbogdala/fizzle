// Copyright 2015, Timothy` Bogdala <tdb@animal-machine.com>
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
	ui "github.com/tbogdala/fizzle/ui"
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
	width          = 800
	height         = 600
	textShaderPath = "./assets/forwardshaders/colortext"

	fontScale    = 24
	fontGlyphs   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890. :[]{}\\|<>;\"'~`?/-+_=()*&^%$#@!"
	fontFilepath = "assets/HammersmithOne.ttf"
)

// main is the entry point for the application.
func main() {
	// start off by initializing the GL and GLFW libraries and creating a window.
	// the default window size we use is 800x600
	mainWindow, gfx := initGraphics("Joystick Mapper", width, height)

	// set the callback function for key input
	mainWindow.SetKeyCallback(keyCallback)

	// create a new renderer
	renderer := forward.NewForwardRenderer(mainWindow, gfx)
	defer renderer.Destroy()

	// setup the user interface manager
	uiManager := ui.NewUIManager()
	uiManager.AdviseResolution(width, height)
	defer uiManager.Destroy()
	renderer.UIManager = uiManager

	// load the text shader
	textShader, err := fizzle.LoadShaderProgramFromFiles(textShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the text shader program!\n%v", err)
		os.Exit(1)
	}
	defer textShader.Destroy()

	// load the font
	font, err := fizzle.NewGLFont(fontFilepath, fontScale, fontGlyphs)
	if err != nil {
		fmt.Print("Failed to process font file!\n" + err.Error())
	}
	font.Shader = textShader
	defer font.Destroy()

	// setup the camera to look at the cube
	camera := fizzle.NewYawPitchCamera(mgl.Vec3{0.0, 1.0, 5.0})
	camera.LookAtDirect(mgl.Vec3{0, 0, 0})

	// set some OpenGL flags
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)

	createStateWidgets := func() {
		// clear the ui manager
		uiManager.Destroy()

		// get the name of the joystick from GLFW ...
		var nameWidget *ui.UILabel
		var joystickName string
		if glfw.JoystickPresent(glfw.Joystick1) {
			joystickName = fmt.Sprintf("Joystick: %s", glfw.GetJoystickName(glfw.Joystick1))
			nameWidget = uiManager.CreateLabel(font, ui.UIAnchorTopMiddle, mgl.Vec3{0.0, 0.0, 10.0}, joystickName)
		} else {
			joystickName = "Joystick1 Not Detected"
			uiManager.CreateLabel(font, ui.UIAnchorTopMiddle, mgl.Vec3{0.0, 0.0, 10.0}, joystickName)
			uiManager.LayoutWidgets()
			return
		}

		// get the axes and button values
		axes := glfw.GetJoystickAxes(glfw.Joystick1)
		buttons := glfw.GetJoystickButtons(glfw.Joystick1)
		startHeight := float32(10.0 + nameWidget.Renderable.BoundingRect.DeltaY()*2.0)

		// setup a column and make labels for the button states
		var w *ui.UILabel
		x := float32(10.0)
		y := startHeight
		for i, b := range buttons {
			wString := fmt.Sprintf("B%02d: %d", i, b)
			w = uiManager.CreateLabel(font, ui.UIAnchorTopLeft, mgl.Vec3{x, -y, 0.0}, wString)
			y += w.Renderable.BoundingRect.DeltaY() * 1.5

			if w.Renderable.BoundingRect.DeltaY()+y > float32(height)*.90 {
				x += 10.0 + w.Renderable.BoundingRect.DeltaX()
				y = startHeight
			}
		}

		// setup a new column and make labels for the axis values
		if w != nil {
			x += 10.0 + w.Renderable.BoundingRect.DeltaX()
		}
		y = startHeight
		for i, f := range axes {
			wString := fmt.Sprintf("Axis %02d: %f", i, f)
			w = uiManager.CreateLabel(font, ui.UIAnchorTopLeft, mgl.Vec3{x, -y, 0.0}, wString)
			y += w.Renderable.BoundingRect.DeltaY() * 1.5

			if w.Renderable.BoundingRect.DeltaY()+y > float32(height)*.90 {
				x += 10.0 + w.Renderable.BoundingRect.DeltaX()
				y = startHeight
			}
		}

		// layout the widgets
		uiManager.LayoutWidgets()
	}

	// loop until something told the mainWindow that it should close
	for !mainWindow.ShouldClose() {
		// recreate all of the state widgets
		createStateWidgets()

		// clear the screen
		gl.Viewport(0, 0, int32(width), int32(height))
		gl.ClearColor(0.05, 0.05, 0.05, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Finish with the user interface
		uiManager.Draw(renderer, nil)

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
