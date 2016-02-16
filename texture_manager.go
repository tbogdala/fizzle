// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

// TextureManager provides an easy way to load textures to OpenGL and
// to access the textures by name elsewhere.
type TextureManager struct {
	// storage keeps references to the OpenGL texture objects referenced by name.
	storage map[string]graphics.Texture
}

// NewTextureManager creates a new TextureManager object with empty storage.
func NewTextureManager() *TextureManager {
	tm := new(TextureManager)
	tm.storage = make(map[string]graphics.Texture)
	return tm
}

// Destroy deletes all of the stored textures from OpenGL
// and resets the storage map.
func (tm *TextureManager) Destroy() {
	for _, t := range tm.storage {
		gfx.DeleteTexture(t)
	}
	tm.storage = make(map[string]graphics.Texture)
}

// GetTexture attempts to access the texture by name in storage and returns
// the OpenGL object and a bool indicating if the texture was found in storage.
func (tm *TextureManager) GetTexture(keyToUse string) (graphics.Texture, bool) {
	// try loading from storage
	glTexture, okay := tm.storage[keyToUse]
	return glTexture, okay
}

// LoadTexture loads a texture specified by path into OpenGL and then
// stores the object in the storage map under the specified keyToUse.
func (tm *TextureManager) LoadTexture(keyToUse string, path string) (graphics.Texture, error) {
	// load the file into a GL texture
	glTexture, err := LoadImageToTexture(path)
	if err != nil {
		return glTexture, err
	}

	// store it for later
	tm.storage[keyToUse] = glTexture
	return glTexture, nil
}
