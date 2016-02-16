// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
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
	gfx.UseProgram(shader.Prog)
	gfx.BindVertexArray(r.Core.Vao)

	texturesBound := int32(0)
	model := r.GetTransformMat4()

	shaderMvp := shader.GetUniformLocation("MVP_MATRIX")
	if shaderMvp >= 0 {
		mvp := perspective.Mul4(view).Mul4(model)
		gfx.UniformMatrix4fv(shaderMvp, 1, false, &mvp[0])
	}

	shaderMv := shader.GetUniformLocation("MV_MATRIX")
	if shaderMv >= 0 {
		mv := view.Mul4(model)
		gfx.UniformMatrix4fv(shaderMv, 1, false, &mv[0])
	}

	shaderV := shader.GetUniformLocation("V_MATRIX")
	if shaderV >= 0 {
		gfx.UniformMatrix4fv(shaderV, 1, false, &view[0])
	}

	shaderM := shader.GetUniformLocation("M_MATRIX")
	if shaderM >= 0 {
		gfx.UniformMatrix4fv(shaderM, 1, false, &model[0])
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
		gfx.UniformMatrix4fv(shaderBones, int32(len(r.Core.Skeleton.Bones)), false, &(r.Core.Skeleton.PoseTransforms[0][0]))
	}

	// do some special binding for the different Renderer types if necessary
	if forwardRenderer, okay := renderer.(*ForwardRenderer); okay {
		var lightCount int32 = int32(forwardRenderer.GetActiveLightCount())
		var shadowLightCount int32 = int32(forwardRenderer.GetActiveShadowLightCount())
		if lightCount >= 1 {
			for lightI := 0; lightI < int(lightCount); lightI++ {
				light := forwardRenderer.ActiveLights[lightI]

				shaderLightPosition := shader.GetUniformLocation(fmt.Sprintf("LIGHT_POSITION[%d]", lightI))
				if shaderLightPosition >= 0 {
					gfx.Uniform3fv(shaderLightPosition, 1, &light.Position[0])
				}

				shaderLightDirection := shader.GetUniformLocation(fmt.Sprintf("LIGHT_DIRECTION[%d]", lightI))
				if shaderLightDirection >= 0 {
					gfx.Uniform3fv(shaderLightDirection, 1, &light.Direction[0])
				}

				shaderLightDiffuse := shader.GetUniformLocation(fmt.Sprintf("LIGHT_DIFFUSE[%d]", lightI))
				if shaderLightDiffuse >= 0 {
					gfx.Uniform4fv(shaderLightDiffuse, 1, &light.DiffuseColor[0])
				}

				shaderLightIntensity := shader.GetUniformLocation(fmt.Sprintf("LIGHT_DIFFUSE_INTENSITY[%d]", lightI))
				if shaderLightIntensity >= 0 {
					gfx.Uniform1fv(shaderLightIntensity, 1, &light.DiffuseIntensity)
				}

				shaderLightAmbientIntensity := shader.GetUniformLocation(fmt.Sprintf("LIGHT_AMBIENT_INTENSITY[%d]", lightI))
				if shaderLightAmbientIntensity >= 0 {
					gfx.Uniform1fv(shaderLightAmbientIntensity, 1, &light.AmbientIntensity)
				}

				shaderLightAttenuation := shader.GetUniformLocation(fmt.Sprintf("LIGHT_ATTENUATION[%d]", lightI))
				if shaderLightAttenuation >= 0 {
					gfx.Uniform1fv(shaderLightAttenuation, 1, &light.Attenuation)
				}

				shaderShadowMaps := shader.GetUniformLocation(fmt.Sprintf("SHADOW_MAPS[%d]", lightI))
				if shaderShadowMaps >= 0 {
					/* There have been problems in the past on Intel drivers on Mac OS if all of the
					   samplers are not bound to something. So this code will bind a 0 if the shadow map
						 does not exist for that light. */
					gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0 + uint32(texturesBound)))
					if light.ShadowMap != nil {
						gfx.BindTexture(graphics.TEXTURE_2D, light.ShadowMap.Texture)
					} else {
						gfx.BindTexture(graphics.TEXTURE_2D, 0)
					}
					gfx.Uniform1i(shaderShadowMaps, texturesBound)
					texturesBound++
				}

				if light.ShadowMap != nil {
					shaderShadowMatrix := shader.GetUniformLocation(fmt.Sprintf("SHADOW_MATRIX[%d]", lightI))
					if shaderShadowMatrix >= 0 {
						gfx.UniformMatrix4fv(shaderShadowMatrix, 1, false, &light.ShadowMap.BiasedMatrix[0])
					}
				}
			} // lightI

			shaderLightCount := shader.GetUniformLocation("LIGHT_COUNT")
			if shaderLightCount >= 0 {
				gfx.Uniform1iv(shaderLightCount, 1, &lightCount)
			}

			shaderShadowLightCount := shader.GetUniformLocation("SHADOW_COUNT")
			if shaderShadowLightCount >= 0 {
				gfx.Uniform1iv(shaderShadowLightCount, 1, &shadowLightCount)
			}

			if forwardRenderer.currentShadowPassLight != nil {
				shaderShadowVP := shader.GetUniformLocation("SHADOW_VP_MATRIX")
				if shaderShadowVP >= 0 {
					gfx.UniformMatrix4fv(shaderShadowVP, 1, false, &forwardRenderer.currentShadowPassLight.ShadowMap.ViewProjMatrix[0])
				}
			}

		} // lightcount
	} // forwardRenderer

	shaderCameraWorldPos := shader.GetUniformLocation("CAMERA_WORLD_POSITION")
	if shaderCameraWorldPos >= 0 {
		gfx.Uniform3f(shaderCameraWorldPos, -view[12], -view[13], -view[14])
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

	shaderNormal := shader.GetAttribLocation("VERTEX_NORMAL")
	if shaderNormal >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.NormsVBO)
		gfx.EnableVertexAttribArray(uint32(shaderNormal))
		gfx.VertexAttribPointer(uint32(shaderNormal), 3, graphics.FLOAT, false, 0, gfx.PtrOffset(0))
	}

	shaderTangent := shader.GetAttribLocation("VERTEX_TANGENT")
	if shaderTangent >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.TangentsVBO)
		gfx.EnableVertexAttribArray(uint32(shaderTangent))
		gfx.VertexAttribPointer(uint32(shaderTangent), 3, graphics.FLOAT, false, 0, gfx.PtrOffset(0))
	}

	shaderBoneFids := shader.GetAttribLocation("VERTEX_BONE_IDS")
	if shaderBoneFids >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.BoneFidsVBO)
		gfx.EnableVertexAttribArray(uint32(shaderBoneFids))
		gfx.VertexAttribPointer(uint32(shaderBoneFids), 4, graphics.FLOAT, false, 0, gfx.PtrOffset(0))
	}

	shaderBoneWeights := shader.GetAttribLocation("VERTEX_BONE_WEIGHTS")
	if shaderBoneWeights >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.BoneWeightsVBO)
		gfx.EnableVertexAttribArray(uint32(shaderBoneWeights))
		gfx.VertexAttribPointer(uint32(shaderBoneWeights), 4, graphics.FLOAT, false, 0, gfx.PtrOffset(0))
	}

	// if a custom binder function was passed in then call it
	if binder != nil {
		binder(renderer, r, shader)
	}

	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.DrawElements(graphics.Enum(mode), int32(r.FaceCount*3), graphics.UNSIGNED_INT, gfx.PtrOffset(0))
	gfx.BindVertexArray(0)
}
