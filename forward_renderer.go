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

var (
	shadowBiasMat = mgl.Mat4{
		0.5, 0.0, 0.0, 0.0,
		0.0, 0.5, 0.0, 0.0,
		0.0, 0.0, 0.5, 0.0,
		0.5, 0.5, 0.5, 1.0,
	}
)

// ShadowMap contains the id of the shadow map texture as well as the associated
// vectors and matrixes needed to render the shadow map for the owning light.
// NOTE: only point lights via a given direction are supported at present.
type ShadowMap struct {
	// Texture is the texture for the shadowmap
	Texture uint32

	// TextureSize is the size of the texture in memory.
	TextureSize int32

	// Direction controls the direction the shadowmap points in.
	Direction mgl.Vec3

	// Near is the near distance for the shadowmap projection
	Near float32

	// Far is the far distance for the shadowmap projection
	Far float32

	// Up defines the Up vector for the projection when casting shadows. Defaults to {0,1,0}
	Up mgl.Vec3

	// Projection is the projection transformation matrix for the shadowmap
	Projection mgl.Mat4

	// View is the view transformation matrix for the shadowmap
	// Updated with UpdateShadowMapData().
	View mgl.Mat4

	// ViewProjMatrix is the combination view-projection matrix.
	// Updated with UpdateShadowMapData().
	ViewProjMatrix mgl.Mat4

	// ShadowBiasedMatrix is the shadow biased matrix to account for the difference between NDC and texture space.
	// Updated with UpdateShadowMapData().
	BiasedMatrix mgl.Mat4
}

// NewShadowMap creates a new shadow map object
func NewShadowMap() *ShadowMap {
	shady := new(ShadowMap)
	shady.Up = mgl.Vec3{0.0, 1.0, 0.0}
	shady.Projection = mgl.Ident4()
	shady.View = mgl.Ident4()
	return shady
}

// Destroy deallocates any data being held onto by the ShadowMap that is not
// controlled by the Go GC.
func (shady *ShadowMap) Destroy() {
	// delete the texture associated with the shadow map
	gl.DeleteTextures(1, &shady.Texture)
}

// Light is a basic light structure used in the forward renderer.
type Light struct {
	// Position is the location of the light in world space
	Position mgl.Vec3

	// Direction is the direction the light points in
	Direction mgl.Vec3

	// DiffuseColor is the color the light emmits
	DiffuseColor mgl.Vec4

	// Intensity is how strong the diffuse light should be
	DiffuseIntensity float32

	// Intensity is how strong the ambient light should be
	AmbientIntensity float32

	// Attenuation is the coefficient for the attenuation factor
	Attenuation float32

	// ShadowMap is the texture, and other data, used to render
	// shadows casted by the light. This member is nil when
	// the light does not cast shadows.
	ShadowMap *ShadowMap
}

// NewLight creates a new light object and returns it
func NewLight() *Light {
	l := new(Light)
	return l
}

// CreateShadowMap allocates a texture and sets up the projections to draw
// the shadows.
func (l *Light) CreateShadowMap(textureSize int32, near float32, far float32, dir mgl.Vec3) {
	// if there was already a shadow map, destroy it
	if l.ShadowMap != nil {
		l.ShadowMap.Destroy()
	}

	// allocate a new structure
	l.ShadowMap = NewShadowMap()

	// setup the projection
	l.ShadowMap.Near = near
	l.ShadowMap.Far = far
	l.ShadowMap.Projection = mgl.Frustum(-1.0, 1.0, -1.0, 1.0, near, far)
	l.ShadowMap.TextureSize = textureSize
	l.ShadowMap.Direction = dir

	// create the shadow map texture
	borderColor := mgl.Vec4{1.0, 0.0, 0.0, 1.0}
	gl.GenTextures(1, &l.ShadowMap.Texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, l.ShadowMap.Texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT32, textureSize, textureSize, 0, gl.DEPTH_COMPONENT, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &borderColor[0])
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_MODE, gl.COMPARE_REF_TO_TEXTURE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_FUNC, gl.LEQUAL)

	// a safety unbind
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// UpdateShadowMapData updates a shadow maps internal structures based on data
// from the light.
func (l *Light) UpdateShadowMapData() {
	// don't do nothin' on no shadowmap havin' lights
	if l.ShadowMap == nil {
		return
	}

	// construct a dummy target along the direction vector
	target := l.Position.Add(l.ShadowMap.Direction)

	// update the view matrix
	l.ShadowMap.View = mgl.LookAtV(l.Position, target, l.ShadowMap.Up)

	// update the view projection matrix
	l.ShadowMap.ViewProjMatrix = l.ShadowMap.Projection.Mul4(l.ShadowMap.View)

	// update the shadow biased matrix
	l.ShadowMap.BiasedMatrix = shadowBiasMat.Mul4(l.ShadowMap.ViewProjMatrix)
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

	// OnScreenSizeChanged is the function called by the renderer after
	// a screen size change is detected.
	OnScreenSizeChanged func(fr *ForwardRenderer, width int32, height int32)

	// MainWindow the window used to show the rendered graphics.
	MainWindow *glfw.Window

	// UIManager is the user interface manager assigned to the renderer.
	UIManager *UIManager

	// ActiveLights are the current lights that should be used while
	// drawing Renderables.
	ActiveLights [MaxForwardLights]*Light

	width  int32
	height int32

	// lastFrameTime logs the last time the renderer started a frame
	lastFrameTime time.Time

	// shadowFBO is the framebuffer used to render shadows
	shadowFBO uint32

	// currentShadowPassLight is the light currently enabled for shadow mapping
	currentShadowPassLight *Light
}

