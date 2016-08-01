// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package deferred

import (
	"fmt"
	"time"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
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
	Frame          graphics.Buffer
	Depth          graphics.Buffer
	Diffuse        graphics.Texture
	Positions      graphics.Texture
	Normals        graphics.Texture
	CompositePlane *Renderable

	// GeometryPass is the function called to render geometry to the
	// framebuffers in the deferred renderer.
	GeometryPass DeferredGeometryPass

	// CompositePass is the function called to render the framebuffers
	// to the screen in the deferred renderer.
	CompositePass DeferredCompositePass

	// BeforeDraw is the function called by the renderer before
	// endtering the geometry draw function.
	BeforeDraw DeferredBeforeDraw

	// AfterDraw is the function called by the renderer after
	// endtering the geometry draw function.
	AfterDraw DeferredAfterDraw

	// OnScreenSizeChanged is the function called by the renderer after
	// a screen size change is detected.
	OnScreenSizeChanged ScreenSizeChanged

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
	dr.OnScreenSizeChanged = func(r *DeferredRenderer, width int32, height int32) {}
	dr.BeforeDraw = func(r *DeferredRenderer, deltaFrameTime float32) {}
	dr.AfterDraw = func(r *DeferredRenderer, deltaFrameTime float32) {}
	dr.GeometryPass = func(dr *DeferredRenderer, deltaFrameTime float32) {}
	dr.CompositePass = func(dr *DeferredRenderer, deltaFrameTime float32) {}

	return dr
}

// Destroy releases all of the OpenGL buffers the DeferredRenderer is holding on to.
func (dr *DeferredRenderer) Destroy() {
	gfx.DeleteRenderbuffer(dr.Depth)
	gfx.DeleteTexture(dr.Diffuse)
	gfx.DeleteTexture(dr.Positions)
	gfx.DeleteTexture(dr.Normals)
	gfx.DeleteFramebuffer(dr.Frame)
	dr.CompositePlane.Core.DestroyCore()
}

// ChangeResolution internally changes the size of the framebuffers and compositing
// plane that are used for rendering.
func (dr *DeferredRenderer) ChangeResolution(width, height int32) {
	dr.Destroy()
	dr.Init(width, height)
	if dr.OnScreenSizeChanged != nil {
		dr.OnScreenSizeChanged(dr, width, height)
	}
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
	dr.Frame = gfx.GenFramebuffer()

	// setup the depth buffer
	dr.Depth = gfx.GenRenderbuffer()
	gfx.BindRenderbuffer(graphics.RENDERBUFFER, dr.Depth)
	gfx.RenderbufferStorage(graphics.RENDERBUFFER, graphics.DEPTH_COMPONENT24, width, height)

	// setup the diffuse texture
	dr.Diffuse = gfx.GenTexture()
	gfx.ActiveTexture(graphics.TEXTURE0)
	gfx.BindTexture(graphics.TEXTURE_2D, dr.Diffuse)
	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.RGBA32F, width, height, 0, graphics.RGBA, graphics.FLOAT, nil, 0)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.CLAMP_TO_EDGE)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.CLAMP_TO_EDGE)

	// setup the positions texture
	dr.Positions = gfx.GenTexture()
	gfx.ActiveTexture(graphics.TEXTURE1)
	gfx.BindTexture(graphics.TEXTURE_2D, dr.Positions)
	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.RGBA32F, width, height, 0, graphics.RGBA, graphics.FLOAT, nil, 0)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.CLAMP_TO_EDGE)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.CLAMP_TO_EDGE)

	// setup the normals texture
	dr.Normals = gfx.GenTexture()
	gfx.ActiveTexture(graphics.TEXTURE2)
	gfx.BindTexture(graphics.TEXTURE_2D, dr.Normals)
	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.RGBA16F, width, height, 0, graphics.RGBA, graphics.FLOAT, nil, 0)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.CLAMP_TO_EDGE)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.CLAMP_TO_EDGE)

	// now bind all of these things to the framebuffer
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, dr.Frame)
	gfx.FramebufferRenderbuffer(graphics.FRAMEBUFFER, graphics.DEPTH_ATTACHMENT, graphics.RENDERBUFFER, dr.Depth)
	gfx.FramebufferTexture2D(graphics.FRAMEBUFFER, graphics.COLOR_ATTACHMENT0, graphics.TEXTURE_2D, dr.Diffuse, 0)
	gfx.FramebufferTexture2D(graphics.FRAMEBUFFER, graphics.COLOR_ATTACHMENT1, graphics.TEXTURE_2D, dr.Positions, 0)
	gfx.FramebufferTexture2D(graphics.FRAMEBUFFER, graphics.COLOR_ATTACHMENT2, graphics.TEXTURE_2D, dr.Normals, 0)

	// how did it all go? lets find out ...
	status := gfx.CheckFramebufferStatus(graphics.FRAMEBUFFER)
	switch {
	case status == graphics.FRAMEBUFFER_UNSUPPORTED:
		return fmt.Errorf("Failed to create the deferred rendering pipeline because the framebuffer was unsupported.\n")
	case status != graphics.FRAMEBUFFER_COMPLETE:
		return fmt.Errorf("Failed to create the deferred rendering pipeline. Code 0x%x\n", status)
	}

	gfx.BindFramebuffer(graphics.FRAMEBUFFER, 0)

	// create a plane for the composite pass
	groggy.Logsf("DEBUG", "Creatiing composite plane %dx%d.", width, height)
	cp := CreatePlaneXY("composite", 0, 0, float32(width), float32(height))
	cp.Core.Tex0 = gfx.GenTexture()
	gfx.BindTexture(graphics.TEXTURE_2D, cp.Core.Tex0)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.REPEAT)
	gfx.TexParameterf(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.REPEAT)
	gfx.BindTexture(graphics.TEXTURE_2D, 0)
	dr.CompositePlane = cp

	return nil
}

