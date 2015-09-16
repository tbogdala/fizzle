// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"fmt"
	"time"

	gl "github.com/go-gl/gl/v3.3-core/gl"
	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/groggy"
)

// ScreenSizeChanged is the type of the function called by the renderer after
// a screen size change is detected.
type ScreenSizeChanged func(dr *DeferredRenderer, width int32, height int32)

// DeferredBeforeDraw is the type of the function called by the renderer before
// endtering the geometry draw function.
type DeferredBeforeDraw func(dr *DeferredRenderer, deltaFrameTime float32)

// DeferredAfterDraw is the type of the function called by the renderer after
// endtering the geometry draw function.
type DeferredAfterDraw func(dr *DeferredRenderer, deltaFrameTime float32)

// DeferredGeometryPass is the type of the function called to render geometry to the
// framebuffers in the deferred renderer.
type DeferredGeometryPass func(dr *DeferredRenderer, deltaFrameTime float32)

// DeferredCompositePass is the type of the function called to render the framebuffers
// to the screen in the deferred renderer.
type DeferredCompositePass func(dr *DeferredRenderer, deltaFrameTime float32)

// DeferredRenderer is a deferred-rendering style renderer. Which means that
// it creates several framebuffers for shaders to write to and has two main
// rendering steps: 1) geometry and 2) compositing.
type DeferredRenderer struct {
	Frame          uint32
	Depth          uint32
	Diffuse        uint32
	Positions      uint32
	Normals        uint32
	CompositePlane *Renderable

	// GeometryPassFn is the function called to render geometry to the
	// framebuffers in the deferred renderer.
	GeometryPassFn DeferredGeometryPass

	// CompositePassFn is the function called to render the framebuffers
	// to the screen in the deferred renderer.
	CompositePassFn DeferredCompositePass

	// BeforeDrawFn is the function called by the renderer before
	// endtering the geometry draw function.
	BeforeDrawFn DeferredBeforeDraw

	// AfterDrawFn is the function called by the renderer after
	// endtering the geometry draw function.
	AfterDrawFn DeferredAfterDraw

	// OnScreenSizeChangedFn is the function called by the renderer after
	// a screen size change is detected.
	OnScreenSizeChangedFn ScreenSizeChanged

	// MainWindow the window used to show the rendered composite plane to.
	MainWindow *glfw.Window

	// UIManager is the user interface manager assigned to the renderer.
	UIManager *UIManager

	shaders       map[string]*RenderShader
	width         int32
	height        int32
	lastFrameTime time.Time
}

// NewDeferredRenderer creates a new DeferredRenderer and sets some of the
// default callback functions as well as other default values.
func NewDeferredRenderer(window *glfw.Window) *DeferredRenderer {
	dr := new(DeferredRenderer)
	dr.shaders = make(map[string]*RenderShader)
	dr.MainWindow = window
	dr.OnScreenSizeChangedFn = func(r *DeferredRenderer, width int32, height int32) {}
	dr.BeforeDrawFn = func(r *DeferredRenderer, deltaFrameTime float32) {}
	dr.AfterDrawFn = func(r *DeferredRenderer, deltaFrameTime float32) {}
	dr.GeometryPassFn = func(dr *DeferredRenderer, deltaFrameTime float32) {}
	dr.CompositePassFn = func(dr *DeferredRenderer, deltaFrameTime float32) {}

	return dr
}

// Destroy releases all of the OpenGL buffers the DeferredRenderer is holding on to.
func (dr *DeferredRenderer) Destroy() {
	gl.DeleteRenderbuffers(1, &dr.Depth)
	gl.DeleteRenderbuffers(1, &dr.Diffuse)
	gl.DeleteRenderbuffers(1, &dr.Positions)
	gl.DeleteRenderbuffers(1, &dr.Normals)
	gl.DeleteFramebuffers(1, &dr.Frame)
	dr.CompositePlane.Core.DestroyCore()
}

// ChangeResolution internally changes the size of the framebuffers and compositing
// plane that are used for rendering.
func (dr *DeferredRenderer) ChangeResolution(width, height int32) {
	dr.Destroy()
	dr.Init(width, height)
}

