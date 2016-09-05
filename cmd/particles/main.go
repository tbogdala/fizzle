// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"flag"
	"fmt"
	"math"
	"runtime"
	"time"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	gui "github.com/tbogdala/eweygewey"
	guiinput "github.com/tbogdala/eweygewey/glfwinput"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	particles "github.com/tbogdala/fizzle/particles"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

var (
	windowWidth     = 1280
	windowHeight    = 720
	mainWindow      *glfw.Window
	uiman           *gui.Manager
	renderer        *forward.ForwardRenderer
	textureFilepath = "../../examples/assets/textures/explosion00.png"
	colorShader     = "../../examples/assets/forwardshaders/color"
)

const (
	fontScale    = 14
	fontFilepath = "../../examples/assets/Oswald-Heavy.ttf"
	fontGlyphs   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890., :[]{}\\|<>;\"'~`?/-+_=()*&^%$#@!"
)

// block of flags set on the command line
var (
	flagDesktopNumber int
)

// spawnerPrototypes keeps track of possible spawner interface implementations
// to switch between
type spawnerPrototypes struct {
	Name string
	particles.ParticleSpawner
	RenderUI func(wnd *gui.Window)
}

var (
	// created instances of particle spawners
	knownSpawners []spawnerPrototypes
)

// GLFW event handling must run on the main OS thread. If this doesn't get
// locked down, you will likely see random crashes on memory access while
// running the application after a few seconds.
//
// So on initialization of the module, lock the OS thread for this goroutine.
func init() {
	runtime.LockOSThread()
	flag.IntVar(&flagDesktopNumber, "desktop", -1, "the index of the desktop to create the main window on")
}

// initSpawners create prototype instances of all known spawner types
func initSpawners() {
	knownSpawners = []spawnerPrototypes{}

	cone := particles.NewConeSpawner(nil, 0.5, 1, 1)
	knownSpawners = append(knownSpawners, spawnerPrototypes{Name: cone.GetName(), ParticleSpawner: cone, RenderUI: func(wnd *gui.Window) {
		const textWidth = 0.33
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Top Radius")
		wnd.DragSliderUFloat("tradius", 0.1, &cone.TopRadius)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Bottom Radius")
		wnd.DragSliderUFloat("bradius", 0.1, &cone.BottomRadius)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Length")
		wnd.DragSliderUFloat("conelength", 0.1, &cone.Length)
	}})

	cube := particles.NewCubeSpawner(nil, mgl.Vec3{-1, -1, -1}, mgl.Vec3{1, 1, 1})
	knownSpawners = append(knownSpawners, spawnerPrototypes{Name: cube.GetName(), ParticleSpawner: cube, RenderUI: func(wnd *gui.Window) {
		const textWidth = 0.33
		const width3Col = 0.22

		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Bottom Left")
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("cubebl1", 0.1, &cube.BottomLeft[0])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("cubebl2", 0.1, &cube.BottomLeft[1])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("cubebl3", 0.1, &cube.BottomLeft[2])

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Top Right")
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("cubetr1", 0.1, &cube.TopRight[0])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("cubetr2", 0.1, &cube.TopRight[1])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("cubetr3", 0.1, &cube.TopRight[2])
	}})
}

// getSpawnerIndex returns the slice index within known spawners for a given spawner interface instance
func getSpawnerIndex(spawner particles.ParticleSpawner) int {
	for i, s := range knownSpawners {
		if s.Name == spawner.GetName() {
			return i
		}
	}

	return 0
}

