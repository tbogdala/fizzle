// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package forward

import (
	"fmt"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	renderer "github.com/tbogdala/fizzle/renderer"
)

const (
	// MaxForwardLights is the maximum amount of lights supported by this renderer.
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
	Texture graphics.Texture

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

	// owner is the owning renderer
	owner *ForwardRenderer
}

// Destroy deallocates any data being held onto by the ShadowMap that is not
// controlled by the Go GC.
func (shady *ShadowMap) Destroy() {
	// delete the texture associated with the shadow map
	shady.owner.GetGraphics().DeleteTexture(shady.Texture)
}

// Light is a basic light structure used in the forward renderer.
type Light struct {
	// Position is the location of the light in world space
	Position mgl.Vec3

	// Direction is the direction the light points in
	Direction mgl.Vec3

	// DiffuseColor is the color the light emmits
	DiffuseColor mgl.Vec4

	// DiffuseIntensity is how strong the diffuse light should be
	DiffuseIntensity float32

	// SpecularIntensity is how strong the specular highlight should be
	SpecularIntensity float32

	// AmbientIntensity is how strong the ambient light should be
	AmbientIntensity float32

	// ConstAttenuation is the constant coefficient for the attenuation factor
	ConstAttenuation float32

	// LinearAttenuation is the linear coefficient for the attenuation factor
	LinearAttenuation float32

	// QuadraticAttenuation is the quadratic coefficient for the attenuation factor
	QuadraticAttenuation float32

	// Strength is the scale factor on the light strength.
	Strength float32

	// ShadowMap is the texture, and other data, used to render
	// shadows casted by the light. This member is nil when
	// the light does not cast shadows.
	ShadowMap *ShadowMap

	// owner is the owning renderer
	owner *ForwardRenderer
}

// CreateShadowMap allocates a texture and sets up the projections to draw
// the shadows.
func (l *Light) CreateShadowMap(textureSize int32, near float32, far float32, dir mgl.Vec3) {
	// if there was already a shadow map, destroy it
	if l.ShadowMap != nil {
		l.ShadowMap.Destroy()
	}

	// allocate a new structure
	l.ShadowMap = l.owner.NewShadowMap()

	// setup the projection
	l.ShadowMap.Near = near
	l.ShadowMap.Far = far

	// Frustum is okay for directional lights
	// FIXME: this will likely need to be customizable
	factor := float32(0.5)
	l.ShadowMap.Projection = mgl.Frustum(-factor, factor, -factor, factor, near, far)

	l.ShadowMap.TextureSize = textureSize
	l.ShadowMap.Direction = dir

	// create the shadow map texture
	gfx := l.owner.GetGraphics()
	l.ShadowMap.Texture = gfx.GenTexture()
	gfx.ActiveTexture(graphics.TEXTURE0)
	gfx.BindTexture(graphics.TEXTURE_2D, l.ShadowMap.Texture)
	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.DEPTH_COMPONENT32, textureSize, textureSize, 0, graphics.DEPTH_COMPONENT, graphics.UNSIGNED_INT, nil, 0)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR)

	// set the border color and clamp to edge as white so that points outside the shadow map
	// are projected to be not in shadow.
	shadowmapBorder := mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	gfx.TexParameterfv(graphics.TEXTURE_2D, graphics.TEXTURE_BORDER_COLOR, &shadowmapBorder[0])
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.CLAMP_TO_BORDER)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.CLAMP_TO_BORDER)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_COMPARE_MODE, graphics.COMPARE_REF_TO_TEXTURE)

	// a safety unbind
	gfx.BindTexture(graphics.TEXTURE_2D, 0)
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

	// ActiveLights are the current lights that should be used while
	// drawing Renderables.
	ActiveLights [MaxForwardLights]*Light

	width  int32
	height int32

	// lastFrameTime logs the last time the renderer started a frame
	lastFrameTime time.Time

	// shadowFBO is the framebuffer used to render shadows
	shadowFBO graphics.Buffer

	// currentShadowPassLight is the light currently enabled for shadow mapping
	currentShadowPassLight *Light

	// gfx is the underlying graphics implementation for the renderer
	gfx graphics.GraphicsProvider
}

