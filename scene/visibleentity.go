// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package scene

import (
	mgl "github.com/go-gl/mathgl/mgl32"

	fizzle "github.com/tbogdala/fizzle"
)

// RenderableEntity is an interface for entities that have a renderable to draw.
type RenderableEntity interface {
	GetRenderable() *fizzle.Renderable
}

// VisibleEntity is a scene entity that can be rendered to screen.
type VisibleEntity struct {
	*BasicEntity

	// Renderable is the model to draw for the entity if one should be
	// drawn -- it is valid to have a nil Renderable here.
	Renderable *fizzle.Renderable
}

// NewVisibleEntity returns a new visible entity object.
func NewVisibleEntity() *VisibleEntity {
	ve := new(VisibleEntity)
	ve.BasicEntity = NewBasicEntity()
	return ve
}

// GetRenderable returns the renderable for the entity.
func (e *VisibleEntity) GetRenderable() *fizzle.Renderable {
	return e.Renderable
}

// SetLocation is a helper function to set the location of the entity as well
// as any renderable.
func (e *VisibleEntity) SetLocation(pos mgl.Vec3) {
	e.BasicEntity.SetLocation(pos)
	if e.Renderable != nil {
		e.Renderable.Location = pos
	}
	for _, c := range e.CoarseColliders {
		c.SetOffset(&pos)
	}
}

// SetOrientation is a helper function to set the orientation of the entity as
// well as any renderable.
func (e *VisibleEntity) SetOrientation(q mgl.Quat) {
	e.BasicEntity.SetOrientation(q)
	if e.Renderable != nil {
		e.Renderable.LocalRotation = q
	}
	// TODO: upate collision objects too at some point (when using them)
}
