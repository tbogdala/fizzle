// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package renderer

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

// Renderer is the common interface between the built-in deferred or forward
// style renderers.
type Renderer interface {
	Init(width, height int32) error
	Destroy()
	ChangeResolution(width, height int32)
	GetResolution() (int32, int32)
	GetGraphics() graphics.GraphicsProvider
	SetGraphics(gp graphics.GraphicsProvider)

	DrawRenderable(r *fizzle.Renderable, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera)
	DrawRenderableWithShader(r *fizzle.Renderable, shader *fizzle.RenderShader, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera)
	DrawLines(r *fizzle.Renderable, shader *fizzle.RenderShader, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera)
	EndRenderFrame()
}

// RenderBinder is the type of the function called when binding shader variables
// which allows for custom binding of VBO objects.
type RenderBinder func(renderer Renderer, r *fizzle.Renderable, shader *fizzle.RenderShader, texturesBound *int32)

// BindAndDraw is a common shader variable binder meant to be called from the
// renderer implementations.
func BindAndDraw(renderer Renderer, r *fizzle.Renderable, shader *fizzle.RenderShader,
	binders []RenderBinder, perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera, mode uint32) {
	gfx := renderer.GetGraphics()
	gfx.UseProgram(shader.Prog)
	gfx.BindVertexArray(r.Core.Vao)

	texturesBound := int32(0)
	model := r.GetTransformMat4()

	shaderMvp := shader.GetUniformLocation("MVP_MATRIX")
	if shaderMvp >= 0 {
		mvp := perspective.Mul4(view).Mul4(model)
		gfx.UniformMatrix4fv(shaderMvp, 1, false, mvp)
	}

	shaderMv := shader.GetUniformLocation("MV_MATRIX")
	if shaderMv >= 0 {
		mv := view.Mul4(model)
		gfx.UniformMatrix4fv(shaderMv, 1, false, mv)
	}

	shaderV := shader.GetUniformLocation("V_MATRIX")
	if shaderV >= 0 {
		gfx.UniformMatrix4fv(shaderV, 1, false, view)
	}

	shaderM := shader.GetUniformLocation("M_MATRIX")
	if shaderM >= 0 {
		gfx.UniformMatrix4fv(shaderM, 1, false, model)
	}

	shaderDiffuse := shader.GetUniformLocation("MATERIAL_DIFFUSE")
	if shaderDiffuse >= 0 {
		gfx.Uniform4f(shaderDiffuse, r.Core.DiffuseColor[0], r.Core.DiffuseColor[1], r.Core.DiffuseColor[2], r.Core.DiffuseColor[3])
	}

	shaderSpecular := shader.GetUniformLocation("MATERIAL_SPECULAR")
	if shaderSpecular >= 0 {
		gfx.Uniform4f(shaderSpecular, r.Core.SpecularColor[0], r.Core.SpecularColor[1], r.Core.SpecularColor[2], r.Core.SpecularColor[3])
	}

	shaderShiny := shader.GetUniformLocation("MATERIAL_SHININESS")
	if shaderShiny >= 0 {
		gfx.Uniform1f(shaderShiny, r.Core.Shininess)
	}

	shaderTex1 := shader.GetUniformLocation("MATERIAL_TEX_0")
	if shaderTex1 >= 0 {
		gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0 + uint32(texturesBound)))
		gfx.BindTexture(graphics.TEXTURE_2D, r.Core.Tex0)
		gfx.Uniform1i(shaderTex1, texturesBound)
		texturesBound++
	}

	shaderTex2 := shader.GetUniformLocation("MATERIAL_TEX_1")
	if shaderTex2 >= 0 {
		gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0 + uint32(texturesBound)))
		gfx.BindTexture(graphics.TEXTURE_2D, r.Core.Tex1)
		gfx.Uniform1i(shaderTex2, texturesBound)
		texturesBound++
	}

	shaderBones := shader.GetUniformLocation("BONES")
	if shaderBones >= 0 && r.Core.Skeleton != nil && len(r.Core.Skeleton.Bones) > 0 {
		gfx.UniformMatrix4fv(shaderBones, int32(len(r.Core.Skeleton.Bones)), false, r.Core.Skeleton.PoseTransforms)
	}

	if camera != nil {
		shaderCameraWorldPos := shader.GetUniformLocation("CAMERA_WORLD_POSITION")
		if shaderCameraWorldPos >= 0 {
			cp := camera.GetPosition()
			gfx.Uniform3f(shaderCameraWorldPos, cp[0], cp[1], cp[2])
		}
	}

	shaderPosition := shader.GetAttribLocation("VERTEX_POSITION")
	if shaderPosition >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
		gfx.EnableVertexAttribArray(uint32(shaderPosition))
		gfx.VertexAttribPointer(uint32(shaderPosition), 3, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.VertVBOOffset))
	}

	shaderVertUv := shader.GetAttribLocation("VERTEX_UV_0")
	if shaderVertUv >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.UvVBO)
		gfx.EnableVertexAttribArray(uint32(shaderVertUv))
		gfx.VertexAttribPointer(uint32(shaderVertUv), 2, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.UvVBOOffset))
	}

	shaderNormal := shader.GetAttribLocation("VERTEX_NORMAL")
	if shaderNormal >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.NormsVBO)
		gfx.EnableVertexAttribArray(uint32(shaderNormal))
		gfx.VertexAttribPointer(uint32(shaderNormal), 3, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.NormsVBOOffset))
	}

	shaderTangent := shader.GetAttribLocation("VERTEX_TANGENT")
	if shaderTangent >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.TangentsVBO)
		gfx.EnableVertexAttribArray(uint32(shaderTangent))
		gfx.VertexAttribPointer(uint32(shaderTangent), 3, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.TangentsVBOOffset))
	}

	shaderBoneFids := shader.GetAttribLocation("VERTEX_BONE_IDS")
	if shaderBoneFids >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.BoneFidsVBO)
		gfx.EnableVertexAttribArray(uint32(shaderBoneFids))
		gfx.VertexAttribPointer(uint32(shaderBoneFids), 4, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.BoneFidsVBOOffset))
	}

	shaderBoneWeights := shader.GetAttribLocation("VERTEX_BONE_WEIGHTS")
	if shaderBoneWeights >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.BoneWeightsVBO)
		gfx.EnableVertexAttribArray(uint32(shaderBoneWeights))
		gfx.VertexAttribPointer(uint32(shaderBoneWeights), 4, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.BoneWeightsVBOOffset))
	}

	// if a custom binder function was passed in then call it
	if len(binders) > 0 {
		for _, binder := range binders {
			if binder != nil {
				binder(renderer, r, shader, &texturesBound)
			}
		}
	}

	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	if mode != graphics.LINES {
		gfx.DrawElements(graphics.Enum(mode), int32(r.FaceCount*3), graphics.UNSIGNED_INT, gfx.PtrOffset(0))
	} else {
		gfx.DrawElements(graphics.Enum(mode), int32(r.FaceCount*2), graphics.UNSIGNED_INT, gfx.PtrOffset(0))
	}
	gfx.BindVertexArray(0)
}
