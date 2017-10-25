// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"fmt"
	"log"
	"math"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/golang-ui/nuklear/nk"
	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/fizzle/component"
	"github.com/tbogdala/fizzle/editor/embedded"
	"github.com/tbogdala/fizzle/renderer/forward"
)

// used by nuklear for rendering
const (
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
	fontPtSize       = 12
)

const (
	maxComponentNameLen = 255
)

// an enumeration to identify what the editor is editing: levels or components for example
const (
	// ModeLevel is for the level editor
	ModeLevel = 1

	// ModeComponent is for the component editor
	ModeComponent = 2
)

// ComponentsState contains all state information relevant to the loaded components
type ComponentsState struct {
	// byte buffer for the edit string for searching components
	nameSearchBuffer []byte

	// the current length of the string placed in nameSearchBuffer
	nameSearchLen int32

	// the component manager for all of the components
	manager *component.Manager

	// should be set to the component being edited
	activeComponent *component.Component
}

// State contains all state information relevant to the level.
type State struct {
	// keeps track of all of the loaded components
	components *ComponentsState

	// the texture manager in the editor
	texMan *fizzle.TextureManager

	// the loaded shaders in the editor
	shaders map[string]*fizzle.RenderShader

	// the main window for the editor
	window *glfw.Window

	// the graphics renderer for use by the editor
	render *forward.ForwardRenderer

	// the camera used to render objects
	camera *fizzle.OrbitCamera

	// the nuklear context for rendering ui controls
	ctx *nk.Context

	// the current editing 'mode' for the editor (e.g. ModeLevel or ModeComponent)
	currentMode int

	// the vfov for the rendered perspective
	vfov float32

	// the near distance for the rendered objects
	nearDist float32

	// the far distance for the rendered objects
	farDist float32

	// the current orbit distance for the camera
	orbitDist float32

	// the list of objects that can be rendered for the given editor view
	visibleObjects []*meshRenderable

	// the last X mouse position tracked
	lastMouseX float32

	// the last Y mouse position tracked
	lastMouseY float32
}

// NewState creates a new editor state object to track related content for the level.
func NewState(win *glfw.Window, rend *forward.ForwardRenderer) (*State, error) {
	// setup Nuklear and put a default font in
	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)
	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	fontBytes, err := embedded.Asset("fonts/Hack-Regular.ttf")
	if err != nil {
		return nil, fmt.Errorf("couldn't load the embedded font: %v", err)
	}
	sansFont := nk.NkFontAtlasAddFromBytes(atlas, fontBytes, fontPtSize, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		nk.NkStyleSetFont(ctx, sansFont.Handle())
	}
	nk.NkFontAtlasCleanup(atlas)

	return NewStateWithContext(win, rend, ctx)
}

// NewStateWithContext creates a new editor state with a given nuklear ui context.
func NewStateWithContext(win *glfw.Window, rend *forward.ForwardRenderer, ctx *nk.Context) (*State, error) {
	s := new(State)
	s.render = rend
	s.texMan = fizzle.NewTextureManager()
	s.shaders = make(map[string]*fizzle.RenderShader)

	// load some basic shaders
	basicShader, err := forward.CreateBasicShader()
	if err != nil {
		return nil, fmt.Errorf("Failed to compile and link the basic shader program! " + err.Error())
	}
	basicSkinnedShader, err := forward.CreateBasicSkinnedShader()
	if err != nil {
		return nil, fmt.Errorf("Failed to compile and link the basic skinned shader program! " + err.Error())
	}
	colorShader, err := forward.CreateColorShader()
	if err != nil {
		return nil, fmt.Errorf("Failed to compile and link the color shader program! " + err.Error())
	}
	s.shaders["Basic"] = basicShader
	s.shaders["BasicSkinned"] = basicSkinnedShader
	s.shaders["Color"] = colorShader

	s.components = new(ComponentsState)
	s.components.nameSearchBuffer = make([]byte, 0, 64)
	s.components.manager = component.NewManager(s.texMan, s.shaders)
	s.visibleObjects = make([]*meshRenderable, 0, 16)

	s.window = win
	s.ctx = ctx
	s.vfov = 60
	s.nearDist = 0.1
	s.farDist = 100.0
	s.orbitDist = 5.0

	// start off with an orbit camera
	s.camera = fizzle.NewOrbitCamera(mgl.Vec3{0, 0, 0}, math.Pi/4.0, s.orbitDist, math.Pi/2.0)

	// setup some event handlers
	win.SetScrollCallback(makeMouseScrollCallback(s))
	win.SetCursorPosCallback(makeMousePosCallback(s))

	return s, nil
}

// SetMode sets the current editing 'mode' for the editor  (e.g. ModeLevel or ModeComponent)
func (s *State) SetMode(mode int) {
	s.currentMode = mode

	switch s.currentMode {
	case ModeComponent:
		// reset the lighting for the renderer
		for i := range s.render.ActiveLights {
			s.render.ActiveLights[i] = nil
		}
		light := s.render.NewDirectionalLight(mgl.Vec3{1.0, -0.5, -1.0})
		light.AmbientIntensity = 0.5
		light.DiffuseIntensity = 0.5
		light.SpecularIntensity = 0.3
		s.render.ActiveLights[0] = light
	}
}