// InitShaders sets up the special shaders used in a deferred rendering pipeline.
func (dr *DeferredRenderer) InitShaders(compositeBaseFilepath string, dirlightShaderFilepath string) error {
	// Load the composite pass shader and assert variables exist
	prog, err := LoadShaderProgramFromFiles(compositeBaseFilepath, func(p graphics.Program) {
		gfx.BindFragDataLocation(p, 0, "frag_color")
	})
	if err != nil {
		return fmt.Errorf("Failed to compile and link the deferred render composite program! %v", err)
	}
	dr.shaders["composite"] = prog

	// Load the directional light shader and assert variables exist
	prog, err = LoadShaderProgramFromFiles(dirlightShaderFilepath, func(p graphics.Program) {
		gfx.BindFragDataLocation(p, 0, "frag_color")
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
	gfx.UseProgram(shader.Prog)
	gfx.BindVertexArray(r.Core.Vao)

	model := r.GetTransformMat4()

	shaderMvp := shader.GetUniformLocation("MVP_MATRIX")
	if shaderMvp >= 0 {
		mvp := ortho.Mul4(model)
		gfx.UniformMatrix4fv(shaderMvp, 1, false, mvp)
	}

	shaderPosition := shader.GetAttribLocation("VERTEX_POSITION")
	if shaderPosition >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
		gfx.EnableVertexAttribArray(uint32(shaderPosition))
		gfx.VertexAttribPointer(uint32(shaderPosition), 3, graphics.FLOAT, false, 0, gfx.PtrOffset(0))
	}

	shaderVertUv := shader.GetAttribLocation("VERTEX_UV_0")
	if shaderVertUv >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.UvVBO)
		gfx.EnableVertexAttribArray(uint32(shaderVertUv))
		gfx.VertexAttribPointer(uint32(shaderVertUv), 2, graphics.FLOAT, false, 0, gfx.PtrOffset(0))
	}

	shaderTex0 := shader.GetUniformLocation("DIFFUSE_TEX")
	if shaderTex0 >= 0 {
		gfx.ActiveTexture(graphics.TEXTURE0)
		gfx.BindTexture(graphics.TEXTURE_2D, dr.Diffuse)
		gfx.Uniform1i(shaderTex0, 0)
	}

	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.DrawElements(graphics.TRIANGLES, int32(r.FaceCount*3), graphics.UNSIGNED_INT, gfx.PtrOffset(0))
	gfx.BindVertexArray(0)
}

