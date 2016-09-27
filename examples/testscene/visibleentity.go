// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	fizzle "github.com/tbogdala/fizzle"
	scene "github.com/tbogdala/fizzle/scene"
)

// RenderableEntity is an interface for entities that have a renderable to draw.
type RenderableEntity interface {
	GetRenderable() *fizzle.Renderable
}

// VisibleEntity is a scene entity that can be rendered to screen.
type VisibleEntity struct {
	*scene.BasicEntity

	Renderable *fizzle.Renderable
}

// NewVisibleEntity returns a new visible entity object.
func NewVisibleEntity() *VisibleEntity {
	ve := new(VisibleEntity)
	ve.BasicEntity = scene.NewBasicEntity()
	return ve
}

// GetRenderable returns the renderable for the entity.
func (e *VisibleEntity) GetRenderable() *fizzle.Renderable {
	return e.Renderable
}
