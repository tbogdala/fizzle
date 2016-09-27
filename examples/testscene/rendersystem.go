// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	forward "github.com/tbogdala/fizzle/renderer/forward"
	"github.com/tbogdala/fizzle/scene"
)

const (
	renderSystemPriority = 100.0
	renderSystemName     = "RenderSystem"
)

// RenderSystem implements fizzle/scene/System interface and handles the rendering
// of entities in the scene.
type RenderSystem struct {
	Renderer   *forward.ForwardRenderer
	MainWindow *glfw.Window
	Camera     fizzle.Camera

	gfx graphics.GraphicsProvider

	visibleEntities []scene.Entity
}

// NewRenderSystem allocates a new RenderSystem object.
func NewRenderSystem() *RenderSystem {
	rs := new(RenderSystem)
	rs.visibleEntities = []scene.Entity{}
	return rs
}

// Initialize will create the main window using glfw and then create the underyling
// renderer.
func (rs *RenderSystem) Initialize(windowName string, w int, h int) error {
	// create the window and iniitialize opengl
	err := rs.initGraphics(windowName, w, h)
	if err != nil {
		return err
	}

	// setup the forward renderer
	rs.Renderer = forward.NewForwardRenderer(rs.gfx)
	rs.Renderer.ChangeResolution(int32(w), int32(h))

	// set some OpenGL flags
	rs.gfx.Enable(graphics.CULL_FACE)
	rs.gfx.Enable(graphics.DEPTH_TEST)

	return nil
}

// initGraphics creates an OpenGL window and initializes the required graphics libraries.
// It will either succeed or panic.
func (rs *RenderSystem) initGraphics(title string, w int, h int) error {
	// GLFW must be initialized before it's called
	err := glfw.Init()
	if err != nil {
		return fmt.Errorf("Failed to initialize GLFW. %v", err)
	}

	// request a OpenGL 3.3 core context
	glfw.WindowHint(glfw.Samples, 0)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// do the actual window creation
	rs.MainWindow, err = glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		return fmt.Errorf("Failed to create the main window. %v", err)
	}

	// set a function to update the renderer on window resize
	rs.MainWindow.SetSizeCallback(func(w *glfw.Window, width int, height int) {
		rs.Renderer.ChangeResolution(int32(width), int32(height))
	})

	rs.MainWindow.MakeContextCurrent()

	// disable v-sync for max draw rate
	glfw.SwapInterval(0)

	// initialize OpenGL
	rs.gfx, err = opengl.InitOpenGL()
	if err != nil {
		return fmt.Errorf("Failed to initialize OpenGL. %v", err)
	}
	fizzle.SetGraphics(rs.gfx)

	return nil
}

// SetLight puts a light in the specified slot for the renderer.
func (rs *RenderSystem) SetLight(i int, l *forward.Light) {
	rs.Renderer.ActiveLights[i] = l
}

// GetRequestedPriority returns the requested priority level for the System
// which may be of significance to a Manager if they want to order Update() calls.
func (rs *RenderSystem) GetRequestedPriority() float32 {
	return renderSystemPriority
}

// GetName returns the name of the system that can be used to identify
// the System within Manager.
func (rs *RenderSystem) GetName() string {
	return renderSystemName
}

// OnAddEntity should get called by the scene Manager each time a new entity
// has been added to the scene.
func (rs *RenderSystem) OnAddEntity(newEntity scene.Entity) {
	_, okay := newEntity.(RenderableEntity)
	if okay {
		rs.visibleEntities = append(rs.visibleEntities, newEntity)
	}
}

// OnRemoveEntity should get called by the scene Manager each time an entity
// has been removed from the scene.
func (rs *RenderSystem) OnRemoveEntity(oldEntity scene.Entity) {
	surviving := rs.visibleEntities[:]
	for _, e := range rs.visibleEntities {
		if e.GetID() != oldEntity.GetID() {
			surviving = append(surviving, e)
		}
	}
	rs.visibleEntities = surviving
}

// Update renderers the known entities.
func (rs *RenderSystem) Update(frameDelta float32) {
	// clear the screen
	width, height := rs.Renderer.GetResolution()
	rs.gfx.Viewport(0, 0, int32(width), int32(height))
	rs.gfx.ClearColor(0.25, 0.25, 0.25, 1.0)
	rs.gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

	// make the projection and view matrixes
	projection := mgl.Perspective(mgl.DegToRad(60.0), float32(width)/float32(height), 1.0, 100.0)
	var view mgl.Mat4
	if rs.Camera != nil {
		view = rs.Camera.GetViewMatrix()
	} else {
		view = mgl.Ident4()
	}

	// draw stuff the visible entities
	for _, e := range rs.visibleEntities {
		visibleEntity, okay := e.(RenderableEntity)
		if okay {
			r := visibleEntity.GetRenderable()
			rs.Renderer.DrawRenderable(r, nil, projection, view, rs.Camera)
		}
	}

	// draw the screen
	rs.MainWindow.SwapBuffers()
}