// GetResolution returns the current dimensions of the renderer.
func (dr *DeferredRenderer) GetResolution() (int32, int32) {
	return dr.width, dr.height
}

// GetAspectRatio returns the ratio of screen width to height.
func (dr *DeferredRenderer) GetAspectRatio() float32 {
	return float32(dr.width) / float32(dr.height)
}

// EndRenderFrame swaps the buffers and calls GLFW to poll for input.
func (dr *DeferredRenderer) EndRenderFrame() {
	dr.MainWindow.SwapBuffers()
	glfw.PollEvents()
}

// Init sets up the DeferredRenderer by creating all of the framebuffers and
// creating the compositing plane.
func (dr *DeferredRenderer) Init(width, height int32) error {
	dr.width = width
	dr.height = height
	gl.GenFramebuffers(1, &dr.Frame)

	// setup the depth buffer
	gl.GenRenderbuffers(1, &dr.Depth)
	gl.BindRenderbuffer(gl.RENDERBUFFER, dr.Depth)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, width, height)

	// setup the diffuse texture
	gl.GenTextures(1, &dr.Diffuse)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, dr.Diffuse)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, width, height, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// setup the positions texture
	gl.GenTextures(1, &dr.Positions)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, dr.Positions)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, width, height, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// setup the normals texture
	gl.GenTextures(1, &dr.Normals)
	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, dr.Normals)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, width, height, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// now bind all of these things to the framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, dr.Frame)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, dr.Depth)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, dr.Diffuse, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, dr.Positions, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.TEXTURE_2D, dr.Normals, 0)

	// how did it all go? lets find out ...
	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	switch {
	case status == gl.FRAMEBUFFER_UNSUPPORTED:
		return fmt.Errorf("Failed to create the deferred rendering pipeline because the framebuffer was unsupported.\n")
	case status != gl.FRAMEBUFFER_COMPLETE:
		return fmt.Errorf("Failed to create the deferred rendering pipeline. Code 0x%x\n", status)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// create a plane for the composite pass
	groggy.Logsf("DEBUG", "Creatiing composite plane %dx%d.", width, height)
	cp := CreatePlaneXY("composite", 0, 0, float32(width), float32(height))
	var cptex uint32
	gl.GenTextures(1, &cptex)
	cp.Core.Tex0 = cptex
	gl.BindTexture(gl.TEXTURE_2D, cp.Core.Tex0)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	dr.CompositePlane = cp

	return nil
}

// InitShaders sets up the special shaders used in a deferred rendering pipeline.
func (dr *DeferredRenderer) InitShaders(compositeBaseFilepath string, dirlightShaderFilepath string) error {
	// Load the composite pass shader and assert variables exist
	prog, err := LoadShaderProgramFromFiles(compositeBaseFilepath, func(p uint32) {
		gl.BindFragDataLocation(p, 0, gl.Str("frag_color\x00"))
	})
	if err != nil {
		return fmt.Errorf("Failed to compile and link the deferred render composite program! %v", err)
	}
	dr.shaders["composite"] = prog

	// Load the directional light shader and assert variables exist
	prog, err = LoadShaderProgramFromFiles(dirlightShaderFilepath, func(p uint32) {
		gl.BindFragDataLocation(p, 0, gl.Str("frag_color\x00"))
	})
	if err != nil {
		return fmt.Errorf("Failed to compile and link the deferred render composite program! %v", err)
	}
	dr.shaders["directional_light"] = prog

	return nil
}

