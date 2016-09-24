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

	"github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	input "github.com/tbogdala/fizzle/input/glfwinput"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

/*
	This example illustrates the lighting setup of lights in a forward rendering situation.
	This differs from the deferred rendering pass because the lighting setup in forward
	rendering is much more limited and and support less than a handfull of lights.

	Besides lights, this will also illustrate how to cast shadows.

*/

// GLFW event handling must run on the main OS thread.
func init() {
	runtime.LockOSThread()
}

const (
	windowWidth                = 1280
	windowHeight               = 720
	shadowTexSize              = 2048
	fov                        = 70.0
	radsPerSec                 = math.Pi / 4.0
	basicShaderPath            = "../assets/forwardshaders/basic"
	shadowmapTextureShaderPath = "../assets/forwardshaders/shadowmap_texture"
	shadowmapShaderPath        = "../assets/forwardshaders/shadowmap_generator"

	testDiffusePath = "../assets/textures/TestCube_D.png"
	testNormalsPath = "../assets/textures/TestCube_N.png"
)

var (
	// mainWindow is the main window of the application
	mainWindow *glfw.Window

	// renderer is the forward renderer used for this example
	renderer *forward.ForwardRenderer
)

// main is the entry point for the application.
func main() {
	// start off by initializing the GL and GLFW libraries and creating a window.
	// the default window size we use is 1280x720
	w, gfx := initGraphics("Forward Lighting", windowWidth, windowHeight)
	mainWindow = w

	// set the callback function for key input
	kbModel := input.NewKeyboardModel(mainWindow)
	kbModel.BindTrigger(glfw.KeyEscape, setShouldClose)
	kbModel.SetupCallbacks()

	// create a new renderer
	renderer = forward.NewForwardRenderer(gfx)
	renderer.ChangeResolution(windowWidth, windowHeight)
	defer renderer.Destroy()

	// setup the camera to look at the cube
	camera := fizzle.NewYawPitchCamera(mgl.Vec3{0.0, 5.0, 5.0})
	camera.SetYawAndPitch(0.0, mgl.DegToRad(60))

	// load the basic shader
	basicShader, err := fizzle.LoadShaderProgramFromFiles(basicShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the basic shader program!\n%v", err)
		os.Exit(1)
	}
	defer basicShader.Destroy()

	// load the shader used to draw the shadowmap as a texture in the UI
	shadowmapTextureShader, err := fizzle.LoadShaderProgramFromFiles(shadowmapTextureShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the shadowmap texture shader program!\n%v", err)
		os.Exit(1)
	}
	defer shadowmapTextureShader.Destroy()

	// loadup the shadowmap shader used to generate the shadows
	shadowmapShader, err := fizzle.LoadShaderProgramFromFiles(shadowmapShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the shadowmap generator shader program!\n%v", err)
		os.Exit(1)
	}
	defer shadowmapShader.Destroy()

	// load up some textures
	textureMan := fizzle.NewTextureManager()

	diffuseTex, err := textureMan.LoadTexture("cube_diffuse", testDiffusePath)
	if err != nil {
		fmt.Printf("Failed to load the diffuse texture at %s!\n%v", testDiffusePath, err)
		os.Exit(1)
	}
	fmt.Printf("Loaded the diffuse texture at %s(%d).\n", testDiffusePath, diffuseTex)

	normalsTex, err := textureMan.LoadTexture("cube_diffuse", testNormalsPath)
	if err != nil {
		fmt.Printf("Failed to load the normals texture at %s!\n%v", testNormalsPath, err)
		os.Exit(1)
	}
	fmt.Printf("Loaded the normals texture at %s(%d).\n", testNormalsPath, normalsTex)

	// create the floor plane
	floorMaterial := fizzle.NewMaterial()
	floorMaterial.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	floorMaterial.SpecularColor = mgl.Vec4{0.3, 0.3, 0.3, 1.0}
	floorMaterial.Shininess = 3.0
	floorMaterial.DiffuseTex = diffuseTex
	floorMaterial.NormalsTex = normalsTex
	floorMaterial.Shader = basicShader

	floorPlane := fizzle.CreatePlaneXZ(-0.5, 0.5, 0.5, -0.5)
	floorPlane.Scale = mgl.Vec3{10, 10, 10}
	floorPlane.Material = floorMaterial

	// create the test cube to rotate
	cubeMaterial := fizzle.NewMaterial()
	cubeMaterial.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	cubeMaterial.SpecularColor = mgl.Vec4{0.3, 0.3, 0.3, 1.0}
	cubeMaterial.Shininess = 6.0
	cubeMaterial.DiffuseTex = diffuseTex
	cubeMaterial.NormalsTex = normalsTex
	cubeMaterial.Shader = basicShader

	testCube := fizzle.CreateCube(-0.5, -0.5, -0.5, 0.5, 0.5, 0.5)
	testCube.Location = mgl.Vec3{-2.5, 1.0, 0.0}
	testCube.Material = cubeMaterial

	// enable shadow mapping in the renderer
	renderer.SetupShadowMapRendering()

	// add light #1
	light := renderer.NewPointLight(mgl.Vec3{5.0, 3.0, 5.0})
	light.DiffuseColor = mgl.Vec4{0.9, 0.9, 0.9, 1.0}
	light.Strength = 5.0
	renderer.ActiveLights[0] = light
	light.CreateShadowMap(shadowTexSize, 0.5, 50.0, mgl.Vec3{-5.0, -3.0, -5.0})

	// add light #2
	light2 := renderer.NewPointLight(mgl.Vec3{-2.0, 3.0, 3.0})
	light2.DiffuseColor = mgl.Vec4{0.9, 0.0, 0.0, 1.0}
	light2.DiffuseIntensity = 1.00
	light2.AmbientIntensity = 0.00
	light2.Strength = 1.0
	renderer.ActiveLights[1] = light2
	light2.CreateShadowMap(shadowTexSize, 0.5, 50.0, mgl.Vec3{2.0, -3.0, -3.0})

	// make a UI image to show the shadowmap texture, scaled down
	shadowMapUIMat := fizzle.NewMaterial()
	shadowMapUIMat.Shader = shadowmapTextureShader
	shadowMapUIMat.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}

	shadowMapUIQuad := fizzle.CreatePlaneXY(0, 0, 256, 256)
	shadowMapUIQuad.Material = shadowMapUIMat

	// set some OpenGL flags
	gfx.Enable(graphics.CULL_FACE)
	gfx.Enable(graphics.DEPTH_TEST)
	gfx.Enable(graphics.TEXTURE_2D)
	gfx.Enable(graphics.BLEND)

	// loop until something told the mainWindow that it should close
	lastFrame := time.Now()
	for !mainWindow.ShouldClose() {
		// calculate the difference in time to control rotation speed
		thisFrame := time.Now()
		frameDelta := float32(thisFrame.Sub(lastFrame).Seconds())

		// rotate the cube around the Y axis at a speed of 0.5*math.Pi / sec
		rotDelta := mgl.QuatRotate(0.5*math.Pi*frameDelta, mgl.Vec3{0.0, 1.0, 0.0})
		testCube.LocalRotation = testCube.LocalRotation.Mul(rotDelta)

		// Shadow time!
		renderer.StartShadowMapping()
		lightCount := renderer.GetActiveLightCount()
		if lightCount >= 1 {
			for lightI := 0; lightI < lightCount; lightI++ {
				// get lights with shadow maps
				lightToCast := renderer.ActiveLights[lightI]
				if lightToCast.ShadowMap == nil {
					continue
				}

				// enable the light to cast shadows
				renderer.EnableShadowMappingLight(lightToCast)
				renderer.DrawRenderableWithShader(testCube, shadowmapShader, nil, lightToCast.ShadowMap.Projection, lightToCast.ShadowMap.View, camera)
				renderer.DrawRenderableWithShader(floorPlane, shadowmapShader, nil, lightToCast.ShadowMap.Projection, lightToCast.ShadowMap.View, camera)

			}
		}

		// stop the shadow generation
		renderer.EndShadowMapping()

		// clear the screen and reset our viewport
		width, height := renderer.GetResolution()
		gfx.Viewport(0, 0, int32(width), int32(height))
		gfx.ClearColor(0.05, 0.05, 0.05, 1.0)
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		// make the projection and view matrixes
		perspective := mgl.Perspective(mgl.DegToRad(fov), float32(width)/float32(height), 1.0, 100.0)
		view := camera.GetViewMatrix()

		// draw the stuff
		renderer.DrawRenderable(testCube, nil, perspective, view, camera)
		renderer.DrawRenderable(floorPlane, nil, perspective, view, camera)

		// for this test, render a quad showing the shadowmap texture
		renderShadowMapUITex(shadowMapUIQuad, renderer.ActiveLights[0].ShadowMap.Texture)

		// draw the screen
		mainWindow.SwapBuffers()
		glfw.PollEvents()

		// update our last frame time
		lastFrame = thisFrame
	}
}

func renderShadowMapUITex(r *fizzle.Renderable, shadow graphics.Texture) {
	gfx := fizzle.GetGraphics()

	gfx.BindTexture(graphics.TEXTURE_2D, shadow)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_COMPARE_MODE, graphics.NONE)
	gfx.BindTexture(graphics.TEXTURE_2D, 0)

	r.Material.CustomTex[0] = shadow
	width, height := renderer.GetResolution()
	ortho := mgl.Ortho(0, float32(width), 0, float32(height), -10, 10)
	view := mgl.Ident4()
	renderer.DrawRenderable(r, nil, ortho, view, nil)

	// reset the shadow map textures used for visualization
	gfx.BindTexture(graphics.TEXTURE_2D, shadow)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_COMPARE_MODE, graphics.COMPARE_REF_TO_TEXTURE)
	gfx.BindTexture(graphics.TEXTURE_2D, 0)

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
