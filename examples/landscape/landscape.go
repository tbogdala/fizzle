// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"math"
	"runtime"
	"time"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	input "github.com/tbogdala/fizzle/input/glfwinput"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

// GLFW event handling must run on the main OS thread. If this doesn't get
// locked down, you will likely see random crashes on memory access while
// running the application after a few seconds.
//
// So on initialization of the module, lock the OS thread for this goroutine.
func init() {
	runtime.LockOSThread()
}

const (
	windowWidth    = 1200
	windowHeight   = 768
	landscapeSize  = 256
	landscapeScale = 512.0
	movementScale  = float32(50.0)
	diffuseFile    = "../assets/textures/sample_terrain_d.png"
	heightmapFile  = "../assets/textures/sample_terrain_h.png"
)

var (
	mainWindow *glfw.Window
	renderer   *forward.ForwardRenderer
	frameDelta float32
	camera     *fizzle.YawPitchCamera
)

// main is the entry point for the application.
func main() {
	// start off by initializing the GL and GLFW libraries and creating a window.
	// the default window size we use is 800x600
	w, gfx := initGraphics("Landscape", windowWidth, windowHeight)
	mainWindow = w

	// set the callback functions for key input
	kbModel := input.NewKeyboardModel(mainWindow)
	kbModel.BindTrigger(glfw.KeyEscape, setShouldClose)
	kbModel.Bind(glfw.KeyW, handleMoveForward)
	kbModel.Bind(glfw.KeyS, handleMoveBackward)
	kbModel.Bind(glfw.KeyA, handleSlideLeft)
	kbModel.Bind(glfw.KeyD, handleSlideRight)
	kbModel.Bind(glfw.KeyQ, handleFloatUp)
	kbModel.Bind(glfw.KeyE, handleFloatDown)
	kbModel.SetupCallbacks()

	// set the callback for the mouse position update
	w.SetCursorPosCallback(handleMouseMovement)

	// create a new renderer
	renderer = forward.NewForwardRenderer(gfx)
	renderer.ChangeResolution(windowWidth, windowHeight)
	defer renderer.Destroy()

	// put a light in there
	light := renderer.NewDirectionalLight(mgl.Vec3{-1.0, 0.0, 0.0}.Normalize())
	light.AmbientIntensity = 0.4
	light.DiffuseIntensity = 0.5
	light.SpecularIntensity = 0.1
	renderer.ActiveLights[0] = light

	// load the basic shader
	basicShader, err := forward.CreateBasicShader()
	if err != nil {
		fmt.Printf("Failed to compile and link the basic shader program!\n%v", err)
		return
	}
	defer basicShader.Destroy()

	// load up some textures
	textureMan := fizzle.NewTextureManager()
	diffuseTex, err := textureMan.LoadTexture("landdiffuse", diffuseFile)
	if err != nil {
		fmt.Printf("Failed to load the diffuse texture at %s!\n%v", diffuseFile, err)
		return
	}

	// setup a shaderd material to use for the landscape
	landMaterial := fizzle.NewMaterial()
	landMaterial.Shader = basicShader
	landMaterial.DiffuseColor = mgl.Vec4{1, 1, 1, 1.0}
	landMaterial.Shininess = 0
	landMaterial.DiffuseTex = diffuseTex

	// create the landscape object
	landMesh, err := fizzle.CreateLandscapeFromFile(landscapeSize, landscapeSize, heightmapFile, mgl.Vec3{2, landscapeScale, 2})
	if err != nil {
		fmt.Printf("Failed to create the landscape mesh: %v.\n", err)
		return
	}

	landMesh.Material = landMaterial
	landMesh.Location[0] -= landscapeSize / 2.0
	landMesh.Location[1] -= landscapeScale / 2.0
	landMesh.Location[2] -= landscapeSize / 2.0

	// skybox shader
	skyShader, err := forward.CreateSkyboxShader()
	if err != nil {
		fmt.Printf("Failed to load the skybox shader: %v\n", err)
		return
	}

	// skybox
	skybox := fizzle.CreateSphere(landscapeSize*4.0, 64, 64)
	skybox.Material = fizzle.NewMaterial()
	skybox.Material.Shader = skyShader

	// setup the camera to look at the cube
	camera = fizzle.NewYawPitchCamera(mgl.Vec3{landscapeSize / 2.0, landscapeScale / 2.0, landscapeSize * 2.5})
	camera.UpdatePitch(math.Pi / 4.0)

	// set some OpenGL flags
	gfx.Enable(graphics.CULL_FACE)
	gfx.Enable(graphics.DEPTH_TEST)

	// loop until something told the mainWindow that it should close
	lastFrame := time.Now()
	for !mainWindow.ShouldClose() {
		// calculate the difference in time to control rotation speed
		thisFrame := time.Now()
		frameDelta = float32(thisFrame.Sub(lastFrame).Seconds())

		// handle any keyboard input
		kbModel.CheckKeyPresses()

		// clear the screen
		width, height := renderer.GetResolution()
		gfx.Viewport(0, 0, int32(width), int32(height))
		gfx.ClearColor(0.25, 0.25, 0.25, 1.0)
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		// make the projection and view matrixes
		perspective := mgl.Perspective(mgl.DegToRad(60.0), float32(width)/float32(height), 1.0, landscapeSize*5.0)
		view := camera.GetViewMatrix()

		// draw the landscape
		renderer.DrawRenderableWithMode(landMesh, landMesh.Material.Shader, nil, perspective, view, camera, graphics.TRIANGLE_STRIP)

		// draw the skybox
		gfx.Disable(graphics.CULL_FACE)
		skybox.Location = camera.GetPosition()
		renderer.DrawRenderable(skybox, nil, perspective, view, camera)
		gfx.Enable(graphics.CULL_FACE)

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
	glfw.WindowHint(glfw.Samples, 4)
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

func handleMoveForward() {
	delta := camera.GetForwardVector().Mul(frameDelta * movementScale)
	pos := camera.GetPosition().Add(delta)
	camera.SetPosition(pos[0], pos[1], pos[2])
}

func handleMoveBackward() {
	delta := camera.GetForwardVector().Mul(frameDelta * movementScale)
	pos := camera.GetPosition().Sub(delta)
	camera.SetPosition(pos[0], pos[1], pos[2])
}

func handleSlideLeft() {
	delta := camera.GetSideVector().Mul(frameDelta * movementScale)
	pos := camera.GetPosition().Sub(delta)
	camera.SetPosition(pos[0], pos[1], pos[2])
}

func handleSlideRight() {
	delta := camera.GetSideVector().Mul(frameDelta * movementScale)
	pos := camera.GetPosition().Add(delta)
	camera.SetPosition(pos[0], pos[1], pos[2])
}

func handleFloatUp() {
	delta := camera.GetUpVector().Mul(frameDelta * movementScale)
	pos := camera.GetPosition().Add(delta)
	camera.SetPosition(pos[0], pos[1], pos[2])
}

func handleFloatDown() {
	delta := camera.GetUpVector().Mul(frameDelta * movementScale)
	pos := camera.GetPosition().Sub(delta)
	camera.SetPosition(pos[0], pos[1], pos[2])
}

var (
	rotationRadsPerPixel = float32(2 * math.Pi / 2000.0)
	lastMouseX           = -1.0
	lastMouseY           = -1.0
)

func handleMouseMovement(w *glfw.Window, xpos float64, ypos float64) {
	if lastMouseX < 0 || lastMouseY < 0 {
		lastMouseX = xpos
		lastMouseY = ypos
		return
	}

	if w.GetMouseButton(glfw.MouseButtonRight) == glfw.Press {
		deltaX := float32(xpos - lastMouseX)
		deltaY := float32(ypos - lastMouseY)

		// modify the camera accordingly
		yaw := camera.GetYaw() + rotationRadsPerPixel*deltaX
		pitch := camera.GetPitch() + rotationRadsPerPixel*deltaY
		camera.SetYawAndPitch(yaw, pitch)
	}

	// update our tracking
	lastMouseX = xpos
	lastMouseY = ypos
}