// CompositeDraw draws the final composite image onto the composite plane using
// the composite shader.
func (dr *DeferredRenderer) CompositeDraw() {
	// the view matrix would be identity
	ortho := mgl.Ortho(0, float32(dr.width), 0, float32(dr.height), -200.0, 200.0)

	r := dr.CompositePlane
	shader := dr.shaders["composite"]
	gl.UseProgram(shader.Prog)
	gl.BindVertexArray(r.Core.Vao)

	model := r.GetTransformMat4()

	shaderMvp := shader.GetUniformLocation("MVP_MATRIX")
	if shaderMvp >= 0 {
		mvp := ortho.Mul4(model)
		gl.UniformMatrix4fv(shaderMvp, 1, false, &mvp[0])
	}

	shaderPosition := shader.GetAttribLocation("VERTEX_POSITION")
	if shaderPosition >= 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.VertVBO)
		gl.EnableVertexAttribArray(uint32(shaderPosition))
		gl.VertexAttribPointer(uint32(shaderPosition), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}

	shaderVertUv := shader.GetAttribLocation("VERTEX_UV_0")
	if shaderVertUv >= 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.UvVBO)
		gl.EnableVertexAttribArray(uint32(shaderVertUv))
		gl.VertexAttribPointer(uint32(shaderVertUv), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}

	shaderTex0 := shader.GetUniformLocation("DIFFUSE_TEX")
	if shaderTex0 >= 0 {
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, dr.Diffuse)
		gl.Uniform1i(shaderTex0, 0)
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.DrawElements(gl.TRIANGLES, int32(r.FaceCount*3), gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)
}

// DrawDirectionalLight draws the composite plane while lighting everything with
// a directional light using the parameters specified.
func (dr *DeferredRenderer) DrawDirectionalLight(eye mgl.Vec3, dir mgl.Vec3, color mgl.Vec3, ambient float32, diffuse float32, specular float32) {
	// the view matrix would be identity
	ortho := mgl.Ortho(0, float32(dr.width), 0, float32(dr.height), -200.0, 200.0)

	r := dr.CompositePlane
	shader := dr.shaders["directional_light"]
	gl.UseProgram(shader.Prog)
	gl.BindVertexArray(r.Core.Vao)

	model := r.GetTransformMat4()

	shaderMvp := shader.GetUniformLocation("MVP_MATRIX")
	if shaderMvp >= 0 {
		mvp := ortho.Mul4(model)
		gl.UniformMatrix4fv(shaderMvp, 1, false, &mvp[0])
	}

	shaderPosition := shader.GetAttribLocation("VERTEX_POSITION")
	if shaderPosition >= 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.VertVBO)
		gl.EnableVertexAttribArray(uint32(shaderPosition))
		gl.VertexAttribPointer(uint32(shaderPosition), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}

	shaderVertUv := shader.GetAttribLocation("VERTEX_UV_0")
	if shaderVertUv >= 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.UvVBO)
		gl.EnableVertexAttribArray(uint32(shaderVertUv))
		gl.VertexAttribPointer(uint32(shaderVertUv), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}

	shaderEyePosition := shader.GetAttribLocation("EYE_WORLD_POSITION")
	if shaderEyePosition >= 0 {
		gl.Uniform3f(shaderEyePosition, eye[0], eye[1], eye[2])
	}

	shaderTex0 := shader.GetUniformLocation("DIFFUSE_TEX")
	if shaderTex0 >= 0 {
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, dr.Diffuse)
		gl.Uniform1i(shaderTex0, 0)
	}

	shaderTex1 := shader.GetUniformLocation("POSITIONS_TEX")
	if shaderTex1 >= 0 {
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, dr.Positions)
		gl.Uniform1i(shaderTex1, 1)
	}

	shaderTex2 := shader.GetUniformLocation("NORMALS_TEX")
	if shaderTex2 >= 0 {
		gl.ActiveTexture(gl.TEXTURE2)
		gl.BindTexture(gl.TEXTURE_2D, dr.Normals)
		gl.Uniform1i(shaderTex2, 2)
	}

	shaderLightDir := shader.GetUniformLocation("LIGHT_DIRECTION")
	if shaderLightDir >= 0 {
		gl.Uniform3f(shaderLightDir, dir[0], dir[1], dir[2])
	}
	shaderLightColor := shader.GetUniformLocation("LIGHT_COLOR")
	if shaderLightColor >= 0 {
		gl.Uniform3f(shaderLightColor, color[0], color[1], color[2])
	}
	shaderLightAmbient := shader.GetUniformLocation("LIGHT_AMBIENT_INTENSITY")
	if shaderLightAmbient >= 0 {
		gl.Uniform1f(shaderLightAmbient, ambient)
	}
	shaderLightDiffuse := shader.GetUniformLocation("LIGHT_DIFFUSE_INTENSITY")
	if shaderLightDiffuse >= 0 {
		gl.Uniform1f(shaderLightDiffuse, diffuse)
	}
	shaderLightSpecPow := shader.GetUniformLocation("LIGHT_SPECULAR_POWER")
	if shaderLightSpecPow >= 0 {
		gl.Uniform1f(shaderLightSpecPow, specular)
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.DrawElements(gl.TRIANGLES, int32(r.FaceCount*3), gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)
}

