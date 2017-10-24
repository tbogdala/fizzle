// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	assimp "github.com/tbogdala/assimp-go"
	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/fizzle/component"
)

// meshRenderable is used to tie together state for the component mesh,
// the renderable for this component mesh and any other state information relating.
type meshRenderable struct {
	ComponentMesh     *component.Mesh
	Renderable        *fizzle.Renderable
	AnimationsEnabled []bool
}

// getComponentPrefix gets the prefix directory for the current component filename.
func (s *State) getComponentPrefix() string {
	comp := s.components.activeComponent
	if comp == nil {
		return ""
	}

	return comp.GetDirPath()
}

func (s *State) makeRenderableForMesh(compMesh *component.Mesh) (*meshRenderable, error) {
	// the component manager should have already attempted to load the mesh and put
	// the result in SrcMesh ... however this will only work for binary file references.
	// if the SrcMesh is still nil, attempt to load from a SrcFile if it's present in the component.
	if compMesh.SrcMesh == nil && compMesh.SrcFile != "" {
		prefixDir := s.getComponentPrefix()
		meshFilepath := prefixDir + compMesh.SrcFile
		srcMeshes, parseErr := assimp.ParseFile(meshFilepath)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to load source mesh %s: %v", meshFilepath, parseErr)
		}

		// NOTE: only the first mesh is actually used in a given source file
		if len(srcMeshes) > 0 {
			compMesh.SrcMesh = srcMeshes[0]
		}
	}

	// if we haven't loaded something by now, then return a nil renderable
	if compMesh.SrcMesh == nil {
		return nil, fmt.Errorf("failed to load a mesh from either source or binary paths")
	}

	r, err := component.CreateRenderableForMesh(s.texMan, s.shaders, compMesh)
	if err != nil {
		return nil, fmt.Errorf("failed to create a renderable for the component mesh %s: %v", compMesh.Name, err)
	}

	compRenderable := new(meshRenderable)
	compRenderable.ComponentMesh = compMesh
	compRenderable.Renderable = r

	// setup the animation enable flag slice
	compRenderable.AnimationsEnabled = []bool{}
	for i := 0; i < len(compMesh.SrcMesh.Animations); i++ {
		compRenderable.AnimationsEnabled = append(compRenderable.AnimationsEnabled, false)
	}

	return compRenderable, nil
}

// updateVisibleMesh copies the settings from the ComponentMesh part of meshRenderable
// to the Renderable so that it renders correctly.
// This also gets attempts to get textures from textureMan as well.
func (s *State) updateVisibleMesh(compRenderable *meshRenderable) {
	// push all settings from the component to the renderable
	compRenderable.Renderable.Location = compRenderable.ComponentMesh.Offset
	compRenderable.Renderable.Scale = compRenderable.ComponentMesh.Scale
	compRenderable.Renderable.Material.DiffuseColor = compRenderable.ComponentMesh.Material.Diffuse
	if compRenderable.ComponentMesh.RotationDegrees != 0.0 {
		compRenderable.Renderable.LocalRotation = mgl.QuatRotate(
			mgl.DegToRad(compRenderable.ComponentMesh.RotationDegrees),
			compRenderable.ComponentMesh.RotationAxis)
	}

	compRenderable.Renderable.Material.SpecularColor = compRenderable.ComponentMesh.Material.Specular
	compRenderable.Renderable.Material.Shininess = compRenderable.ComponentMesh.Material.Shininess

	// FIXME: lots of silent fails below

	// try to find a shader
	shader, shaderFound := s.shaders[compRenderable.ComponentMesh.Material.ShaderName]
	if shaderFound {
		compRenderable.Renderable.Material.Shader = shader
	}

	// assign textures
	textures := compRenderable.ComponentMesh.Material.Textures
	for i := 0; i < len(textures); i++ {
		glTex, texFound := s.texMan.GetTexture(textures[i])
		if texFound && i < fizzle.MaxCustomTextures {
			compRenderable.Renderable.Material.CustomTex[i] = glTex
		}
	}
	if len(compRenderable.ComponentMesh.Material.DiffuseTexture) > 0 {
		glTex, texFound := s.texMan.GetTexture(compRenderable.ComponentMesh.Material.DiffuseTexture)
		if texFound {
			compRenderable.Renderable.Material.DiffuseTex = glTex
		}
	}
	if len(compRenderable.ComponentMesh.Material.NormalsTexture) > 0 {
		glTex, texFound := s.texMan.GetTexture(compRenderable.ComponentMesh.Material.NormalsTexture)
		if texFound {
			compRenderable.Renderable.Material.NormalsTex = glTex
		}
	}
	if len(compRenderable.ComponentMesh.Material.SpecularTexture) > 0 {
		glTex, texFound := s.texMan.GetTexture(compRenderable.ComponentMesh.Material.SpecularTexture)
		if texFound {
			compRenderable.Renderable.Material.SpecularTex = glTex
		}
	}

}

// updateChildComponentRenderable copies the location, scale and rotation from the
// child component reference to the renderable object.
func updateChildComponentRenderable(childRenderable *fizzle.Renderable, childComp *component.ChildRef) {
	// push all settings from the child component to the renderable
	childRenderable.Location = childComp.Location
	childRenderable.Scale = childComp.Scale
	if childComp.RotationDegrees != 0.0 {
		childRenderable.LocalRotation = mgl.QuatRotate(mgl.DegToRad(childComp.RotationDegrees), childComp.RotationAxis)
	}
}
