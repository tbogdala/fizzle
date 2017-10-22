// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"

	"github.com/tbogdala/fizzle/renderer"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
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

// ComponentState contains all of the state specific for a component
type ComponentState struct {
	Selected int32
}

// ComponentsState contains all state information relevant to the loaded components
type ComponentsState struct {
	// byte buffer for the edit string for searching components
	nameSearchBuffer []byte

	// the current length of the string placed in nameSearchBuffer
	nameSearchLen int32

	// the component manager for all of the components
	manager *component.Manager

	// component selection map of component.Name -> int
	componentStates map[string]*ComponentState
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
	render renderer.Renderer

	// the nuklear context for rendering ui controls
	ctx *nk.Context

	// the current editing 'mode' for the editor (e.g. ModeLevel or ModeComponent)
	currentMode int
}

// NewState creates a new editor state object to track related content for the level.
func NewState(win *glfw.Window, rend renderer.Renderer) (*State, error) {
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

	return NewStateWithContext(win, rend, ctx)
}

// NewStateWithContext creates a new editor state with a given nuklear ui context.
func NewStateWithContext(win *glfw.Window, rend renderer.Renderer, ctx *nk.Context) (*State, error) {
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
	s.components.componentStates = make(map[string]*ComponentState)
	s.components.manager = component.NewManager(s.texMan, s.shaders)

	s.window = win
	s.ctx = ctx
	return s, nil
}

// SetMode sets the current editing 'mode' for the editor  (e.g. ModeLevel or ModeComponent)
func (s *State) SetMode(mode int) {
	s.currentMode = mode
}

// Render draws the editor interface.
func (s *State) Render() {
	// start a new frame
	nk.NkPlatformNewFrame()

	// render basic user interface
	s.renderModeToolbar()

	switch s.currentMode {
	case ModeComponent:
		s.renderComponentBrowser()
	}

	// render out the nuklear ui
	width, height := s.window.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
}

// LoadComponentFile attempts to load the component JSON file into the editor.
func (s *State) LoadComponentFile(filepath string) error {
	var theComponent component.Component
	existingCompJSON, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read %s as a component JSON file", filepath)
	}

	err = json.Unmarshal(existingCompJSON, &theComponent)
	if err != nil {
		return fmt.Errorf("failed to load component %s: %v", filepath, err)
	}

	s.components.manager.AddComponent(theComponent.Name, &theComponent)
	s.components.componentStates[theComponent.Name] = new(ComponentState)

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
			}
			if nk.NkButtonLabel(s.ctx, "component") > 0 {
				log.Println("[DEBUG] mode:component pressed!")
			}
		}
	}
	nk.NkEnd(s.ctx)
}

// renderComponentBrowser draws the window listing all of the known
// components for the level and provides operations related to this.
func (s *State) renderComponentBrowser() {
	bounds := nk.NkRect(10, 75, 300, 600)
	update := nk.NkBegin(s.ctx, "Components", bounds, nk.WindowBorder|nk.WindowMovable|nk.WindowMinimizable)
	if update > 0 {
		nk.NkLayoutRow(s.ctx, nk.Dynamic, 35, 3, []float32{0.7, 0.15, 0.15})
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
				// setup some hash information for this root node
				_, fileName, fileLine, _ := runtime.Caller(1)
				hashStr := fmt.Sprintf("%s:%d", fileName, fileLine)
				if nk.NkTreePushHashed(s.ctx, nk.TreeNode, "Component Lists", nk.Maximized, hashStr, int32(len(hashStr)), int32(fileLine)) != 0 {
					// add in labels for all components known to the level
					s.components.manager.MapComponents(func(c *component.Component) {
						state := s.components.componentStates[c.Name]
						nk.NkLayoutRowDynamic(s.ctx, 30, 1)
						nk.NkSelectableLabel(s.ctx, c.Name, nk.TextLeft, &state.Selected)
						nk.NkTreePop(s.ctx)
					})
				}
			}
		}
		nk.NkGroupEnd(s.ctx)
	}
	nk.NkEnd(s.ctx)
}
