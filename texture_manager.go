// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle


type TextureManager struct {
  storage map[string]uint32
}

func NewTextureManager() *TextureManager {
  tm := new(TextureManager)
	tm.storage = make(map[string]uint32)
	return tm
}

func (tm *TextureManager) GetTexture(keyToUse string) (uint32, bool) {
  // try loading from storage
  glTexture,  okay := tm.storage[keyToUse]
  return glTexture, okay
}

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