// DrawDirectionalLight draws the composite plane while lighting everything with
// a directional light using the parameters specified.
func (dr *DeferredRenderer) DrawDirectionalLight(eye mgl.Vec3, dir mgl.Vec3, color mgl.Vec3, ambient float32, diffuse float32, specular float32) {
	// the view matrix would be identity
	ortho := mgl.Ortho(0, float32(dr.width), 0, float32(dr.height), -200.0, 200.0)

	r := dr.CompositePlane
	shader := dr.shaders["directional_light"]
	gfx.UseProgram(shader.Prog)
	gfx.BindVertexArray(r.Core.Vao)

	model := r.GetTransformMat4()

	shaderMvp := shader.GetUniformLocation("MVP_MATRIX")
	if shaderMvp >= 0 {
		mvp := ortho.Mul4(model)
		gfx.UniformMatrix4fv(shaderMvp, 1, false, mvp)
	}

	shaderPosition := shader.GetAttribLocation("VERTEX_POSITION")
	if shaderPosition >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
		gfx.EnableVertexAttribArray(uint32(shaderPosition))
		gfx.VertexAttribPointer(uint32(shaderPosition), 3, graphics.FLOAT, false, 0, gfx.PtrOffset(0))
	}

	shaderVertUv := shader.GetAttribLocation("VERTEX_UV_0")
	if shaderVertUv >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.UvVBO)
		gfx.EnableVertexAttribArray(uint32(shaderVertUv))
		gfx.VertexAttribPointer(uint32(shaderVertUv), 2, graphics.FLOAT, false, 0, gfx.PtrOffset(0))
	}

	shaderEyePosition := shader.GetAttribLocation("EYE_WORLD_POSITION")
	if shaderEyePosition >= 0 {
		gfx.Uniform3f(shaderEyePosition, eye[0], eye[1], eye[2])
	}

	shaderTex0 := shader.GetUniformLocation("DIFFUSE_TEX")
	if shaderTex0 >= 0 {
		gfx.ActiveTexture(graphics.TEXTURE0)
		gfx.BindTexture(graphics.TEXTURE_2D, dr.Diffuse)
		gfx.Uniform1i(shaderTex0, 0)
	}

	shaderTex1 := shader.GetUniformLocation("POSITIONS_TEX")
	if shaderTex1 >= 0 {
		gfx.ActiveTexture(graphics.TEXTURE1)
		gfx.BindTexture(graphics.TEXTURE_2D, dr.Positions)
		gfx.Uniform1i(shaderTex1, 1)
	}

	shaderTex2 := shader.GetUniformLocation("NORMALS_TEX")
	if shaderTex2 >= 0 {
		gfx.ActiveTexture(graphics.TEXTURE2)
		gfx.BindTexture(graphics.TEXTURE_2D, dr.Normals)
		gfx.Uniform1i(shaderTex2, 2)
	}

	shaderLightDir := shader.GetUniformLocation("LIGHT_DIRECTION")
	if shaderLightDir >= 0 {
		gfx.Uniform3f(shaderLightDir, dir[0], dir[1], dir[2])
	}
	shaderLightColor := shader.GetUniformLocation("LIGHT_COLOR")
	if shaderLightColor >= 0 {
		gfx.Uniform3f(shaderLightColor, color[0], color[1], color[2])
	}
	shaderLightAmbient := shader.GetUniformLocation("LIGHT_AMBIENT_INTENSITY")
	if shaderLightAmbient >= 0 {
		gfx.Uniform1f(shaderLightAmbient, ambient)
	}
	shaderLightDiffuse := shader.GetUniformLocation("LIGHT_DIFFUSE_INTENSITY")
	if shaderLightDiffuse >= 0 {
		gfx.Uniform1f(shaderLightDiffuse, diffuse)
	}
	shaderLightSpecPow := shader.GetUniformLocation("LIGHT_SPECULAR_POWER")
	if shaderLightSpecPow >= 0 {
		gfx.Uniform1f(shaderLightSpecPow, specular)
	}

	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.DrawElements(graphics.TRIANGLES, int32(r.FaceCount*3), graphics.UNSIGNED_INT, gfx.PtrOffset(0))
	gfx.BindVertexArray(0)
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

	bindAndDraw(dr, r, r.Core.Shader, binder, perspective, view, graphics.TRIANGLES)
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

	bindAndDraw(dr, r, shader, binder, perspective, view, graphics.TRIANGLES)
}

// DrawLines draws the Renderable using graphics.LINES mode instead of graphics.TRIANGLES.
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

	bindAndDraw(dr, r, shader, binder, perspective, view, graphics.LINES)
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
		}

		////////////////////////////////////////////////////////////////////////////
		// BEFORE DRAW
		dr.BeforeDraw(dr, deltaFrameTime)

		////////////////////////////////////////////////////////////////////////////
		// GEOMETRY PASS
		// setup the view matrixes
		gfx.DepthMask(true)
		//gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT) // necessary?
		gfx.Enable(graphics.DEPTH_TEST)
		gfx.Disable(graphics.BLEND)

		gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, dr.Frame)
		gfx.Viewport(0, 0, dr.width, dr.height)
		buffsToClear := []uint32{graphics.COLOR_ATTACHMENT0, graphics.COLOR_ATTACHMENT1, graphics.COLOR_ATTACHMENT2}
		gfx.DrawBuffers(buffsToClear)
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		// do the geometry pass on the renderables
		dr.GeometryPass(dr, deltaFrameTime)

		gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, 0)
		gfx.DepthMask(false)
		gfx.Disable(graphics.DEPTH_TEST)

		////////////////////////////////////////////////////////////////////////////
		// COMPOSITE PASS START
		gfx.Clear(graphics.COLOR_BUFFER_BIT)
		gfx.Enable(graphics.BLEND)
		gfx.BlendEquation(graphics.FUNC_ADD)
		gfx.BlendFunc(graphics.ONE, graphics.ONE)

		dr.CompositePass(dr, deltaFrameTime)

		gfx.BindVertexArray(0)

		dr.MainWindow.SwapBuffers()
		glfw.PollEvents()

		////////////////////////////////////////////////////////////////////////////
		// AFTER DRAW
		dr.AfterDraw(dr, deltaFrameTime)

		dr.lastFrameTime = currentFrameTime
	}
}