// NewForwardRenderer creates a new forward rendering style render engine object.
func NewForwardRenderer(g graphics.GraphicsProvider) *ForwardRenderer {
	fr := new(ForwardRenderer)
	fr.gfx = g
	fr.OnScreenSizeChanged = func(r *ForwardRenderer, width int32, height int32) {}
	return fr
}

// Destroy releases any data the renderer was holding that it 'owns'.
func (fr *ForwardRenderer) Destroy() {
}

// NewShadowMap creates a new shadow map object
func (fr *ForwardRenderer) NewShadowMap() *ShadowMap {
	shady := new(ShadowMap)
	shady.owner = fr
	shady.Up = mgl.Vec3{0.0, 1.0, 0.0}
	shady.Projection = mgl.Ident4()
	shady.View = mgl.Ident4()
	return shady
}

// NewLight creates a new light object and returns it without
// setting any default attributes.
func (fr *ForwardRenderer) NewLight() *Light {
	l := new(Light)
	l.owner = fr
	return l
}

// NewPointLight creates a new light and sets it up to be a point light.
func (fr *ForwardRenderer) NewPointLight(location mgl.Vec3) *Light {
	light := fr.NewLight()
	light.Position = location
	light.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	light.DiffuseIntensity = 0.70
	light.SpecularIntensity = 0.10
	light.AmbientIntensity = 0.30
	light.ConstAttenuation = 0.20
	light.LinearAttenuation = 0.18
	light.QuadraticAttenuation = 0.15
	light.Strength = 20.0
	return light
}