// Render draws the editor interface.
func (s *State) Render() {
	width, height := s.window.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))

	// start a new frame
	nk.NkPlatformNewFrame()

	// depending on what mode the editor is in, render a different set of objects
	switch s.currentMode {
	case ModeComponent:
		// if we have a selected component, render it now
		if s.components.activeComponent != nil {
			perspective := mgl.Perspective(mgl.DegToRad(s.vfov), float32(width)/float32(height), s.nearDist, s.farDist)
			view := s.camera.GetViewMatrix()

			// draw the meshes that are visible
			for _, visObj := range s.visibleObjects {
				// push all settings from the component to the renderable
				s.updateVisibleMesh(visObj)

				// draw the thing
				s.render.DrawRenderable(visObj.Renderable, nil, perspective, view, s.camera)
			}

			// draw the child components
			for _, childRef := range s.components.activeComponent.ChildReferences {
				childComp := s.components.manager.GetComponentByFilepath(childRef.File)
				if childComp == nil {
					fmt.Printf("DEBUG: missing child component in the render loop??\n")
				} else {
					r, err := childComp.GetRenderable(s.texMan, s.shaders)
					if err != nil {
						fmt.Printf("Error: couldn't get the renderable for child component %s: %v", childComp.Name, err)
					} else {
						updateChildComponentRenderable(r, childRef)
						s.render.DrawRenderable(r, nil, perspective, view, s.camera)
					}
				}
			}
		}
	}

	// render basic user interface
	s.renderModeToolbar()

	switch s.currentMode {
	case ModeComponent:
		if s.components.activeComponent != nil {
			s.renderComponentEditor()
		} else {
			s.renderComponentBrowser()
		}
	}

	// render out the nuklear ui
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
}

// SetActiveComponent will set the component currently being edited.
func (s *State) SetActiveComponent(c *component.Component) error {
	s.components.activeComponent = c

	// generate the renderables for all of the component meshes
	s.visibleObjects = s.visibleObjects[:0]
	for _, compMesh := range s.components.activeComponent.Meshes {
		compMesh, err := s.makeRenderableForMesh(compMesh)
		if err != nil {
			return fmt.Errorf("Unable to render the component's meshs: %v", err)
		}
		s.visibleObjects = append(s.visibleObjects, compMesh)
	}
	return nil
}

// renderModeToolbar draws the mode toolbar on the screen
func (s *State) renderModeToolbar() {
	bounds := nk.NkRect(10, 10, 200, 40)
	update := nk.NkBegin(s.ctx, "ModeBar", bounds, nk.WindowNoScrollbar)
	if update > 0 {
		nk.NkLayoutRowDynamic(s.ctx, 30, 2)
		{
			if nk.NkButtonLabel(s.ctx, "level") > 0 {
				log.Println("[DEBUG] mode:level pressed!")
				s.SetMode(ModeLevel)
			}
			if nk.NkButtonLabel(s.ctx, "component") > 0 {
				log.Println("[DEBUG] mode:component pressed!")
				s.SetMode(ModeComponent)
			}
		}
	}
	nk.NkEnd(s.ctx)
}

