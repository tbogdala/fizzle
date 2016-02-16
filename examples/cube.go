// Copyright 2015, Timothy` Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/fizzle/graphicsprovider/opengl"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

/*
  This example illustrates the bare minimum to set up an application
  using the fizzle library.

  It does the following:

    1) creates a GFLW window for rendering
    2) creates a renderer
    3) loads some shaders
    4) creates a cube and a sphere
    5) in a loop, render the cube or sphere
		6) when escape is pressed, exit the loop
		7) when spacebar is pressed toggle which shape to draw

  This example also does not use the 'example app' framework so that
  it can be as compact and illustrative of the minimal requirements
  as possible.
*/

// GLFW event handling must run on the main OS thread. If this doesn't get
// locked down, you will likely see random crashes on memory access while
// running the application after a few seconds.
//
// So on initialization of the module, lock the OS thread for this goroutine.
func init() {
	runtime.LockOSThread()
}

const (
	width             = 800
	height            = 600
	radsPerSec        = math.Pi / 4.0
	diffuseShaderPath = "./assets/forwardshaders/diffuse"
)

var (
	// renderCube indicates if the cube should be drawn or the sphere
	renderCube bool = true
)

// main is the entry point for the application.
func main() {
	// start off by initializing the GL and GLFW libraries and creating a window.
	// the default window size we use is 800x600
	mainWindow, gfx := initGraphics("Simple Cube", width, height)

	// set the callback function for key input
	mainWindow.SetKeyCallback(keyCallback)

	// create a new renderer
	renderer := forward.NewForwardRenderer(mainWindow, gfx)
	defer renderer.Destroy()

	// put a light in there
	light := renderer.NewLight()
	//light.Position = mgl.Vec3{-10.0, 5.0, 10}
	light.DiffuseColor = mgl.Vec4{1.0, 0.0, 0.0, 1.0}
	light.Direction = mgl.Vec3{1.0, -0.5, -1.0}
	light.DiffuseIntensity = 0.80
	light.AmbientIntensity = 0.20
	light.Attenuation = 1.0
	renderer.ActiveLights[0] = light

	// load the diffuse shader
	diffuseShader, err := fizzle.LoadShaderProgramFromFiles(diffuseShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the diffuse shader program!\n%v", err)
		os.Exit(1)
	}
	defer diffuseShader.Destroy()

	// create a 2x2x2 cube to render
	cube := fizzle.CreateCube("diffuse", -1, -1, -1, 1, 1, 1)
	cube.Core.Shader = diffuseShader
	cube.Core.DiffuseColor = mgl.Vec4{0.9, 0.9, 0.9, 1.0}
	cube.Core.SpecularColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	cube.Core.Shininess = 4.8

	// create a sphere to render
	sphere := fizzle.CreateSphere("diffuse", 1, 16, 16)
	sphere.Core.Shader = diffuseShader
	sphere.Core.DiffuseColor = mgl.Vec4{0.9, 0.9, 0.9, 1.0}
	sphere.Core.SpecularColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	sphere.Core.Shininess = 4.8

	renderCube = true

	// setup the camera to look at the cube
	camera := fizzle.NewYawPitchCamera(mgl.Vec3{0.0, 1.0, 5.0})
	camera.LookAtDirect(mgl.Vec3{0, 0, 0})

	// set some OpenGL flags
	gfx.Enable(graphics.CULL_FACE)
	gfx.Enable(graphics.DEPTH_TEST)

	// loop until something told the mainWindow that it should close
	lastFrame := time.Now()
	for !mainWindow.ShouldClose() {
		// calculate the difference in time to control rotation speed
		thisFrame := time.Now()
		frameDelta := float32(thisFrame.Sub(lastFrame).Seconds())

		// rotate the cube and sphere around the Y axis at a speed of radsPerSec
		rotDelta := mgl.QuatRotate(radsPerSec*frameDelta, mgl.Vec3{0.0, 1.0, 0.0})
		cube.LocalRotation = cube.LocalRotation.Mul(rotDelta)
		sphere.LocalRotation = sphere.LocalRotation.Mul(rotDelta)

		// clear the screen
		gfx.Viewport(0, 0, int32(width), int32(height))
		gfx.ClearColor(0.05, 0.05, 0.05, 1.0)
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		// make the projection and view matrixes
		perspective := mgl.Perspective(mgl.DegToRad(60.0), float32(width)/float32(height), 1.0, 100.0)
		view := camera.GetViewMatrix()

		// draw the cube or the sphere
		if renderCube {
			renderer.DrawRenderable(cube, nil, perspective, view)
		} else {
			renderer.DrawRenderable(sphere, nil, perspective, view)
		}

		// draw the screen
		mainWindow.SwapBuffers()

		// advise GLFW to poll for input. without this the window appears to hang.
		glfw.PollEvents()

		// update our last frame time
		lastFrame = thisFrame
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
	glfw.WindowHint(glfw.Samples, 0)
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

	// disable v-sync for max draw rate
	glfw.SwapInterval(0)

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

	if key == glfw.KeySpace && action == glfw.Press {
		// spacebar toggles the drawing of the cube or the sphere
		if renderCube {
			renderCube = false
		} else {
			renderCube = true
		}
	}
}
