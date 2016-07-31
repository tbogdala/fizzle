// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
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

	gui "github.com/tbogdala/eweygewey"
	guiinput "github.com/tbogdala/eweygewey/glfwinput"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	input "github.com/tbogdala/fizzle/input/glfwinput"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

const (
	fontScale    = 14
	fontFilepath = "../assets/Oswald-Heavy.ttf"
	fontGlyphs   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890., :[]{}\\|<>;\"'~`?/-+_=()*&^%$#@!"

	windowWidth       = 1280
	windowHeight      = 720
	radsPerSec        = math.Pi / 4.0
	diffuseShaderPath = "../assets/forwardshaders/basic"
)

var (
	// renderCube indicates if the cube should be drawn or the sphere
	renderCube = true
	rotateObj  = true

	mainWindow *glfw.Window
	renderer   *forward.ForwardRenderer
	uiman      *gui.Manager
)

// GLFW must run on the same OS thread.
func init() {
	runtime.LockOSThread()
}

// main is the entry point for the application.
func main() {
	// -------------------------------------------------------------------------
	// Window and GUI creation
	// -------------------------------------------------------------------------

	// start off by initializing the GL and GLFW libraries and creating a window.
	w, gfx := initGraphics("Shader Explorer", windowWidth, windowHeight)
	mainWindow = w

	// create and initialize the gui Manager
	uiman = gui.NewManager(gfx)
	err := uiman.Initialize(gui.VertShader330, gui.FragShader330, windowWidth, windowHeight, windowHeight)
	if err != nil {
		panic("Failed to initialize the user interface! " + err.Error())
	}
	guiinput.SetInputHandlers(uiman, mainWindow)

	// load a font
	_, err = uiman.NewFont("Default", fontFilepath, fontScale, fontGlyphs)
	if err != nil {
		panic("Failed to load the font file! " + err.Error())
	}

	// set the callback functions for key input
	kbModel := input.NewKeyboardModel(mainWindow)
	kbModel.BindTrigger(glfw.KeyEscape, setShouldClose)
	kbModel.BindTrigger(glfw.KeySpace, toggleModel)
	kbModel.SetupCallbacks()

	// -------------------------------------------------------------------------
	// Renderer and scene creation
	// -------------------------------------------------------------------------

	// create a new renderer
	renderer = forward.NewForwardRenderer(gfx)
	renderer.ChangeResolution(windowWidth, windowHeight)
	defer renderer.Destroy()

	// load the diffuse shader
	diffuseShader, err := fizzle.LoadShaderProgramFromFiles(diffuseShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the diffuse shader program!\n%v", err)
		os.Exit(1)
	}
	defer diffuseShader.Destroy()

	// put a light in there
	light := renderer.NewLight()
	light.Position = mgl.Vec3{-10.0, 5.0, 10}
	light.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	//light.Direction = mgl.Vec3{1.0, -0.5, -1.0}
	light.DiffuseIntensity = 0.70
	light.SpecularIntensity = 0.10
	light.AmbientIntensity = 0.20
	light.ConstAttenuation = 0.20
	light.LinearAttenuation = 0.20
	light.QuadraticAttenuation = 0.20
	light.Strength = 10.0
	renderer.ActiveLights[0] = light

	// create a 2x2x2 cube to render
	cube := fizzle.CreateCube("diffuse", -1, -1, -1, 1, 1, 1)
	cube.Core.Shader = diffuseShader
	cube.Core.DiffuseColor = mgl.Vec4{0.9, 0.05, 0.05, 1.0}
	cube.Core.SpecularColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	cube.Core.Shininess = 10.0

	// create a sphere to render
	sphere := fizzle.CreateSphere("diffuse", 1, 16, 16)
	sphere.Core.Shader = diffuseShader
	sphere.Core.DiffuseColor = mgl.Vec4{0.9, 0.05, 0.05, 1.0}
	sphere.Core.SpecularColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	sphere.Core.Shininess = 10.0

	renderCube = false

	// setup the camera to look at the cube
	camera := fizzle.NewOrbitCamera(mgl.Vec3{0, 0, 0}, math.Pi/2.0, 5.0, math.Pi/2.0)

	// set some OpenGL flags
	gfx.BlendEquation(graphics.FUNC_ADD)
	gfx.BlendFunc(graphics.SRC_ALPHA, graphics.ONE_MINUS_SRC_ALPHA)
	gfx.Enable(graphics.BLEND)
	gfx.Enable(graphics.CULL_FACE)
	gfx.Enable(graphics.DEPTH_TEST)

	// -------------------------------------------------------------------------
	// Create the windows to manage the properties of the shader and lights
	// -------------------------------------------------------------------------
	shininess := float32(10.0)
	color := [4]int{255, 0, 0, 255}
	specular := [4]int{255, 255, 255, 255}

	materialWindow := uiman.NewWindow("Property", 0.01, 0.85, 0.2, 0.25, func(wnd *gui.Window) {
		const colWidth = 0.33
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("Diffuse")
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderInt("ColorR", &color[0], 0, 255)
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderInt("ColorG", &color[1], 0, 255)
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderInt("ColorB", &color[2], 0, 255)
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderInt("ColorA", &color[3], 0, 255)

		wnd.StartRow()
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("Specular")
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderInt("SpecR", &specular[0], 0, 255)
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderInt("SpecG", &specular[1], 0, 255)
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderInt("SpecB", &specular[2], 0, 255)
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderInt("SpecA", &specular[3], 0, 255)

		wnd.StartRow()
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("Shininess")
		wnd.DragSliderUFloat("Shininess", 0.1, &shininess)
	})
	materialWindow.Title = "Material Properties"
	materialWindow.ShowTitleBar = true
	materialWindow.IsMoveable = true
	materialWindow.AutoAdjustHeight = false
	materialWindow.ShowScrollBar = true
	materialWindow.IsScrollable = true

	lightWindow := uiman.NewWindow("Light", 0.7, 0.85, 0.2, 0.5, func(wnd *gui.Window) {
		const colWidth = 0.33
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("Diffuse")
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderFloat("LDiffR", &light.DiffuseColor[0], 0, 1)
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderFloat("LDiffG", &light.DiffuseColor[1], 0, 1)
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderFloat("LDiffB", &light.DiffuseColor[2], 0, 1)
		wnd.RequestItemWidthMax(0.165)
		wnd.SliderFloat("LDiffA", &light.DiffuseColor[3], 0, 1)

		wnd.StartRow()
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("Strength")
		wnd.DragSliderUFloat("LStr", 0.01, &light.Strength)

		wnd.StartRow()
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("AmbientF")
		wnd.DragSliderUFloat("LAmb", 0.01, &light.AmbientIntensity)

		wnd.StartRow()
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("DiffuseF")
		wnd.DragSliderUFloat("LDiff", 0.01, &light.DiffuseIntensity)

		wnd.StartRow()
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("SpecularF")
		wnd.DragSliderUFloat("LSpec", 0.01, &light.SpecularIntensity)

		wnd.StartRow()
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("Const AttenuationF")
		wnd.DragSliderUFloat("LConstAtt", 0.01, &light.ConstAttenuation)

		wnd.StartRow()
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("Linear AttenuationF")
		wnd.DragSliderUFloat("LLinearAtt", 0.01, &light.LinearAttenuation)

		wnd.StartRow()
		wnd.RequestItemWidthMin(colWidth)
		wnd.Text("Quadratic AttenuationF")
		wnd.DragSliderUFloat("LQuadAtt", 0.01, &light.QuadraticAttenuation)
	})
	lightWindow.Title = "Light Properties"
	lightWindow.ShowTitleBar = true
	lightWindow.IsMoveable = true
	lightWindow.AutoAdjustHeight = false
	lightWindow.ShowScrollBar = true
	lightWindow.IsScrollable = true

	objCtrlWindow := uiman.NewWindow("ObjControl", 0.01, 0.99, 0.2, 0.35, func(wnd *gui.Window) {
		wnd.Checkbox("RotateObjs", &rotateObj)
		wnd.Text("Rotate Object")

		// setup the controls to switch between spawnwers
		wnd.StartRow()
		prevPressed, _ := wnd.Button("prevSpawner", " < ")
		nextPressed, _ := wnd.Button("nextSpawner", " > ")
		if prevPressed || nextPressed {
			// we only have two models for now, so if the button
			// is pressed, then toggle the model to draw
			renderCube = !renderCube
		}

		if renderCube {
			wnd.Text("Cube")
		} else {
			wnd.Text("Sphere")
		}

	})
	objCtrlWindow.ShowTitleBar = false
	objCtrlWindow.IsMoveable = false
	objCtrlWindow.AutoAdjustHeight = true

	// -------------------------------------------------------------------------
	// Main loop
	// -------------------------------------------------------------------------

	// loop until something told the mainWindow that it should close
	lastFrame := time.Now()
	for !mainWindow.ShouldClose() {
		// calculate the difference in time to control rotation speed
		thisFrame := time.Now()
		frameDelta := float32(thisFrame.Sub(lastFrame).Seconds())

		// handle any keyboard input
		kbModel.CheckKeyPresses()

		// rotate the cube and sphere around the Y axis at a speed of radsPerSec
		if rotateObj {
			rotDelta := mgl.QuatRotate(radsPerSec*frameDelta, mgl.Vec3{0.0, 1.0, 0.0})
			cube.LocalRotation = cube.LocalRotation.Mul(rotDelta)
			sphere.LocalRotation = sphere.LocalRotation.Mul(rotDelta)
		}

		// clear the screen
		width, height := renderer.GetResolution()
		gfx.Viewport(0, 0, int32(width), int32(height))
		gfx.ClearColor(0.25, 0.25, 0.25, 1.0)
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		// make the projection and view matrixes
		perspective := mgl.Perspective(mgl.DegToRad(60.0), float32(width)/float32(height), 1.0, 100.0)
		view := camera.GetViewMatrix()

		// draw the cube or the sphere
		var activeCore *fizzle.RenderableCore
		if renderCube {
			activeCore = cube.Core
			renderer.DrawRenderable(cube, nil, perspective, view, camera)
		} else {
			activeCore = sphere.Core
			renderer.DrawRenderable(sphere, nil, perspective, view, camera)
		}

		// update the material on the active object with the values from the editor
		for i := 0; i < 4; i++ {
			activeCore.DiffuseColor[i] = float32(color[i]) / 255.0
			activeCore.SpecularColor[i] = float32(specular[i]) / 255.0
		}
		activeCore.Shininess = shininess

		// draw the user interface
		uiman.Construct(float64(frameDelta))
		uiman.Draw()

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

// toggleModel sets whether or not the cube or the sphere should be rendered.
func toggleModel() {
	// spacebar toggles the drawing of the cube or the sphere
	if renderCube {
		renderCube = false
	} else {
		renderCube = true
	}
}
