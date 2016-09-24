// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

const (
	// MaxCustomTextures is the maximum number of custom textures that can get assigned
	// to a renderable.
	MaxCustomTextures = 8
)

// Material is a type that represents the visual properties for a Renderable.
type Material struct {
	// Shader is the program used to render this material; This can be overridden
	// by using a DrawWithShader* function so that this material's shader doesn't
	// get used.
	Shader *RenderShader

	// DiffuseTex is the diffuse texture for the material.
	DiffuseTex graphics.Texture

	// NormalsTex is the normal map texture for the material.
	NormalsTex graphics.Texture

	// SpecularTex is the spcular map texture for the material.
	SpecularTex graphics.Texture

	// CustomTex is an array of textures that can be used for specific purposes
	// by client code that are not covered by other textures specified in this
	// structure.
	CustomTex [MaxCustomTextures]graphics.Texture

	// DiffuseColor is the material color for the renderable. This is
	// displayed outright by the shader or often blended with the
	// diffuse texture.
	DiffuseColor mgl.Vec4

	// SpecularColor is the material specular color for the renderable
	// and is used to control the color of the specular highlight.
	//
	// It can be thought of the topcoat layer color to the DiffuseColor's
	// base paint layer color.
	SpecularColor mgl.Vec4

	// Shininess is the specular coefficient used to control the tightness
	// of the specular highlight. It represents the power the specular factor will
	// be raised to -- therefore values between (0.0 - 1.0) will yield different
	// results than values >= 1.0.
	Shininess float32
}

// NewMaterial creates a new material with sane defaults.
func NewMaterial() *Material {
	m := new(Material)
	m.DiffuseColor = mgl.Vec4{1, 1, 1, 1}
	m.SpecularColor = mgl.Vec4{1, 1, 1, 1}
	m.Shininess = 1.0
	return m
}
