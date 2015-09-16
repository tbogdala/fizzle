// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"fmt"

	gl "github.com/go-gl/gl/v3.3-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
)

// Renderer is the common interface between the built-in deferred or forward
// style renderers.
type Renderer interface {
	Init(width, height int32) error
	Destroy()
	ChangeResolution(width, height int32)
	GetResolution() (int32, int32)

	DrawRenderable(r *Renderable, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4)
	DrawRenderableWithShader(r *Renderable, shader *RenderShader, binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4)
	EndRenderFrame()
}

// RenderBinder is the type of the function called when binding shader variables
// which allows for custom binding of VBO objects.
type RenderBinder func(renderer Renderer, r *Renderable, shader *RenderShader)

func bindAndDraw(renderer Renderer, r *Renderable, shader *RenderShader,
	binder RenderBinder, perspective mgl.Mat4, view mgl.Mat4, mode uint32) {
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
		gl.Uniform4f(shaderDiffuse, r.Core.DiffuseColor[0], r.Core.DiffuseColor[1], r.Core.DiffuseColor[2], r.Core.DiffuseColor[3])
	}

	shaderSpecular := shader.GetUniformLocation("MATERIAL_SPECULAR")
	if shaderSpecular >= 0 {
		gl.Uniform4f(shaderSpecular, r.Core.SpecularColor[0], r.Core.SpecularColor[1], r.Core.SpecularColor[2], r.Core.SpecularColor[3])
	}

	shaderShiny := shader.GetUniformLocation("MATERIAL_SHININESS")
	if shaderShiny >= 0 {
		gl.Uniform1f(shaderShiny, r.Core.Shininess)
	}

	shaderTex1 := shader.GetUniformLocation("MATERIAL_TEX_0")
	if shaderTex1 >= 0 {
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, r.Core.Tex0)
		gl.Uniform1i(shaderTex1, 0)
	}

	shaderBones := shader.GetUniformLocation("BONES")
	if shaderBones >= 0 && r.Core.Skeleton != nil && len(r.Core.Skeleton.Bones) > 0 {
		gl.UniformMatrix4fv(shaderBones, int32(len(r.Core.Skeleton.Bones)), false, &(r.Core.Skeleton.PoseTransforms[0][0]))
	}

	// do some special binding for the differe Renderer types if necessary
	if forwardRenderer, okay := renderer.(*ForwardRenderer); okay {
		var lightCount int32 = int32(forwardRenderer.GetActiveLightCount())
		if lightCount >= 1 {
			for lightI := 0; lightI < int(lightCount); lightI++ {
				light := forwardRenderer.ActiveLights[lightI]

				// NOTE: this gets bound in eye-space coordinates
				shaderLightPosition := shader.GetUniformLocation(fmt.Sprintf("LIGHT_POSITION[%d]", lightI))
				if shaderLightPosition >= 0 {
					lightPosEyeSpace := view.Mul4x1(mgl.Vec4{light.Position[0], light.Position[1], light.Position[2], 1.0})
					gl.Uniform3fv(shaderLightPosition, 1, &lightPosEyeSpace[0])
				}

				shaderLightDirection := shader.GetUniformLocation(fmt.Sprintf("LIGHT_DIRECTION[%d]", lightI))
				if shaderLightDirection >= 0 {
					gl.Uniform3fv(shaderLightDirection, 1, &light.Direction[0])
				}

				shaderLightDiffuse := shader.GetUniformLocation(fmt.Sprintf("LIGHT_DIFFUSE[%d]", lightI))
				if shaderLightDiffuse >= 0 {
					gl.Uniform4fv(shaderLightDiffuse, 1, &light.DiffuseColor[0])
				}

				shaderLightIntensity := shader.GetUniformLocation(fmt.Sprintf("LIGHT_DIFFUSE_INTENSITY[%d]", lightI))
				if shaderLightIntensity >= 0 {
					gl.Uniform1fv(shaderLightIntensity, 1, &light.DiffuseIntensity)
				}

				shaderLightAmbientIntensity := shader.GetUniformLocation(fmt.Sprintf("LIGHT_AMBIENT_INTENSITY[%d]", lightI))
				if shaderLightAmbientIntensity >= 0 {
					gl.Uniform1fv(shaderLightAmbientIntensity, 1, &light.AmbientIntensity)
				}

				shaderLightAttenuation := shader.GetUniformLocation(fmt.Sprintf("LIGHT_ATTENUATION[%d]", lightI))
				if shaderLightAttenuation >= 0 {
					gl.Uniform1fv(shaderLightAttenuation, 1, &light.Attenuation)
				}
			} // lightI
			shaderLightCount := shader.GetUniformLocation("LIGHT_COUNT")
			if shaderLightCount >= 0 {
				gl.Uniform1iv(shaderLightCount, 1, &lightCount)
			}
		}
	}

	shaderCameraWorldPos := shader.GetUniformLocation("CAMERA_WORLD_POSITION")
	if shaderCameraWorldPos >= 0 {
		gl.Uniform3f(shaderCameraWorldPos, -view[12], -view[13], -view[14])
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

	shaderBoneFids := shader.GetAttribLocation("VERTEX_BONE_IDS")
	if shaderBoneFids >= 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.BoneFidsVBO)
		gl.EnableVertexAttribArray(uint32(shaderBoneFids))
		gl.VertexAttribPointer(uint32(shaderBoneFids), 4, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}

	shaderBoneWeights := shader.GetAttribLocation("VERTEX_BONE_WEIGHTS")
	if shaderBoneWeights >= 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.BoneWeightsVBO)
		gl.EnableVertexAttribArray(uint32(shaderBoneWeights))
		gl.VertexAttribPointer(uint32(shaderBoneWeights), 4, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}

	// if a custom binder function was passed in then call it
	if binder != nil {
		binder(renderer, r, shader)
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.DrawElements(mode, int32(r.FaceCount*3), gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)
}
