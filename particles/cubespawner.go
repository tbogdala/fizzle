// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package particles

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	fizzle "github.com/tbogdala/fizzle"
	renderer "github.com/tbogdala/fizzle/renderer"
)

// CubeSpawner is a particle spawner that creates particles within the
// volume of a cube as specified by the settings in the struct.
type CubeSpawner struct {
	BottomLeft mgl.Vec3
	TopRight   mgl.Vec3
	Owner      *Emitter

	volumeRenderable *fizzle.Renderable
}

// NewCubeSpawner creates a new cube shaped particle spawner.
func NewCubeSpawner(owner *Emitter, bl, tr mgl.Vec3) *CubeSpawner {
	cube := new(CubeSpawner)
	cube.BottomLeft = bl
	cube.TopRight = tr
	cube.Owner = owner
	return cube
}

// GetName returns a user friendly name for the spawner
func (cube *CubeSpawner) GetName() string {
	return "Cube Spawner"
}

// SetOwner sets the owning emitter for the spawner
func (cube *CubeSpawner) SetOwner(e *Emitter) {
	cube.Owner = e
}

// GetLocation returns the location in world space for the spawner.
func (cube *CubeSpawner) GetLocation() mgl.Vec3 {
	return cube.Owner.GetLocation()
}

// NewParticle creates a new particle that fits within the volume of a cone section.
func (cube *CubeSpawner) NewParticle() (p Particle) {
	// get the standard properties from the emitter itself
	p.StartTime = cube.Owner.Owner.runtime
	p.Size = cube.Owner.Properties.Size
	p.Speed = cube.Owner.Properties.Speed
	p.Color = cube.Owner.Properties.Color
	p.Acceleration = cube.Owner.Properties.Acceleration
	p.EndTime = cube.Owner.Properties.TTL + p.StartTime

	// get a random point within the bottom circle
	w := cube.TopRight[0] - cube.BottomLeft[0]
	h := cube.TopRight[1] - cube.BottomLeft[1]
	d := cube.TopRight[2] - cube.BottomLeft[2]

	x := cube.BottomLeft[0] + cube.Owner.rng.Float32()*w
	y := cube.BottomLeft[1] + cube.Owner.rng.Float32()*h
	z := cube.BottomLeft[2] + cube.Owner.rng.Float32()*d

	p.Location[0] = x
	p.Location[1] = y
	p.Location[2] = z
	p.Location = cube.Owner.Properties.Rotation.Rotate(p.Location)

	p.Velocity = cube.Owner.Properties.Velocity.Normalize()
	p.Velocity = cube.Owner.Properties.Rotation.Rotate(p.Velocity)

	return p
}

// CreateRenderable creates a cached renderable for the spawner that represents
// the spawning volume for particles.
func (cube *CubeSpawner) CreateRenderable() *fizzle.Renderable {
	cube.volumeRenderable = fizzle.CreateWireframeCube(cube.BottomLeft[0], cube.BottomLeft[1], cube.BottomLeft[2],
		cube.TopRight[0], cube.TopRight[1], cube.TopRight[2])
	return cube.volumeRenderable
}

// DrawSpawnVolume renders a visual representation of the particle spawning volume.
func (cube *CubeSpawner) DrawSpawnVolume(r renderer.Renderer, shader *fizzle.RenderShader, projection mgl.Mat4, view mgl.Mat4, camera fizzle.Camera) {
	if cube.volumeRenderable == nil {
		cube.CreateRenderable()
	}

	// sync the position
	cube.volumeRenderable.Location = cube.Owner.Properties.Origin
	cube.volumeRenderable.LocalRotation = cube.Owner.Properties.Rotation

	r.DrawLines(cube.volumeRenderable, shader, nil, projection, view, camera)
}
