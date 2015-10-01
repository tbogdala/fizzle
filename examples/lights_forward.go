// Copyright 2015, Timothy` Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	gl "github.com/go-gl/gl/v3.3-core/gl"
	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	//	"github.com/tbogdala/gombz"
	"github.com/tbogdala/fizzle"
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
	width                      = 800
	height                     = 600
	shadowTexSize              = 1024
	fov                        = 70.0
	radsPerSec                 = math.Pi / 4.0
	diffuseTexBumpedShaderPath = "./assets/forwardshaders/diffuse_texbumped_shadows"
	shadowmapShaderPath        = "./assets/forwardshaders/shadowmap_generator"

	testDiffusePath = "./assets/textures/TestCube_D.png"
	testNormalsPath = "./assets/textures/TestCube_N.png"
)

// main is the entry point for the application.
func main() {
	// start off by initializing the GL and GLFW libraries and creating a window.
	// the default window size we use is 800x600
	mainWindow := initGraphics("Forward Lighting", width, height)

	// set the callback function for key input
	mainWindow.SetKeyCallback(keyCallback)

	// create a new renderer
	renderer := fizzle.NewForwardRenderer(mainWindow)
	defer renderer.Destroy()

	// setup the camera to look at the cube
	camera := fizzle.NewCamera(mgl.Vec3{0.0, 5.0, 5.0})
	camera.SetYawAndPitch(0.0, mgl.DegToRad(60))

	// load the diffuse, textured and normal mapped shader
	diffuseTexBumpedShader, err := fizzle.LoadShaderProgramFromFiles(diffuseTexBumpedShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the diffuse shader program!\n%v", err)
		os.Exit(1)
	}
	defer diffuseTexBumpedShader.Destroy()

	// loadup the shadowmap shaders
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
	floorPlane := fizzle.CreatePlaneXZ("diffuse_texbumped", -0.5, 0.5, 0.5, -0.5)
	floorPlane.Scale = mgl.Vec3{10, 10, 10}
	floorPlane.Core.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	floorPlane.Core.SpecularColor = mgl.Vec4{0.3, 0.3, 0.3, 1.0}
	floorPlane.Core.Shininess = 3.0
	floorPlane.Core.Tex0 = diffuseTex
	floorPlane.Core.Tex1 = normalsTex
	floorPlane.Core.Shader = diffuseTexBumpedShader

	// create the test cube to rotate
	testCube := fizzle.CreateCube("diffuse_texbumped", -0.5, -0.5, -0.5, 0.5, 0.5, 0.5)
	testCube.Core.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	testCube.Core.SpecularColor = mgl.Vec4{0.3, 0.3, 0.3, 1.0}
	testCube.Location = mgl.Vec3{-2.5, 1.0, 0.0}
	testCube.Core.Shininess = 6.0
	testCube.Core.Tex0 = diffuseTex
	testCube.Core.Tex1 = normalsTex
	testCube.Core.Shader = diffuseTexBumpedShader

	// enable shadow mapping in the renderer
	renderer.SetupShadowMapRendering()

	// add light #1
	light := fizzle.NewLight()
	light.Position = mgl.Vec3{5.0, 3.0, 5.0}
	light.DiffuseColor = mgl.Vec4{0.9, 0.9, 0.9, 1.0}
	light.DiffuseIntensity = 5.00
	light.AmbientIntensity = 0.20
	light.Attenuation = 0.2
	renderer.ActiveLights[0] = light
	light.CreateShadowMap(shadowTexSize, 0.5, 50.0, mgl.Vec3{-5.0, -3.0, -5.0})

	// add light #2
	light = fizzle.NewLight()
	light.Position = mgl.Vec3{-2.0, 3.0, 3.0}
	light.DiffuseColor = mgl.Vec4{0.9, 0.0, 0.0, 1.0}
	light.DiffuseIntensity = 1.00
	light.AmbientIntensity = 0.00
	light.Attenuation = 0.2
	renderer.ActiveLights[1] = light

	// set some OpenGL flags
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.TEXTURE_2D)
	gl.Enable(gl.BLEND)

	// loop until something told the mainWindow that it should close
	lastFrame := time.Now()
	for !mainWindow.ShouldClose() {
		// calculate the difference in time to control rotation speed
		thisFrame := time.Now()
		frameDelta := float32(thisFrame.Sub(lastFrame).Seconds())

		// rotate the cube and sphere around the Y axis at a speed of 0.5*math.Pi / sec
		rotDelta := mgl.QuatRotate(0.5*math.Pi*frameDelta, mgl.Vec3{0.0, 1.0, 0.0})
		testCube.LocalRotation = testCube.LocalRotation.Mul(rotDelta)

		// Shadow time!
		renderer.StartShadowMapping()
		lightCount := renderer.GetActiveLightCount()
		if lightCount >= 1 {
			for lightI := 0; lightI < lightCount; lightI++ {
				// get lights with shadow maps
				light := renderer.ActiveLights[lightI]
				if light.ShadowMap == nil {
					continue
				}

				// enable the light to cast shadows
				renderer.EnableShadowMappingLight(light)
				renderer.DrawRenderableWithShader(testCube, shadowmapShader, nil, light.ShadowMap.Projection, light.ShadowMap.View)
				renderer.DrawRenderableWithShader(floorPlane, shadowmapShader, nil, light.ShadowMap.Projection, light.ShadowMap.View)
			}
		}
		renderer.EndShadowMapping()

		// clear the screen and reset our viewport
		gl.Viewport(0, 0, int32(width), int32(height))
		gl.ClearColor(0.05, 0.05, 0.05, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// make the projection and view matrixes
		perspective := mgl.Perspective(mgl.DegToRad(fov), float32(width)/float32(height), 1.0, 100.0)
		view := camera.GetViewMatrix()

		// draw the stuff
		renderer.DrawRenderable(testCube, nil, perspective, view)
		renderer.DrawRenderable(floorPlane, nil, perspective, view)

		// draw the screen
		mainWindow.SwapBuffers()
		glfw.PollEvents()

		// update our last frame time
		lastFrame = thisFrame
	}
}

// initGraphics creates an OpenGL window and initializes the required graphics libraries.
// It will either succeed or panic.
func initGraphics(title string, w int, h int) *glfw.Window {
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

	// disable v-sync for max draw rate
	glfw.SwapInterval(0)

	// make sure that all of the GL functions are initialized
	err = gl.Init()
	if err != nil {
		panic("Failed to initialize GL! " + err.Error())
	}

	return mainWindow
}

// keyCallback is set as a callback in main() and is used to close the window
// when the escape key is hit.
func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}
}