// NewDirectionalLight creates a new light and sets it up to be a directional light.
func (fr *ForwardRenderer) NewDirectionalLight(dir mgl.Vec3) *Light {
	light := fr.NewLight()
	light.Direction = dir
	light.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	light.DiffuseIntensity = 0.70
	light.SpecularIntensity = 0.10
	light.AmbientIntensity = 0.30
	light.Strength = 20.0
	return light
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

// SetGraphics initializes then renderer with the graphics provider.
func (fr *ForwardRenderer) SetGraphics(gp graphics.GraphicsProvider) {
	fr.gfx = gp
}

// GetGraphics returns the renderer's the graphics provider.
func (fr *ForwardRenderer) GetGraphics() graphics.GraphicsProvider {
	return fr.gfx
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

// EndRenderFrame is the function called at end of the frame.
func (fr *ForwardRenderer) EndRenderFrame() {
	// nothing to do
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
	fr.shadowFBO = fr.gfx.GenFramebuffer()
	fr.gfx.BindFramebuffer(graphics.FRAMEBUFFER, fr.shadowFBO)

	drawBuffers := []uint32{graphics.NONE}
	fr.gfx.DrawBuffers(drawBuffers)
	fr.gfx.ReadBuffer(graphics.NONE)

	/*
		// we attach a shadowmap here just to check the framebuffer completion status
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, light.ShadowMap.Texture, 0);
		if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		}
	*/

	fr.gfx.BindFramebuffer(graphics.FRAMEBUFFER, 0)
}

// StartShadowMapping binds the shadow map framebuffer for use by the lights
// to render shadows.
func (fr *ForwardRenderer) StartShadowMapping() {
	fr.gfx.BindFramebuffer(graphics.FRAMEBUFFER, fr.shadowFBO)
	fr.gfx.Enable(graphics.POLYGON_OFFSET_FILL)
	fr.gfx.PolygonOffset(4.0, 4.0)
	fr.gfx.Enable(graphics.CULL_FACE)
	fr.gfx.CullFace(graphics.FRONT)
	fr.currentShadowPassLight = nil
}

// EndShadowMapping unbinds the shadow map framebuffer and lets the renderer
// proceed as normal.
func (fr *ForwardRenderer) EndShadowMapping() {
	fr.gfx.CullFace(graphics.BACK)
	fr.gfx.Disable(graphics.CULL_FACE)
	fr.gfx.Disable(graphics.POLYGON_OFFSET_FILL)
	fr.gfx.BindFramebuffer(graphics.FRAMEBUFFER, 0)
	fr.currentShadowPassLight = nil
}

// EnableShadowMappingLight enables the light to start casting shadows with draw functions
// and the appropriate shaders.
// NOTE: A good client would call StartShadowMapping() and EndShadowMapping() before
// and after doing shadow draws.
func (fr *ForwardRenderer) EnableShadowMappingLight(l *Light) {
	fr.currentShadowPassLight = l
	l.UpdateShadowMapData()
	fr.gfx.FramebufferTexture2D(graphics.FRAMEBUFFER, graphics.DEPTH_ATTACHMENT, graphics.TEXTURE_2D, l.ShadowMap.Texture, 0)
	fr.gfx.Clear(graphics.DEPTH_BUFFER_BIT)
	fr.gfx.Viewport(0, 0, l.ShadowMap.TextureSize, l.ShadowMap.TextureSize)
}

// do some special binding for the different Renderer types if necessary
func (fr *ForwardRenderer) chainedBinder(renderer renderer.Renderer, r *fizzle.Renderable, shader *fizzle.RenderShader, texturesBound *int32) {
	gfx := fr.gfx
	var lightCount = int32(fr.GetActiveLightCount())
	var shadowLightCount = int32(fr.GetActiveShadowLightCount())
	if lightCount >= 1 {
		for lightI := 0; lightI < int(lightCount); lightI++ {
			light := fr.ActiveLights[lightI]

			shaderLightPosition := shader.GetUniformLocation(fmt.Sprintf("LIGHT_POSITION[%d]", lightI))
			if shaderLightPosition >= 0 {
				gfx.Uniform3f(shaderLightPosition, light.Position[0], light.Position[1], light.Position[2])
			}

			shaderLightDirection := shader.GetUniformLocation(fmt.Sprintf("LIGHT_DIRECTION[%d]", lightI))
			if shaderLightDirection >= 0 {
				gfx.Uniform3f(shaderLightDirection, light.Direction[0], light.Direction[1], light.Direction[2])
			}

			shaderLightDiffuse := shader.GetUniformLocation(fmt.Sprintf("LIGHT_DIFFUSE[%d]", lightI))
			if shaderLightDiffuse >= 0 {
				gfx.Uniform4f(shaderLightDiffuse, light.DiffuseColor[0], light.DiffuseColor[1], light.DiffuseColor[2], light.DiffuseColor[3])
			}

			shaderLightIntensity := shader.GetUniformLocation(fmt.Sprintf("LIGHT_DIFFUSE_INTENSITY[%d]", lightI))
			if shaderLightIntensity >= 0 {
				gfx.Uniform1f(shaderLightIntensity, light.DiffuseIntensity)
			}

			shaderLightSpecularIntensity := shader.GetUniformLocation(fmt.Sprintf("LIGHT_SPECULAR_INTENSITY[%d]", lightI))
			if shaderLightSpecularIntensity >= 0 {
				gfx.Uniform1f(shaderLightSpecularIntensity, light.SpecularIntensity)
			}

			shaderLightAmbientIntensity := shader.GetUniformLocation(fmt.Sprintf("LIGHT_AMBIENT_INTENSITY[%d]", lightI))
			if shaderLightAmbientIntensity >= 0 {
				gfx.Uniform1f(shaderLightAmbientIntensity, light.AmbientIntensity)
			}

			shaderLightConstAttenuation := shader.GetUniformLocation(fmt.Sprintf("LIGHT_CONST_ATTENUATION[%d]", lightI))
			if shaderLightConstAttenuation >= 0 {
				gfx.Uniform1f(shaderLightConstAttenuation, light.ConstAttenuation)
			}

			shaderLightLinearAttenuation := shader.GetUniformLocation(fmt.Sprintf("LIGHT_LINEAR_ATTENUATION[%d]", lightI))
			if shaderLightLinearAttenuation >= 0 {
				gfx.Uniform1f(shaderLightLinearAttenuation, light.LinearAttenuation)
			}

			shaderLightQuadraticAttenuation := shader.GetUniformLocation(fmt.Sprintf("LIGHT_QUADRATIC_ATTENUATION[%d]", lightI))
			if shaderLightQuadraticAttenuation >= 0 {
				gfx.Uniform1f(shaderLightQuadraticAttenuation, light.QuadraticAttenuation)
			}

			shaderLightStrength := shader.GetUniformLocation(fmt.Sprintf("LIGHT_STRENGTH[%d]", lightI))
			if shaderLightStrength >= 0 {
				gfx.Uniform1f(shaderLightStrength, light.Strength)
			}

			shaderShadowMaps := shader.GetUniformLocation(fmt.Sprintf("SHADOW_MAPS[%d]", lightI))
			if shaderShadowMaps >= 0 {
				///* There have been problems in the past on Intel drivers on Mac OS if all of the
				///  samplers are not bound to something. So this code will bind a 0 if the shadow map
				///	 does not exist for that light. */
				gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0 + uint32(*texturesBound)))
				if light.ShadowMap != nil {
					gfx.BindTexture(graphics.TEXTURE_2D, light.ShadowMap.Texture)
				} else {
					gfx.BindTexture(graphics.TEXTURE_2D, 0)
				}
				gfx.Uniform1i(shaderShadowMaps, *texturesBound)
				*texturesBound++
			}

			if light.ShadowMap != nil {
				shaderShadowMatrix := shader.GetUniformLocation(fmt.Sprintf("SHADOW_MATRIX[%d]", lightI))
				if shaderShadowMatrix >= 0 {
					gfx.UniformMatrix4fv(shaderShadowMatrix, 1, false, light.ShadowMap.BiasedMatrix)
				}
			}
		} // lightI

		shaderLightCount := shader.GetUniformLocation("LIGHT_COUNT")
		if shaderLightCount >= 0 {
			gfx.Uniform1i(shaderLightCount, lightCount)
		}

		shaderShadowLightCount := shader.GetUniformLocation("SHADOW_COUNT")
		if shaderShadowLightCount >= 0 {
			gfx.Uniform1i(shaderShadowLightCount, shadowLightCount)
		}

		if fr.currentShadowPassLight != nil {
			shaderShadowVP := shader.GetUniformLocation("SHADOW_VP_MATRIX")
			if shaderShadowVP >= 0 {
				gfx.UniformMatrix4fv(shaderShadowVP, 1, false, fr.currentShadowPassLight.ShadowMap.ViewProjMatrix)
			}
		}

	} // lightcount
}