// DrawRenderable draws a Renderable object with the supplied projection and view matrixes.
func (dr *DeferredRenderer) DrawRenderable(r *Renderable, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			dr.DrawRenderable(child, binder, perspective, view)
		}
		return
	}

	bindAndDraw(dr, r, r.Core.Shader, binder, perspective, view, gl.TRIANGLES)
}

// DrawRenderableWithShader draws a Renderable object with the supplied projection and view matrixes
// and a different shader than what is set in the Renderable.
func (dr *DeferredRenderer) DrawRenderableWithShader(r *Renderable, shader *RenderShader, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			dr.DrawRenderableWithShader(child, shader, binder, perspective, view)
		}
		return
	}

	bindAndDraw(dr, r, shader, binder, perspective, view, gl.TRIANGLES)
}

// DrawLines draws the Renderable using gl.LINES mode instead of gl.TRIANGLES.
func (dr *DeferredRenderer) DrawLines(r *Renderable, shader *RenderShader, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			dr.DrawLines(child, shader, binder, perspective, view)
		}
		return
	}

	bindAndDraw(dr, r, shader, binder, perspective, view, gl.LINES)
}

// RenderLoop keeps running a render loop function until MainWindow is
// set to should close
func (dr *DeferredRenderer) RenderLoop() {
	dr.lastFrameTime = time.Now()
	for !dr.MainWindow.ShouldClose() {
		currentFrameTime := time.Now()
		deltaFrameTime := float32(currentFrameTime.Sub(dr.lastFrameTime).Seconds())

		// setup the camera matrixes
		tempW, tempH := dr.MainWindow.GetFramebufferSize()
		currentWidth, currentHeight := int32(tempW), int32(tempH)
		if dr.width != currentWidth || dr.height != currentHeight {
			fmt.Printf("Updating resoluation to %d,%d\n", currentWidth, currentHeight)
			dr.ChangeResolution(currentWidth, currentHeight)
			dr.width, dr.height = currentWidth, currentHeight
			dr.OnScreenSizeChangedFn(dr, currentWidth, currentHeight)
		}

		////////////////////////////////////////////////////////////////////////////
		// BEFORE DRAW
		dr.BeforeDrawFn(dr, deltaFrameTime)

		////////////////////////////////////////////////////////////////////////////
		// GEOMETRY PASS
		// setup the view matrixes
		gl.DepthMask(true)
		//gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT) // necessary?
		gl.Enable(gl.DEPTH_TEST)
		gl.Disable(gl.BLEND)

		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, dr.Frame)
		gl.Viewport(0, 0, dr.width, dr.height)
		buffsToClear := []uint32{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1, gl.COLOR_ATTACHMENT2}
		gl.DrawBuffers(3, &buffsToClear[0])
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// do the geometry pass on the renderables
		dr.GeometryPassFn(dr, deltaFrameTime)

		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
		gl.DepthMask(false)
		gl.Disable(gl.DEPTH_TEST)

		////////////////////////////////////////////////////////////////////////////
		// COMPOSITE PASS START
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Enable(gl.BLEND)
		gl.BlendEquation(gl.FUNC_ADD)
		gl.BlendFunc(gl.ONE, gl.ONE)

		dr.CompositePassFn(dr, deltaFrameTime)

		gl.BindVertexArray(0)

		dr.MainWindow.SwapBuffers()
		glfw.PollEvents()

		////////////////////////////////////////////////////////////////////////////
		// AFTER DRAW
		dr.AfterDrawFn(dr, deltaFrameTime)

		dr.lastFrameTime = currentFrameTime
	}
}
