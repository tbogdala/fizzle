// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package scene

import (
	mgl "github.com/go-gl/mathgl/mgl32"

	glider "github.com/tbogdala/glider"
)

// Entity is a common interface that represents any object in the scene.
type Entity interface {
	// GetID should return an identifier that no other Entity has.
	GetID() uint64

	// GetLocation returns the location of the entity in world space.
	GetLocation() mgl.Vec3

	// SetLocation sets the world space location of the entity.
	SetLocation(mgl.Vec3)
}

// BasicEntity represents any type of object in a scene from a moving character
// to a stationary crate.
type BasicEntity struct {
	ID              uint64
	Location        mgl.Vec3
	Orientation     mgl.Quat
	CoarseColliders []glider.Collider
}

// NewBasicEntity creates a new BasicEntity object with sane defaults and empty slices.
func NewBasicEntity() *BasicEntity {
	e := new(BasicEntity)
	e.Orientation = mgl.QuatIdent()
	e.CoarseColliders = []glider.Collider{}
	return e
}

// GetLocation returns the location of the entity in world space.
func (e *BasicEntity) GetLocation() mgl.Vec3 {
	return e.Location
}

// SetLocation sets the world space location of the entity.
func (e *BasicEntity) SetLocation(p mgl.Vec3) {
	e.Location = p
}

// GetID should return an identifier that no other Entity has.
func (e *BasicEntity) GetID() uint64 {
	return e.ID
}
