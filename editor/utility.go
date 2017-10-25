// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"fmt"
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
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

func makeMouseScrollCallback(s *State) glfw.ScrollCallback {
	return func(w *glfw.Window, xoff float64, yoff float64) {
		var scale float32 = 1.0

		if w.GetKey(glfw.KeyLeftShift) == glfw.Press {
			scale = 0.1
		}

		s.orbitDist += float32(yoff) * scale
		s.camera.SetDistance(s.orbitDist)
	}
}

func makeMousePosCallback(s *State) glfw.CursorPosCallback {
	// relative to upper left corner of screen
	return func(w *glfw.Window, x float64, y float64) {
		width, height := s.window.GetSize()
		radsPerX := 2.0 * float32(math.Pi) / float32(width)
		radsPerY := 2.0 * float32(math.Pi) / float32(height)
		diffX := float32(x) - s.lastMouseX
		diffY := float32(y) - s.lastMouseY

		// if we have the RMB down we orbit the cam
		rmbStatus := w.GetMouseButton(glfw.MouseButton2)
		if rmbStatus == glfw.Press {
			s.camera.Rotate(diffX * radsPerX)
			s.camera.RotateVertical(-diffY * radsPerY)
		}

		s.lastMouseX = float32(x)
		s.lastMouseY = float32(y)
	}
}
