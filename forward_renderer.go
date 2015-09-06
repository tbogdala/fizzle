// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"time"

	gl "github.com/go-gl/gl/v3.3-core/gl"
	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
)

const (
	MaxForwardLights = 4
)

// Light is a basic light structure used in the forward renderer.
type Light struct {
	// Position is the location of the light in world space
	Position mgl.Vec3

	// Direction is the direction the light points in
	Direction mgl.Vec3

	// DiffuseColor is the color the light emmits
	DiffuseColor mgl.Vec4

	// SpecularColor is the color of the specular highlight
	SpecularColor mgl.Vec4

	// Intensity is how strong the diffuse light should be
	DiffuseIntensity float32

	// Intensity is how strong the ambient light should be
	AmbientIntensity float32

	// Attenuation is the coefficient for the attenuation factor
	Attenuation float32
}

// NewLight creates a new light object and returns it
func NewLight() *Light {
	l := new(Light)
	return l
}

// ForwardRenderer is a forward-rendering style renderer, meaning that when
// it draws the geometry it lights it at the same time and the output goes
// to the output framebuffer, which is the only framebuffer.
type ForwardRenderer struct {

	// GeometryPassFn is the function called to render geometry to the
	// framebuffers in the deferred renderer.
	//GeometryPassFn DeferredGeometryPass

	// BeforeDrawFn is the function called by the renderer before
	// endtering the geometry draw function.
	//BeforeDrawFn DeferredBeforeDraw

	// AfterDrawFn is the function called by the renderer after
	// endtering the geometry draw function.
	//AfterDrawFn DeferredAfterDraw

	// OnScreenSizeChangedFn is the function called by the renderer after
	// a screen size change is detected.
	//OnScreenSizeChangedFn ScreenSizeChanged

	// MainWindow the window used to show the rendered graphics.
	MainWindow *glfw.Window

	// UIManager is the user interface manager assigned to the renderer.
	UIManager *UIManager

	// ActiveLights are the current lights that should be used while
	// drawing Renderables.
	ActiveLights [MaxForwardLights]*Light

	shaders       map[string]*RenderShader
	width         int32
	height        int32
	lastFrameTime time.Time
}

// NewForwardRenderer creates a new forward rendering style render engine object.
func NewForwardRenderer(window *glfw.Window) *ForwardRenderer {
	fr := new(ForwardRenderer)
	fr.shaders = make(map[string]*RenderShader)
	fr.MainWindow = window
	return fr
}

// Destroy releases any data the renderer was holding that it 'owns'.
func (fr *ForwardRenderer) Destroy() {
}

// ChangeResolution should be called when the underlying rendering
// window changes size.
func (fr *ForwardRenderer) ChangeResolution(width, height int32) {
	fr.Init(width, height)
}

// GetResolution returns the current dimensions of the renderer.
func (fr *ForwardRenderer) GetResolution() (int32, int32) {
	return fr.width, fr.height
}

// Init initializes the renderer.
func (fr *ForwardRenderer) Init(width, height int32) error {
	fr.width = width
	fr.height = height

	return nil
}

// GetAspectRatio returns the ratio of screen width to height.
func (fr *ForwardRenderer) GetAspectRatio() float32 {
	return float32(fr.width) / float32(fr.height)
}

// GetActiveLightCount counts the number of *Light set in
// the ForwardRenderer's ActiveLights array until a nil is hit.
// NOTE: Obviously requires ActiveLights to be packed sequentially.
func (fr *ForwardRenderer) GetActiveLightCount() int {
	for i := 0; i < MaxForwardLights; i++ {
		if fr.ActiveLights[i] == nil {
			return i
		}
	}
	return MaxForwardLights
}

// DrawRenderable draws a Renderable object with the supplied projection and view matrixes.
func (fr *ForwardRenderer) DrawRenderable(r *Renderable, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			fr.DrawRenderable(child, binder, perspective, view)
		}
		return
	}

	bindAndDraw(fr, r, r.Core.Shader, binder, perspective, view, gl.TRIANGLES)
}

// DrawRenderableWithShader draws a Renderable object with the supplied projection and view matrixes
// and a different shader than what is set in the Renderable.
func (fr *ForwardRenderer) DrawRenderableWithShader(r *Renderable, shader *RenderShader, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			fr.DrawRenderableWithShader(child, shader, binder, perspective, view)
		}
		return
	}

	bindAndDraw(fr, r, shader, binder, perspective, view, gl.TRIANGLES)
}

// DrawLines draws the Renderable using gl.LINES mode instead of gl.TRIANGLES.
func (fr *ForwardRenderer) DrawLines(r *Renderable, shader *RenderShader, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			fr.DrawLines(child, shader, binder, perspective, view)
		}
		return
	}

	bindAndDraw(fr, r, shader, binder, perspective, view, gl.LINES)
}
