// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package scene

import (
	mgl "github.com/go-gl/mathgl/mgl32"

	component "github.com/tbogdala/fizzle/component"
	glider "github.com/tbogdala/glider"
)

// Entity is a common interface that represents any object in the scene.
type Entity interface {
	// GetID should return an identifier that no other Entity has.
	GetID() uint64

	// GetName should return a user readable name for the Entity, which does
	// not have to be unique.
	GetName() string

	// GetLocation returns the location of the entity in world space.
	GetLocation() mgl.Vec3

	// SetLocation sets the world space location of the entity.
	SetLocation(mgl.Vec3)

	// GetOrientation gets the local rotation of the entity as a quaternion.
	GetOrientation() mgl.Quat

	// SetOrientation sets the local rotation of the entity from a quaternion.
	SetOrientation(mgl.Quat)
}

// BasicEntity represents any type of object in a scene from a moving character
// to a stationary crate.
type BasicEntity struct {
	ID              uint64
	Name            string
	location        mgl.Vec3
	orientation     mgl.Quat // local rotation of th entity
	CoarseColliders []glider.Collider
}

// NewBasicEntity creates a new BasicEntity object with sane defaults and empty slices.
func NewBasicEntity() *BasicEntity {
	e := new(BasicEntity)
	e.orientation = mgl.QuatIdent()
	e.CoarseColliders = []glider.Collider{}
	return e
}

// GetLocation returns the location of the entity in world space.
func (e *BasicEntity) GetLocation() mgl.Vec3 {
	return e.location
}

// SetLocation sets the world space location of the entity.
func (e *BasicEntity) SetLocation(p mgl.Vec3) {
	e.location = p
}

// GetOrientation gets the local rotation of the entity as a quaternion.
func (e *BasicEntity) GetOrientation() mgl.Quat {
	return e.orientation
}

// SetOrientation sets the local rotation of the entity from a quaternion.
func (e *BasicEntity) SetOrientation(q mgl.Quat) {
	e.orientation = q
}

// GetID should return an identifier that no other Entity has.
func (e *BasicEntity) GetID() uint64 {
	return e.ID
}

// GetName should return a user readable name for the Entity, which does
// not have to be unique.
func (e *BasicEntity) GetName() string {
	return e.Name
}

// CreateCollidersFromComponent will create the coarse collision objects
// for the basic entity based on the component definition.
func (e *BasicEntity) CreateCollidersFromComponent(c *component.Component) {
	// sanity check to see if we have collisions to create
	if c == nil || c.Collisions == nil || len(c.Collisions) <= 0 {
		return
	}

	for _, ref := range c.Collisions {
		switch ref.Type {
		case component.ColliderTypeAABB:
			aabb := glider.NewAABBox()
			aabb.Min = ref.Min
			aabb.Max = ref.Max
			e.CoarseColliders = append(e.CoarseColliders, aabb)
		case component.ColliderTypeSphere:
			sphere := glider.NewSphere()
			sphere.Center = ref.Offset
			sphere.Radius = ref.Radius
			e.CoarseColliders = append(e.CoarseColliders, sphere)
		}
	}
}
