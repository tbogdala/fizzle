// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"log"
	"math"
	"runtime"

	"github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	input "github.com/tbogdala/fizzle/input/glfwinput"
	"github.com/tbogdala/fizzle/renderer/forward"
)

func init() {
	runtime.LockOSThread()
}

const (
	windowWidth  = 400
	windowHeight = 400
	shaderPath   = "diffuse"
)

var (
	mainWindow *glfw.Window
	renderer   *forward.ForwardRenderer
	gfx        graphics.GraphicsProvider

	shader *fizzle.RenderShader

	objects = make(map[string]*fizzle.Renderable)
	shapes  = make(map[string]*fizzle.Renderable)

	camera *fizzle.OrbitCamera

	kbModel *input.KeyboardModel
)

func newWindow() {
	mainWindow, gfx = initGraphics("Simple Cube", windowWidth, windowHeight)

	// set the callback functions for key input
	kbModel = input.NewKeyboardModel(mainWindow)
	kbModel.BindTrigger(glfw.KeyEscape, setShouldClose)
	kbModel.SetupCallbacks()

	// create a new renderer
	renderer = forward.NewForwardRenderer(gfx)
	renderer.ChangeResolution(windowWidth, windowHeight)
	// defer renderer.Destroy()

	// put a light in there
	light := renderer.NewLight()
	//light.Position = mgl.Vec3{-10.0, 5.0, 10}
	light.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	light.Direction = mgl.Vec3{1.0, -0.5, -1.0}
	light.DiffuseIntensity = 0.70
	light.SpecularIntensity = 0.10
	light.AmbientIntensity = 0.20
	light.LinearAttenuation = 1.0
	renderer.ActiveLights[0] = light

	var err error
	// load the shader
	shader, err = fizzle.LoadShaderProgramFromFiles(shaderPath, nil)
	if err != nil {
		log.Fatalln("Failed to compile and link the diffuse shader program!\n%v", err)
	}

	camera = fizzle.NewOrbitCamera(mgl.Vec3{0, 0, 0}, math.Pi/1.5, 20, math.Pi/2.0)

	gfx.Enable(graphics.CULL_FACE)
	gfx.Enable(graphics.DEPTH_TEST)

}

func renderLoop() {
	for !mainWindow.ShouldClose() {
		width, height := renderer.GetResolution()
		gfx.Viewport(0, 0, int32(width), int32(height))
		gfx.ClearColor(0.25, 0.25, 0.25, 1.0)
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)
		perspective := mgl.Perspective(mgl.DegToRad(60.0), float32(width)/float32(height), 1.0, 100.0)
		view := camera.GetViewMatrix()

		startTests()

		for _, o := range objects {
			renderer.DrawRenderable(o, nil, perspective, view, camera)
		}

		for _, o := range shapes {
			renderer.DrawLines(o, shader, nil, perspective, view, camera)
		}

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
	glfw.WindowHint(glfw.Samples, 0)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// do the actual window creation
	mainWindow, err = glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		panic("Failed to create the main window! " + err.Error())
	}
	mainWindow.SetSizeCallback(onWindowResize)
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

// setShouldClose should be called to close the window and kill the app.
func setShouldClose() {
	mainWindow.SetShouldClose(true)
}

// onWindowResize is called when the window changes size
func onWindowResize(w *glfw.Window, width int, height int) {
	renderer.ChangeResolution(int32(width), int32(height))
}
