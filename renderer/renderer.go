// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

/*

Package renderer is a package that defines a common interface for the
deferred and forward renderers.

Client applications will need to import a subpackage to create
instances of concrete implementations of Renderer.

*/
package renderer

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

var (
	shaderTexUniformNames      [fizzle.MaxCustomTextures]string
	shaderTexValidUniformNames [fizzle.MaxCustomTextures]string
)

func init() {
	for i := 0; i < fizzle.MaxCustomTextures; i++ {
		shaderTexUniformNames[i] = fmt.Sprintf("MATERIAL_TEX_%d", i)
		shaderTexValidUniformNames[i] = fmt.Sprintf("MATERIAL_TEX_%d_VALID", i)
	}
}

// Renderer is the common interface between the built-in deferred or forward
// style renderers.
type Renderer interface {
	// Init should initialize the renderer.
	Init(width, height int32) error

	// Destroy should free all resources needed for the renderer.
	Destroy()

	// ChangeResolution should advise the renderer there's a new screen
	// resolution to render too.
	ChangeResolution(width, height int32)

	// GetResolution returns the width and height of the renderer's known screen size.
	GetResolution() (int32, int32)

	// GetGraphics returns an object that can be used to make low-level OpenGL calls.
	GetGraphics() graphics.GraphicsProvider

	// SetGraphics should set the OpenGL implementation the renderer should use.
	SetGraphics(gp graphics.GraphicsProvider)

	// DrawRenderable draws the Renderable with the shader specified on the object.
	DrawRenderable(r *fizzle.Renderable, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera)

	// DrawRenderableWithShader will draw the Renderable with the shader specified
	// in the function call instead of the one in the object.
	DrawRenderableWithShader(r *fizzle.Renderable, shader *fizzle.RenderShader, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera)

	// DrawLines draws the renderable as a GL_LINES type of object.
	DrawLines(r *fizzle.Renderable, shader *fizzle.RenderShader, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera)

	// EndRenderFrame should be called to finish the rendering of a frame.
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
	if shaderDiffuse >= 0 && r.Material != nil {
		gfx.Uniform4f(shaderDiffuse, r.Material.DiffuseColor[0], r.Material.DiffuseColor[1], r.Material.DiffuseColor[2], r.Material.DiffuseColor[3])
	}

	shaderSpecular := shader.GetUniformLocation("MATERIAL_SPECULAR")
	if shaderSpecular >= 0 && r.Material != nil {
		gfx.Uniform4f(shaderSpecular, r.Material.SpecularColor[0], r.Material.SpecularColor[1], r.Material.SpecularColor[2], r.Material.SpecularColor[3])
	}

	shaderShiny := shader.GetUniformLocation("MATERIAL_SHININESS")
	if shaderShiny >= 0 && r.Material != nil {
		gfx.Uniform1f(shaderShiny, r.Material.Shininess)
	}

	shaderMatTexDiff := shader.GetUniformLocation("MATERIAL_TEX_DIFFUSE")
	if shaderMatTexDiff >= 0 && r.Material != nil {
		gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0 + uint32(texturesBound)))
		gfx.BindTexture(graphics.TEXTURE_2D, r.Material.DiffuseTex)
		gfx.Uniform1i(shaderMatTexDiff, texturesBound)
		texturesBound++

		shaderMatTexDiffValid := shader.GetUniformLocation("MATERIAL_TEX_DIFFUSE_VALID")
		if shaderMatTexDiffValid >= 0 {
			if r.Material.DiffuseTex > 0 {
				gfx.Uniform1f(shaderMatTexDiffValid, 1.0)
			} else {
				gfx.Uniform1f(shaderMatTexDiffValid, 0.0)
			}
		}
	}

	shaderMatTexNorms := shader.GetUniformLocation("MATERIAL_TEX_NORMALS")
	if shaderMatTexNorms >= 0 && r.Material != nil {
		gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0 + uint32(texturesBound)))
		gfx.BindTexture(graphics.TEXTURE_2D, r.Material.NormalsTex)
		gfx.Uniform1i(shaderMatTexNorms, texturesBound)
		texturesBound++

		shaderMatTexNormsValid := shader.GetUniformLocation("MATERIAL_TEX_NORMALS_VALID")
		if shaderMatTexNormsValid >= 0 {
			if r.Material.NormalsTex > 0 {
				gfx.Uniform1f(shaderMatTexNormsValid, 1.0)
			} else {
				gfx.Uniform1f(shaderMatTexNormsValid, 0.0)
			}
		}
	}

	shaderMatTexSpec := shader.GetUniformLocation("MATERIAL_TEX_SPECULAR")
	if shaderMatTexSpec >= 0 && r.Material != nil {
		gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0 + uint32(texturesBound)))
		gfx.BindTexture(graphics.TEXTURE_2D, r.Material.SpecularTex)
		gfx.Uniform1i(shaderMatTexSpec, texturesBound)
		texturesBound++

		shaderMatTexSpecValid := shader.GetUniformLocation("MATERIAL_TEX_SPECULAR_VALID")
		if shaderMatTexSpecValid >= 0 {
			if r.Material.SpecularTex > 0 {
				gfx.Uniform1f(shaderMatTexSpecValid, 1.0)
			} else {
				gfx.Uniform1f(shaderMatTexSpecValid, 0.0)
			}
		}
	}

	for texI := 0; texI < fizzle.MaxCustomTextures; texI++ {
		shaderTex := shader.GetUniformLocation(shaderTexUniformNames[texI])
		if shaderTex >= 0 {
			gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0 + uint32(texturesBound)))
			gfx.BindTexture(graphics.TEXTURE_2D, r.Material.CustomTex[texI])
			gfx.Uniform1i(shaderTex, texturesBound)
			texturesBound++

			shaderTexValid := shader.GetUniformLocation(shaderTexValidUniformNames[texI])
			if shaderTexValid >= 0 {
				if r.Material.CustomTex[texI] > 0 {
					gfx.Uniform1f(shaderTexValid, 1.0)
				} else {
					gfx.Uniform1f(shaderTexValid, 0.0)
				}
			}
		}
	}

	shaderHasBones := shader.GetUniformLocation("HAS_BONES")
	if shaderHasBones >= 0 {
		if r.Core.Skeleton != nil && len(r.Core.Skeleton.Bones) > 0 {
			gfx.Uniform1f(shaderHasBones, 1.0)
		} else {
			gfx.Uniform1f(shaderHasBones, 0.0)

		}
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

	shaderVertColor := shader.GetAttribLocation("VERTEX_COLOR")
	if shaderVertColor >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertColorVBO)
		gfx.EnableVertexAttribArray(uint32(shaderVertColor))
		gfx.VertexAttribPointer(uint32(shaderVertColor), 4, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.VertColorVBOOffset))
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
	if shaderTangent >= 0 && r.Core.TangentsVBO > 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.TangentsVBO)
		gfx.EnableVertexAttribArray(uint32(shaderTangent))
		gfx.VertexAttribPointer(uint32(shaderTangent), 3, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.TangentsVBOOffset))
	}

	if r.Core.Skeleton != nil {
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
