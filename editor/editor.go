// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/tbogdala/fizzle/editor/embedded"
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
}

// State contains all state information relevant to the level.
type State struct {
	components *ComponentsState

	// the main window for the editor
	win *glfw.Window

	// the nuklear context for rendering ui controls
	ctx *nk.Context

	// the current editing 'mode' for the editor (e.g. ModeLevel or ModeComponent)
	currentMode int
}

// NewState creates a new editor state object to track related content for the level.
func NewState(win *glfw.Window) (*State, error) {
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

	return NewStateWithContext(win, ctx)
}

// NewStateWithContext creates a new editor state with a given nuklear ui context.
func NewStateWithContext(win *glfw.Window, ctx *nk.Context) (*State, error) {
	s := new(State)
	s.components = new(ComponentsState)
	s.components.nameSearchBuffer = make([]byte, 0, 64)
	s.win = win
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
	width, height := s.win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
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
		nk.NkLayoutRow(s.ctx, nk.Dynamic, 35, 2, []float32{0.8, 0.2})
		{
			// component search edit box
			nk.NkEditString(s.ctx, nk.EditField, s.components.nameSearchBuffer,
				&s.components.nameSearchLen, maxComponentNameLen, nk.NkFilterDefault)

			if nk.NkButtonLabel(s.ctx, "F") > 0 {
				log.Println("[DEBUG] comp:find pressed!")
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
					nk.NkLayoutRowDynamic(s.ctx, 30, 1)
					var selected int32
					nk.NkSelectableLabel(s.ctx, "Test Component", nk.TextLeft, &selected)
					nk.NkTreePop(s.ctx)
				}
			}
		}
		nk.NkGroupEnd(s.ctx)
	}
	nk.NkEnd(s.ctx)
}