// NewForwardRenderer creates a new forward rendering style render engine object.
func NewForwardRenderer(window *glfw.Window) *ForwardRenderer {
	fr := new(ForwardRenderer)
	fr.MainWindow = window
	fr.OnScreenSizeChanged = func(r *ForwardRenderer, width int32, height int32) {}
	return fr
}

// Destroy releases any data the renderer was holding that it 'owns'.
func (fr *ForwardRenderer) Destroy() {
}

// ChangeResolution should be called when the underlying rendering
// window changes size.
func (fr *ForwardRenderer) ChangeResolution(width, height int32) {
	fr.Init(width, height)
	if fr.OnScreenSizeChanged != nil {
		fr.OnScreenSizeChanged(fr, width, height)
	}
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

// EndRenderFrame swaps the buffers and calls GLFW to poll for input.
func (fr *ForwardRenderer) EndRenderFrame() {
	fr.MainWindow.SwapBuffers()
	glfw.PollEvents()
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

// GetActiveShadowLightCount counts the number of *Light set in
// the ForwardRenderer's ActiveLights array that support ShadowMaps until
// a nil is hit or a light doesn't support shadows.
// NOTE: Obviously requires ActiveLights to be packed sequentially
// with lights that support shadow maps in front. Life's not perfect.
func (fr *ForwardRenderer) GetActiveShadowLightCount() int {
	for i := 0; i < MaxForwardLights; i++ {
		if fr.ActiveLights[i] == nil || fr.ActiveLights[i].ShadowMap == nil {
			return i
		}
	}
	return MaxForwardLights
}

// SetupShadowMapRendering is called to create the framebuffer to render the shadows
// and must be called before rendering shadow maps.
func (fr *ForwardRenderer) SetupShadowMapRendering() {
	// create the FBO for the shadows
	gl.GenFramebuffers(1, &fr.shadowFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fr.shadowFBO)

	drawBuffers := [1]uint32{gl.NONE}
	gl.DrawBuffers(1, &drawBuffers[0])
	gl.ReadBuffer(gl.NONE)

	/*
		// we attach a shadowmap here just to check the framebuffer completion status
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, light.ShadowMap.Texture, 0);
		if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		}
	*/

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

// StartShadowMapping binds the shadow map framebuffer for use by the lights
// to render shadows.
func (fr *ForwardRenderer) StartShadowMapping() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, fr.shadowFBO)
	gl.Enable(gl.POLYGON_OFFSET_FILL)
	gl.PolygonOffset(4.0, 4.0)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.FRONT)
	fr.currentShadowPassLight = nil
}

// EndShadowMapping unbinds the shadow map framebuffer and lets the renderer
// proceed as normal.
func (fr *ForwardRenderer) EndShadowMapping() {
	gl.CullFace(gl.BACK)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.POLYGON_OFFSET_FILL)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	fr.currentShadowPassLight = nil
}

// EnableShadowMappingLight enables the light to start casting shadows with draw functions
// and the appropriate shaders.
// NOTE: A good client would call StartShadowMapping() and EndShadowMapping() before
// and after doing shadow draws.
func (fr *ForwardRenderer) EnableShadowMappingLight(l *Light) {
	fr.currentShadowPassLight = l
	l.UpdateShadowMapData()
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, l.ShadowMap.Texture, 0)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	gl.Viewport(0, 0, l.ShadowMap.TextureSize, l.ShadowMap.TextureSize)
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
