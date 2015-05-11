// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"fmt"
	gl "github.com/go-gl/gl/v3.3-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/groggy"
)

type DeferredRenderer struct {
	Frame          uint32
	Depth          uint32
	Diffuse        uint32
	Positions      uint32
	Normals        uint32
	CompositePlane *Renderable

	shaders                 map[string]*RenderShader
	width int32
	height int32
}

func NewDeferredRenderer() *DeferredRenderer {
	dr := new(DeferredRenderer)
	dr.shaders = make(map[string]*RenderShader)
	return dr
}

func (dr *DeferredRenderer) Destroy() {
	gl.DeleteRenderbuffers(1, &dr.Depth)
	gl.DeleteRenderbuffers(1, &dr.Diffuse)
	gl.DeleteRenderbuffers(1, &dr.Positions)
	gl.DeleteRenderbuffers(1, &dr.Normals)
	gl.DeleteFramebuffers(1, &dr.Frame)
	dr.CompositePlane.Core.DestroyCore()
}

func (dr *DeferredRenderer) ChangeResolution(width, height int32) {
	dr.Destroy()
	dr.Init(width, height)
}

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
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB32F, width, height, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// setup the positions texture
	gl.GenTextures(1, &dr.Positions)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, dr.Positions)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB32F, width, height, 0, gl.RGBA, gl.FLOAT, nil)
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
	if status != gl.FRAMEBUFFER_COMPLETE {
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

func (dr *DeferredRenderer) DrawRenderable(r *Renderable, perspective mgl.Mat4, view mgl.Mat4) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			dr.DrawRenderable(child, perspective, view)
		}
		return
	}

	dr.bindAndDraw(r, r.Core.Shader, perspective, view)
}

func (dr *DeferredRenderer) DrawRenderableWithShader(r *Renderable, shader *RenderShader, perspective mgl.Mat4, view mgl.Mat4) {
	// only draw visible nodes
	if !r.IsVisible {
		return
	}

	// if the renderable is a group, just try to draw the children
	if r.IsGroup {
		for _, child := range r.Children {
			dr.DrawRenderableWithShader(child, shader, perspective, view)
		}
		return
	}

	dr.bindAndDraw(r, shader, perspective, view)
}

func (dr *DeferredRenderer) bindAndDraw(r *Renderable, shader *RenderShader, perspective mgl.Mat4, view mgl.Mat4) {
	gl.UseProgram(shader.Prog)
	gl.BindVertexArray(r.Core.Vao)

	model := r.GetTransformMat4()

	shaderMvp := shader.GetUniformLocation("MVP_MATRIX")
	if shaderMvp >= 0 {
		mvp := perspective.Mul4(view).Mul4(model)
		gl.UniformMatrix4fv(shaderMvp, 1, false, &mvp[0])
	}

	shaderMv := shader.GetUniformLocation("MV_MATRIX")
	if shaderMv >= 0 {
		mv := view.Mul4(model)
		gl.UniformMatrix4fv(shaderMv, 1, false, &mv[0])
	}

	shaderM := shader.GetUniformLocation("M_MATRIX")
	if shaderM >= 0 {
		gl.UniformMatrix4fv(shaderM, 1, false, &model[0])
	}

	shaderDiffuse := shader.GetUniformLocation("MATERIAL_DIFFUSE")
	if shaderDiffuse >= 0 {
		// FIXME: currently unsupported
		gl.Uniform4f(shaderDiffuse, 1.0, 1.0, 1.0, 1.0)
	}

	shaderTex1 := shader.GetUniformLocation("MATERIAL_TEX_0")
	if shaderTex1 >= 0 {
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, r.Core.Tex0)
		gl.Uniform1i(shaderTex1, 0)
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

	shaderNormal := shader.GetAttribLocation("VERTEX_NORMAL")
	if shaderNormal >= 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.NormsVBO)
		gl.EnableVertexAttribArray(uint32(shaderNormal))
		gl.VertexAttribPointer(uint32(shaderNormal), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.DrawElements(gl.TRIANGLES, int32(r.FaceCount*3), gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)
}