// renderComponentEditor draws the window showing properties
// of the active component.
func (s *State) renderComponentEditor() {
	bounds := nk.NkRect(10, 75, 300, 600)
	update := nk.NkBegin(s.ctx, fmt.Sprintf("Component: %s", s.components.activeComponent.Name), bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowMinimizable|nk.WindowScalable)
	if update > 0 {
		// put in the collapsable mesh list
		_, fileName, fileLine, _ := runtime.Caller(0)
		hashStr := fmt.Sprintf("%s:%d", fileName, fileLine)
		if nk.NkTreePushHashed(s.ctx, nk.TreeTab, "Meshes", nk.Minimized, hashStr, int32(len(hashStr)), int32(fileLine)) != 0 {
			nk.NkLayoutRowDynamic(s.ctx, 120, 1)
			{
				nk.NkGroupBegin(s.ctx, "Mesh List", nk.WindowBorder)
				{
					// put a label in for each mesh the component has
					if len(s.components.activeComponent.Meshes) > 0 {
						for _, compMesh := range s.components.activeComponent.Meshes {
							nk.NkLayoutRowTemplateBegin(s.ctx, 30)
							nk.NkLayoutRowTemplatePushVariable(s.ctx, 80)
							nk.NkLayoutRowTemplatePushStatic(s.ctx, 40)
							nk.NkLayoutRowTemplateEnd(s.ctx)

							nk.NkLabel(s.ctx, compMesh.Name, nk.TextLeft)
							if nk.NkButtonLabel(s.ctx, "E") > 0 {
								log.Println("[DEBUG] comp:mesh:edit pressed!")
							}
						}
					}
				}
			}
			nk.NkGroupEnd(s.ctx)
			nk.NkTreePop(s.ctx)
		}

		// put in the collapsable collisions list
		_, fileName, fileLine, _ = runtime.Caller(0)
		hashStr = fmt.Sprintf("%s:%d", fileName, fileLine)
		if nk.NkTreePushHashed(s.ctx, nk.TreeTab, "Colliders", nk.Minimized, hashStr, int32(len(hashStr)), int32(fileLine)) != 0 {
			nk.NkLayoutRowDynamic(s.ctx, 120, 1)
			{
				nk.NkGroupBegin(s.ctx, "Collider List", nk.WindowBorder)
				{
					// put a label in for each collider the component has
					if len(s.components.activeComponent.Collisions) > 0 {
						for i := range s.components.activeComponent.Collisions {
							nk.NkLayoutRowTemplateBegin(s.ctx, 30)
							nk.NkLayoutRowTemplatePushVariable(s.ctx, 80)
							nk.NkLayoutRowTemplatePushStatic(s.ctx, 40)
							nk.NkLayoutRowTemplateEnd(s.ctx)

							nk.NkLabel(s.ctx, fmt.Sprintf("Collider %d", i), nk.TextLeft)
							if nk.NkButtonLabel(s.ctx, "E") > 0 {
								log.Println("[DEBUG] comp:collider:edit pressed!")
							}
						}
					}
				}
			}
			nk.NkGroupEnd(s.ctx)
			nk.NkTreePop(s.ctx)
		}

		// put in the collapsable child component reference list
		_, fileName, fileLine, _ = runtime.Caller(0)
		hashStr = fmt.Sprintf("%s:%d", fileName, fileLine)
		if nk.NkTreePushHashed(s.ctx, nk.TreeTab, "Child Components", nk.Minimized, hashStr, int32(len(hashStr)), int32(fileLine)) != 0 {
			nk.NkLayoutRowDynamic(s.ctx, 120, 1)
			{
				nk.NkGroupBegin(s.ctx, "Child Components List", nk.WindowBorder)
				{
					// put a label in for each child component the component has
					if len(s.components.activeComponent.ChildReferences) > 0 {
						for _, childRef := range s.components.activeComponent.ChildReferences {
							nk.NkLayoutRowTemplateBegin(s.ctx, 30)
							nk.NkLayoutRowTemplatePushVariable(s.ctx, 80)
							nk.NkLayoutRowTemplatePushStatic(s.ctx, 40)
							nk.NkLayoutRowTemplateEnd(s.ctx)

							nk.NkLabel(s.ctx, childRef.File, nk.TextLeft)
							if nk.NkButtonLabel(s.ctx, "E") > 0 {
								log.Println("[DEBUG] comp:childref:edit pressed!")
							}
						}
					}
				}
			}
			nk.NkGroupEnd(s.ctx)
			nk.NkTreePop(s.ctx)
		}

		// properties
	}
	nk.NkEnd(s.ctx)
}

// renderComponentBrowser draws the window listing all of the known
// components for the level and provides operations related to this.
func (s *State) renderComponentBrowser() {
	bounds := nk.NkRect(10, 75, 300, 600)
	update := nk.NkBegin(s.ctx, "Components", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowMinimizable|nk.WindowScalable)
	if update > 0 {
		// do a layout template so that the buttons are static width
		nk.NkLayoutRowTemplateBegin(s.ctx, 35)
		nk.NkLayoutRowTemplatePushVariable(s.ctx, 80)
		nk.NkLayoutRowTemplatePushStatic(s.ctx, 40)
		nk.NkLayoutRowTemplatePushStatic(s.ctx, 40)
		nk.NkLayoutRowTemplateEnd(s.ctx)
		{
			// component search edit box
			nk.NkEditString(s.ctx, nk.EditField, s.components.nameSearchBuffer,
				&s.components.nameSearchLen, maxComponentNameLen, nk.NkFilterDefault)

			if nk.NkButtonLabel(s.ctx, "F") > 0 {
				log.Println("[DEBUG] comp:find pressed!")
			}

			if nk.NkButtonLabel(s.ctx, "L") > 0 {
				log.Println("[DEBUG] comp:load pressed!")
			}
		}

		nk.NkLayoutRowDynamic(s.ctx, 500, 1)
		{
			nk.NkGroupBegin(s.ctx, "ComponentList", nk.WindowBorder)
			{
				if s.components.manager.GetComponentCount() > 0 {
					// setup some hash information for this root node
					_, fileName, fileLine, _ := runtime.Caller(1)
					hashStr := fmt.Sprintf("%s:%d", fileName, fileLine)
					if nk.NkTreePushHashed(s.ctx, nk.TreeNode, "Component Lists", nk.Maximized, hashStr, int32(len(hashStr)), int32(fileLine)) != 0 {
						// add in labels for all components known to the level
						s.components.manager.MapComponents(func(c *component.Component) {
							nk.NkLayoutRowDynamic(s.ctx, 30, 1)
							nk.NkLabel(s.ctx, c.Name, nk.TextLeft)
							nk.NkTreePop(s.ctx)
						})
					}
				}
			}
		}
		nk.NkGroupEnd(s.ctx)
	}
	nk.NkEnd(s.ctx)
}
