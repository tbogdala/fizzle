// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"fmt"

	"github.com/tbogdala/fizzle/component"
)

// LoadComponentFile attempts to load the component JSON file into the editor
// and will return the component on success. A non-nil error is returned on failure.
func (s *State) LoadComponentFile(filepath string) (*component.Component, error) {
	theComponent, err := s.components.manager.LoadComponentFromFile(filepath)
	if err != nil {
		return nil, err
	}

	return theComponent, nil
}

// doLoadTexture loads a relative filepath texture into the
// texture manager.
func (s *State) doLoadTexture(texFile string) error {
	prefixDir := s.getComponentPrefix()
	texFilepath := prefixDir + texFile
	_, err := s.texMan.LoadTexture(texFile, texFilepath)
	if err != nil {
		return fmt.Errorf("Failed to load texture %s: %v", texFile, err)
	}

	return nil
}
