// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

/*
	#include  <string.h>
*/
import "C"
import (
	"fmt"
	"log"
	"math"
	"os"
	"runtime"

	"github.com/tbogdala/glider"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/golang-ui/nuklear/nk"
	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/fizzle/component"
	"github.com/tbogdala/fizzle/editor/embedded"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
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

// define readable strings for the mgl.RotationOrder typoe
var (
	rotationOrderStrings = []string{"XYX", "XYZ", "XZX", "XZY", "YXY", "YXZ", "YZY", "YZX", "ZYZ", "ZYX", "ZXZ", "ZXY"}
)

// an enumeration to identify the transform mode for the gizmo
const (
	TransformNone   = 0
	TransformMove   = 1
	TransformRotate = 2
	TransformScale  = 3
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

	// should be set to the component mesh being edited
	activeMesh *component.Mesh

	// the list of meshes that can be rendered for the given editor view
	visibleMeshes []*meshRenderable

	// the component property editor window
	compPropsWindow *ComponentPropsWindow
}

// DisplayConfiguration holds details of display options for the editor.
type DisplayConfiguration struct {
	ShowGridFloor int32
}

// State contains all state information relevant to the level.
type State struct {
	// keeps track of all of the loaded components
	components *ComponentsState

	displayCfg *DisplayConfiguration

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

	// the transform gizmo to render in the editor
	gizmo *Gizmo

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
	vertexColorShader, err := forward.CreateVertexColorShader()
	if err != nil {
		return nil, fmt.Errorf("Failed to compile and link the vertex color shader program! " + err.Error())
	}
	s.shaders["Basic"] = basicShader
	s.shaders["BasicSkinned"] = basicSkinnedShader
	s.shaders["Color"] = colorShader
	s.shaders["VertexColor"] = vertexColorShader

	s.displayCfg = new(DisplayConfiguration)
	s.components = new(ComponentsState)
	s.components.nameSearchBuffer = make([]byte, 0, 64)
	s.components.manager = component.NewManager(s.texMan, s.shaders)
	s.components.visibleMeshes = make([]*meshRenderable, 0, 16)
	s.gizmo = CreateGizmo(vertexColorShader)

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

// SetActiveComponent will set the component currently being edited.
func (s *State) SetActiveComponent(c *component.Component) error {
	s.components.activeComponent = c

	// generate the renderables for all of the component meshes
	s.components.visibleMeshes = s.components.visibleMeshes[:0]
	for _, compMesh := range s.components.activeComponent.Meshes {
		compMesh, err := s.makeRenderableForMesh(compMesh)
		if err != nil {
			return fmt.Errorf("Unable to render the component's meshs: %v", err)
		}
		s.components.visibleMeshes = append(s.components.visibleMeshes, compMesh)
	}

	bounds := nk.NkRect(10, 75, 200, 600)
	s.components.compPropsWindow = NewComponentPropsWindow(fmt.Sprintf("%s Properties", c.Name), c, &bounds)

	return nil
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

func unprojectMouse(x, y, z float32, w, h float32, projection mgl.Mat4, view mgl.Mat4) mgl.Vec3 {
	// Note: thanks be to http://antongerdelan.net/opengl/raycasting.html

	// normalized device coordinates
	dcX := float32((2.0*x)/w - 1.0)
	dcY := float32(1.0 - (2.0*y)/h)
	dcZ := float32(1.0)
	rayNds := mgl.Vec3{dcX, dcY, dcZ}

	// 4d homogeneous clip coordinates
	rayClip := mgl.Vec4{rayNds[0], rayNds[1], -1.0, 1.0}

	// 4d camera coordinates
	rayEye := projection.Inv().Mul4x1(rayClip)
	rayEye[2] = -1.0
	rayEye[3] = 0.0

	// 4d world coordinates
	rayWorld4 := view.Inv().Mul4x1(rayEye)
	rayWorld := mgl.Vec3{rayWorld4[0], rayWorld4[1], rayWorld4[2]}
	rayWorld = rayWorld.Normalize()

	return rayWorld
}

// updateGizmoScale resizes the gizmo if necessary
func (s *State) updateGizmoScale() {
	camScale := s.camera.GetDistance() * 0.20 // a little arbitrary, but seems to work well
	s.gizmo.UpdateScale(camScale)
}

// Update should be called to do interface checks that do not come through via callbacks.
func (s *State) Update() {
	w, h := s.window.GetSize()
	mx, my := s.window.GetCursorPos()

	// resize the gizmo if necessary
	s.updateGizmoScale()

	// if we have an active component, do some extra input checks
	active := s.components.activeComponent
	if active != nil {
		// do LMB press queries and update the gizmo accordingly
		lmbStatus := s.window.GetMouseButton(glfw.MouseButton1)
		if lmbStatus == glfw.Press {
			perspective := mgl.Perspective(mgl.DegToRad(s.vfov), float32(w)/float32(h), s.nearDist, s.farDist)
			view := s.camera.GetViewMatrix()

			x, y := s.window.GetCursorPos()
			z := s.camera.GetPosition()[2]
			clickLoc := unprojectMouse(float32(x), float32(y), z, float32(w), float32(h), perspective, view)

			var ray glider.CollisionRay
			ray.Origin = s.camera.GetPosition()
			ray.SetDirection(clickLoc)

			s.gizmo.OnLMBDown(float32(mx/float64(w)), float32(my/float64(h)), &ray, s.components.activeComponent)
		} else {
			s.gizmo.OnLMBUp()
		}

		// set the camera to lock onto the selected component if there is one
		s.camera.SetTarget(active.Offset)
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
			for _, visMesh := range s.components.visibleMeshes {
				// push all settings from the component to the renderable
				s.updateVisibleMesh(visMesh)

				// draw the thing
				s.render.DrawRenderable(visMesh.Renderable, nil, perspective, view, s.camera)
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

			if s.gizmo.GetTransformMode() != TransformNone {
				gizRenderable := s.gizmo.Gizmo.GetRenderable()
				if gizRenderable != nil {
					gfx := fizzle.GetGraphics()
					gfx.Disable(graphics.DEPTH_TEST)
					s.render.DrawRenderable(gizRenderable, nil, perspective, view, s.camera)
					gfx.Enable(graphics.DEPTH_TEST)
				}
			}
		}
	}

	// render basic user interface
	s.renderModeToolbar()

	switch s.currentMode {
	case ModeComponent:
		if s.components.activeComponent != nil {
			if s.components.compPropsWindow != nil {
				s.components.compPropsWindow.Render(s.ctx, s.window)
			}
		} else {
			s.renderComponentBrowser()
		}
	}

	// render out the nuklear ui
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
}

var (
	// used to track drawing a popup message box for information in the menu bar
	toolbarMsgPopupAlive = false
	toolbarMsgPopupMsg   = ""

	// used to track drawing a file save-as box
	saveAsPopup *FileDialog
)

// renderModeToolbar draws the mode toolbar on the screen
func (s *State) renderModeToolbar() {
	width, height := s.window.GetSize()
	bounds := nk.NkRect(0, 0, float32(width), 35)
	update := nk.NkBegin(s.ctx, "ModeBar", bounds, nk.WindowNoScrollbar)
	if update > 0 {
		// draw a menu
		nk.NkMenubarBegin(s.ctx)
		nk.NkLayoutRowBegin(s.ctx, nk.Static, 25, 9)
		nk.NkLayoutRowPush(s.ctx, 55)
		menuSize := nk.NkVec2(200, 200)
		if nk.NkMenuBeginLabel(s.ctx, "Project", nk.TextLeft, menuSize) > 0 {
			nk.NkLayoutRowDynamic(s.ctx, 25, 1)
			if nk.NkMenuItemLabel(s.ctx, "Level Editor", nk.TextLeft) > 0 {
				fmt.Println("DEBUG Project:Level Editor pressed")
				s.SetMode(ModeLevel)
			}
			nk.NkLayoutRowDynamic(s.ctx, 25, 1)
			if nk.NkMenuItemLabel(s.ctx, "Component Editor", nk.TextLeft) > 0 {
				fmt.Println("DEBUG Project:Component Editor pressed")
				s.SetMode(ModeComponent)
			}
			nk.NkMenuEnd(s.ctx)
		}
		nk.NkLayoutRowPush(s.ctx, 55)
		if nk.NkMenuBeginLabel(s.ctx, "Display", nk.TextLeft, menuSize) > 0 {
			nk.NkLayoutRowDynamic(s.ctx, 25, 1)
			nk.NkCheckboxLabel(s.ctx, "Grid Floor", &s.displayCfg.ShowGridFloor)
			nk.NkMenuEnd(s.ctx)
		}

		// context dependant menus
		if s.currentMode == ModeComponent {
			nk.NkLayoutRowPush(s.ctx, 75)
			if nk.NkMenuBeginLabel(s.ctx, "Component", nk.TextLeft, nk.NkVec2(250, 200)) > 0 {
				if s.components.activeComponent != nil {
					nk.NkLayoutRowDynamic(s.ctx, 25, 1)
					if nk.NkMenuItemLabel(s.ctx, "Save component file", nk.TextLeft) > 0 {
						destFile := s.components.activeComponent.GetFilepath()
						err := s.SaveComponentFile(s.components.activeComponent, destFile)
						if err != nil {
							toolbarMsgPopupMsg = fmt.Sprintf("Error: %v", err)
						} else {
							toolbarMsgPopupMsg = fmt.Sprintf("File saved to: %s", destFile)
						}
						toolbarMsgPopupAlive = true
					}

					nk.NkLayoutRowDynamic(s.ctx, 25, 1)
					if nk.NkMenuItemLabel(s.ctx, "Save component file as ...", nk.TextLeft) > 0 {
						dir := s.components.activeComponent.GetDirPath()
						filename := s.components.activeComponent.GetFilename()
						bounds := nk.NkRect(float32(width)/2.0-200.0, float32(height)/2.0-225.0, 400, 450)
						saveAsPopup = NewFileDialog("Save component file as ...", dir, filename, &bounds)
						err := saveAsPopup.UpdateFileList()
						if err != nil {
							saveAsPopup = nil
							toolbarMsgPopupAlive = true
							toolbarMsgPopupMsg = fmt.Sprintf("Error: %v", err)
						}
					}
				}

				// do we have a selected mesh?
				if s.components.activeMesh != nil {
					nk.NkLayoutRowDynamic(s.ctx, 25, 1)
					if nk.NkMenuItemLabel(s.ctx, "Reload source mesh", nk.TextLeft) > 0 {
						err := s.reloadSourceComponentMesh(s.components.activeMesh)
						if err != nil {
							toolbarMsgPopupMsg = err.Error()
							toolbarMsgPopupAlive = true
						}
					}

					nk.NkLayoutRowDynamic(s.ctx, 25, 1)
					if nk.NkMenuItemLabel(s.ctx, "Compile selected mesh", nk.TextLeft) > 0 {
						destFile, err := saveComponentMesh(s.components.activeMesh)
						if err != nil {
							toolbarMsgPopupMsg = fmt.Sprintf("Error: %v", err)
						} else {
							toolbarMsgPopupMsg = fmt.Sprintf("File saved to: %s", destFile)
						}
						toolbarMsgPopupAlive = true
					}
				}
				nk.NkMenuEnd(s.ctx)
			}
		}

		// widgets ...

		nk.NkLayoutRowPush(s.ctx, 20)
		nk.NkLabel(s.ctx, "|", nk.TextCentered)

		nk.NkLayoutRowPush(s.ctx, 60)
		if nk.NkButtonLabel(s.ctx, "Move") > 0 {
			log.Println("[DEBUG] gizmo:Move pressed!")
			m := s.gizmo.GetTransformMode()
			if m == TransformMove {
				s.gizmo.SetTransformMode(TransformNone)
			} else {
				s.gizmo.SetTransformMode(TransformMove)
			}
			s.updateGizmoScale()
		}
		nk.NkLayoutRowPush(s.ctx, 60)
		if nk.NkButtonLabel(s.ctx, "Rotate") > 0 {
			log.Println("[DEBUG] gizmo:Rotate pressed!")
			m := s.gizmo.GetTransformMode()
			if m == TransformRotate {
				s.gizmo.SetTransformMode(TransformNone)
			} else {
				s.gizmo.SetTransformMode(TransformRotate)
			}
			s.updateGizmoScale()
		}
		nk.NkLayoutRowPush(s.ctx, 60)
		if nk.NkButtonLabel(s.ctx, "Scale") > 0 {
			log.Println("[DEBUG] gizmo:Scale pressed!")
			m := s.gizmo.GetTransformMode()
			if m == TransformScale {
				s.gizmo.SetTransformMode(TransformNone)
			} else {
				s.gizmo.SetTransformMode(TransformScale)
			}
			s.updateGizmoScale()
		}

		nk.NkLayoutRowPush(s.ctx, 20)
		nk.NkLabel(s.ctx, "|", nk.TextCentered)

		nk.NkLayoutRowPush(s.ctx, 120)
		if s.components.activeMesh != nil {
			nk.NkLabel(s.ctx, fmt.Sprintf("Selected: %s", s.components.activeMesh.Name), nk.TextLeft)
		} else {
			nk.NkLabel(s.ctx, "No selection", nk.TextLeft)
		}

		nk.NkMenubarEnd(s.ctx)

		// show any messages that occurred from operations triggered by the menu
		if toolbarMsgPopupAlive {
			// we're missing the font.width() function call to do a legit calculation of this,
			// so we're going to fudge the numbers.
			textWidth := float32(fontPtSize) * float32(len(toolbarMsgPopupMsg)) * 0.6

			msgBoxW := float32(textWidth)
			msgBoxH := float32(90)
			msgBoxX := (float32(width) - msgBoxW) / 2.0
			msgBoxY := float32(height)/2.0 - msgBoxH
			msgBoxRect := nk.NkRect(msgBoxX, msgBoxY, msgBoxW, msgBoxH)
			if nk.NkPopupBegin(s.ctx, nk.PopupStatic, "Save Component Mesh", 0, msgBoxRect) > 0 {
				nk.NkLayoutRowDynamic(s.ctx, 25, 1)
				nk.NkLabel(s.ctx, toolbarMsgPopupMsg, nk.TextAlignLeft)

				nk.NkLayoutRowDynamic(s.ctx, 25, 1)
				if nk.NkButtonLabel(s.ctx, "Close") > 0 {
					toolbarMsgPopupAlive = false
					nk.NkPopupClose(s.ctx)
				}
				nk.NkPopupEnd(s.ctx)
			} else {
				toolbarMsgPopupAlive = false
			}
		}
	}
	nk.NkEnd(s.ctx)

	// if the save-as dialog is active, render that
	if saveAsPopup != nil {
		buttonPress := saveAsPopup.Render(s.ctx)
		if buttonPress != FileDialogNoPress {
			if buttonPress == FileDialogOkay {
				destFile := fmt.Sprintf("%s%c%s", saveAsPopup.Directory, os.PathSeparator, saveAsPopup.Filename)
				err := s.SaveComponentFile(s.components.activeComponent, destFile)
				if err != nil {
					toolbarMsgPopupMsg = fmt.Sprintf("Error: %v", err)
				} else {
					toolbarMsgPopupMsg = fmt.Sprintf("File saved to: %s", destFile)
				}
				toolbarMsgPopupAlive = true
			}
			saveAsPopup = nil
		}
	}

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