// DrawRenderable draws a Renderable object with the supplied projection and view matrixes.
func (fr *ForwardRenderer) DrawRenderable(r *fizzle.Renderable, binder renderer.RenderBinder, perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			fr.DrawRenderable(child, binder, perspective, view, camera)
		}
		return
	}

	binders := []renderer.RenderBinder{fr.chainedBinder}
	if binder != nil {
		binders = append(binders, binder)
	}
	renderer.BindAndDraw(fr, r, r.Core.Shader, binders, perspective, view, camera, graphics.TRIANGLES)
}

// DrawRenderableWithShader draws a Renderable object with the supplied projection and view matrixes
// and a different shader than what is set in the Renderable.
func (fr *ForwardRenderer) DrawRenderableWithShader(r *fizzle.Renderable, shader *fizzle.RenderShader,
	binder renderer.RenderBinder, perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			fr.DrawRenderableWithShader(child, shader, binder, perspective, view, camera)
		}
		return
	}

	binders := []renderer.RenderBinder{fr.chainedBinder}
	if binder != nil {
		binders = append(binders, binder)
	}
	renderer.BindAndDraw(fr, r, shader, binders, perspective, view, camera, graphics.TRIANGLES)
}

// DrawLines draws the Renderable using graphics.LINES mode instead of graphics.TRIANGLES.
func (fr *ForwardRenderer) DrawLines(r *fizzle.Renderable, shader *fizzle.RenderShader, binder renderer.RenderBinder,
	perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			fr.DrawLines(child, shader, binder, perspective, view, camera)
		}
		return
	}

	binders := []renderer.RenderBinder{fr.chainedBinder}
	if binder != nil {
		binders = append(binders, binder)
	}
	renderer.BindAndDraw(fr, r, shader, binders, perspective, view, camera, graphics.LINES)
}
