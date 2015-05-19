// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	gl "github.com/go-gl/gl/v3.3-core/gl"
)

// TextureManager provides an easy way to load textures to OpenGL and
// to access the textures by name elsewhere.
type TextureManager struct {
	// storage keeps references to the OpenGL texture objects referenced by name.
	storage map[string]uint32
}

// NewTextureManager creates a new TextureManager object with empty storage.
func NewTextureManager() *TextureManager {
	tm := new(TextureManager)
	tm.storage = make(map[string]uint32)
	return tm
}

// Destroy deletes all of the stored textures from OpenGL
// and resets the storage map.
func (tm *TextureManager) Destroy() {
	for _, t := range tm.storage {
		gl.DeleteTextures(1, &t)
	}
	tm.storage = make(map[string]uint32)
}

// GetTexture attempts to access the texture by name in storage and returns
// the OpenGL object and a bool indicating if the texture was found in storage.
func (tm *TextureManager) GetTexture(keyToUse string) (uint32, bool) {
	// try loading from storage
	glTexture, okay := tm.storage[keyToUse]
	return glTexture, okay
}

// LoadTexture loads a texture specified by path into OpenGL and then
// stores the object in the storage map under the specified keyToUse.
func (tm *TextureManager) LoadTexture(keyToUse string, path string) (uint32, error) {
	// load the file into a GL texture
	glTexture, err := LoadImageToTexture(path)
	if err != nil {
		return glTexture, err
	}

	// store it for later
	tm.storage[keyToUse] = glTexture
	return glTexture, nil
}