func main() {
	// parse the command line options
	flag.Parse()

	// start off by initializing the GL and GLFW libraries and creating a window.
	w, gfx := initGraphics("Particle Editor", windowWidth, windowHeight)
	mainWindow = w

	// create prototype instances of all known spawner types
	initSpawners()

	/////////////////////////////////////////////////////////////////////////////
	// create and initialize the gui Manager
	uiman = gui.NewManager(gfx)
	err := uiman.Initialize(gui.VertShader330, gui.FragShader330, int32(windowWidth), int32(windowHeight), int32(windowHeight))
	if err != nil {
		panic("Failed to initialize the user interface! " + err.Error())
	}
	guiinput.SetInputHandlers(uiman, mainWindow)

	// load a font
	_, err = uiman.NewFont("Default", fontFilepath, fontScale, fontGlyphs)
	if err != nil {
		panic("Failed to load the font file! " + err.Error())
	}

	/////////////////////////////////////////////////////////////////////////////
	// make a window that will render the particle system
	const particleWindowSize = 512
	customMargin := mgl.Vec4{0, 0, 0, 0}
	customWindowWS, customWindowHS := uiman.DisplayToScreen(particleWindowSize+8, particleWindowSize+8) // offset by 8 for windowPadding
	customWS, customHS := uiman.DisplayToScreen(particleWindowSize, particleWindowSize)

	renderer = forward.NewForwardRenderer(gfx)
	renderer.ChangeResolution(particleWindowSize, particleWindowSize)
	defer renderer.Destroy()

	// load the particle shader
	particleShader, err := fizzle.LoadShaderProgram(particles.VertShader330, particles.FragShader330, nil)
	if err != nil {
		panic("Failed to compile and link the particle shader program! " + err.Error())
	}
	defer particleShader.Destroy()

	// load the color shader
	colorShader, err := fizzle.LoadShaderProgramFromFiles(colorShader, nil)
	if err != nil {
		panic("Failed to compile and link the color shader program! " + err.Error())
	}
	defer colorShader.Destroy()

	// create a particle system
	particleSystem := particles.NewSystem(gfx)
	emitter := particleSystem.NewEmitter(nil)
	emitter.Properties.TextureFilepath = textureFilepath
	emitter.Properties.MaxParticles = 300
	emitter.Properties.SpawnRate = 40
	emitter.Properties.Size = 32.0
	emitter.Properties.Color = mgl.Vec4{0.0, 0.9, 0.0, 1.0}
	emitter.Properties.Velocity = mgl.Vec3{0, 1, 0}
	emitter.Properties.Acceleration = mgl.Vec3{0, -0.1, 0}
	emitter.Properties.TTL = 3.0
	emitter.Shader = particleShader.Prog

	// load the texture
	err = emitter.LoadTexture()
	if err != nil {
		panic(err.Error())
	}

	// reset the spawner to the first known spawner instance
	emitter.Spawner = knownSpawners[0]
	emitter.Spawner.SetOwner(emitter)

	// setup the camera to look at the cube
	camera := fizzle.NewOrbitCamera(mgl.Vec3{0, 0, 0}, math.Pi/2.0, 5.0, math.Pi/2.0)

	// now create the window itself
	customWindow := uiman.NewWindow("Particle Output", 0.01, 0.99, customWindowWS, customWindowHS, func(wnd *gui.Window) {
		wnd.Custom(customWS, customHS, customMargin, func() {
			// rotate the cube and sphere around the Y axis at a speed of radsPerSec
			gfx.ClearColor(0.0, 0.0, 0.0, 1.0)
			gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

			perspective := mgl.Perspective(mgl.DegToRad(60.0), float32(particleWindowSize)/float32(particleWindowSize), 0.1, 50.0)
			view := camera.GetViewMatrix()
			particleSystem.Draw(perspective, view)

			// draw the emitter volumes
			for _, e := range particleSystem.Emitters {
				e.Spawner.CreateRenderable()
				e.Spawner.DrawSpawnVolume(renderer, colorShader, perspective, view, camera)
			}
		})
	})
	customWindow.Title = "Particle Output"
	customWindow.ShowTitleBar = true
	customWindow.Style.WindowPadding = mgl.Vec4{4, 4, 4, 4}

	// create a window for editing the emitter properites
	var yaw, pitch, roll int
	propertyWindow := uiman.NewWindow("Emitter", 0.5, 0.99, 0.45, 0.75, func(wnd *gui.Window) {
		const textWidth = 0.33
		const width4Col = 0.165
		const width3Col = 0.22
		props := &emitter.Properties

		wnd.StartRow()
		wnd.Checkbox("isAlive", &emitter.Owner.IsActive)
		wnd.Text("Is Alive")
		wnd.Space(0.05)
		wnd.Checkbox("isEmitting", &emitter.Owner.IsEmitting)
		wnd.Text("Is Emitting")

		// setup the controls to switch between spawnwers
		wnd.Separator()
		wnd.RequestItemWidthMin(0.1)
		prevPressed, _ := wnd.Button("prevSpawner", "<")
		wnd.RequestItemWidthMin(0.1)
		nextPressed, _ := wnd.Button("nextSpawner", ">")
		ki := getSpawnerIndex(emitter.Spawner)
		if prevPressed {
			if ki > 0 {
				emitter.Spawner = knownSpawners[ki-1].ParticleSpawner
				emitter.Spawner.SetOwner(emitter)
				ki--
			}
		}
		if nextPressed {
			if ki < len(knownSpawners)-1 {
				emitter.Spawner = knownSpawners[ki+1].ParticleSpawner
				emitter.Spawner.SetOwner(emitter)
				ki++
			}
		}
		wnd.Text(emitter.Spawner.GetName())

		// render the spawner interface
		wnd.StartRow()
		knownSpawners[ki].RenderUI(wnd)

		wnd.Separator()
		wnd.RequestItemWidthMin(textWidth)
		loadBillboardPressed, _ := wnd.Button("LoadTexture", "Load Texture")
		wnd.Editbox("textureedit", &props.TextureFilepath)
		if loadBillboardPressed {
			err := emitter.LoadTexture()
			if err != nil {
				fmt.Printf("Failed to load new texture: %v\n", err)
			}
		}

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Max Particles")
		wnd.DragSliderUInt("maxparticles", 0.5, &props.MaxParticles)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Spawn Rate")
		wnd.DragSliderUInt("spawnrate", 0.5, &props.SpawnRate)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("TTL")
		wnd.DragSliderUFloat64("ttl", 0.1, &props.TTL)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Size")
		wnd.DragSliderUFloat("size", 0.1, &props.Size)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Speed")
		wnd.DragSliderUFloat("speed", 0.1, &props.Speed)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Color")
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("color1", &props.Color[0], 0.0, 1.0)
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("color2", &props.Color[1], 0.0, 1.0)
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("color3", &props.Color[2], 0.0, 1.0)
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("color4", &props.Color[3], 0.0, 1.0)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Origin")
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("origin1", 0.1, &props.Origin[0])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("origin2", 0.1, &props.Origin[1])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("origin3", 0.1, &props.Origin[2])

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("YawPitchRoll") // Y, X, Z order
		wnd.RequestItemWidthMax(width3Col)
		wnd.SliderInt("yaw", &yaw, 0, 359)
		wnd.RequestItemWidthMax(width3Col)
		wnd.SliderInt("pitch", &pitch, 0, 359)
		wnd.RequestItemWidthMax(width3Col)
		wnd.SliderInt("roll", &roll, 0, 359)

		// set the rotation based on the UI selected ypr angle values (in degrees)
		props.Rotation = mgl.AnglesToQuat(mgl.DegToRad(float32(pitch)), mgl.DegToRad(float32(yaw)), mgl.DegToRad(float32(roll)), mgl.XYZ)
	})
	propertyWindow.Title = "Emitter Properties"
	propertyWindow.ShowTitleBar = true
	propertyWindow.IsMoveable = true
	propertyWindow.AutoAdjustHeight = true
	//propertyWindow.ShowScrollBar = true
	//propertyWindow.IsScrollable = true

	/////////////////////////////////////////////////////////////////////////////
	// loop until something told the mainWindow that it should close
	// set some OpenGL flags
	gfx.Enable(graphics.CULL_FACE)
	gfx.Enable(graphics.DEPTH_TEST)
	gfx.Enable(graphics.PROGRAM_POINT_SIZE)
	gfx.Enable(graphics.BLEND)
	gfx.BlendFunc(graphics.SRC_ALPHA, graphics.ONE_MINUS_SRC_ALPHA)

	lastFrame := time.Now()
	for !mainWindow.ShouldClose() {
		// calculate the difference in time to control rotation speed
		thisFrame := time.Now()
		frameDelta := thisFrame.Sub(lastFrame).Seconds()

		// update the data for the application
		particleSystem.Update(frameDelta)

		// clear the screen
		//width, height := renderer.GetResolution()
		gfx.Viewport(0, 0, int32(windowWidth), int32(windowHeight))
		clearColor := gui.ColorIToV(114, 144, 154, 255)
		gfx.ClearColor(clearColor[0], clearColor[1], clearColor[2], clearColor[3])
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		// draw the user interface
		uiman.Construct(frameDelta)
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
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// get a list of all the monitors to use and then take the one
	// specified by the command line flag
	monitors := glfw.GetMonitors()
	if flagDesktopNumber >= len(monitors) {
		flagDesktopNumber = -1
	}
	var monitorToUse *glfw.Monitor
	if flagDesktopNumber >= 0 {
		monitorToUse = monitors[flagDesktopNumber]
	}

	// do the actual window creation
	mainWindow, err = glfw.CreateWindow(w, h, title, monitorToUse, nil)
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

// onWindowResize is called when the window changes size
func onWindowResize(w *glfw.Window, width int, height int) {
	uiman.AdviseResolution(int32(width), int32(height))
	//renderer.ChangeResolution(int32(width), int32(height))
}
