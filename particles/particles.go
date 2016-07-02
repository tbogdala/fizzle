// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package particles

import (
	"math"
	"math/rand"

	mgl "github.com/go-gl/mathgl/mgl32"
	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	renderer "github.com/tbogdala/fizzle/renderer"
)

var (
	// VertShader330 is the GLSL vertex shader program for the basic particle emitter.
	VertShader330 = `#version 330
  uniform mat4 MVP;
  in vec3 POSITION;
  in vec4 COLOR;
  in float SIZE;

  out vec4 vs_color;

  void main()
  {
    vs_color = COLOR;

    gl_PointSize = SIZE;
    gl_Position = MVP * vec4(POSITION, 1.0);
  }`

	// FragShader330 is the GLSL fragment shader program for the asic bparticle emitter.
	FragShader330 = `#version 330
  uniform sampler2D TEX;
  in vec4 vs_color;

  out vec4 frag_color;

  void main()
  {
	frag_color = vs_color * texture(TEX, gl_PointCoord.st);
  }`
)

// System is a particle system master collection that keeps track of all of the
// particle emitters and updates them accordingly.
type System struct {
	Emitters []*Emitter
	Origin   mgl.Vec3
	gfx      graphics.GraphicsProvider
	runtime  float64
}

// ParticleSpawner is a type of interface for objects that are able to spawn
// particles for a particle emitter.
type ParticleSpawner interface {
	NewParticle() Particle
	DrawSpawnVolume(r renderer.Renderer, shader *fizzle.RenderShader, projection mgl.Mat4, view mgl.Mat4, camera fizzle.Camera)
	GetLocation() mgl.Vec3
}

// Emitter is a particle emmiter object that will keep track of all of the particles
// created by the emitter and update them accordingly.
type Emitter struct {
	Owner      *System
	Particles  []Particle
	Billboard  graphics.Texture
	Shader     graphics.Program
	Properties EmitterProperties
	Spawner    ParticleSpawner

	vao            uint32
	comboVBO       graphics.Buffer
	comboBuffer    []float32
	timeSinceSpawn float64
	rng            *rand.Rand
}

// EmitterProperties describes the behavior of an Emitter object and is it's own
// type to facilitate sharing of parameter defaults and serialization.
type EmitterProperties struct {
	MaxParticles uint
	SpawnRate    uint     // particles per second
	Velocity     mgl.Vec3 // should be normalized
	Speed        float32
	Acceleration mgl.Vec3
	TTL          float64  // in seconds
	Origin       mgl.Vec3 // relative to Emitter.Owner.Origin
	Rotation     mgl.Quat
	Color        mgl.Vec4
	Size         float32
}

// Particle is an individual particle in an Emitter.
type Particle struct {
	Size         float32
	Color        mgl.Vec4
	Location     mgl.Vec3
	Velocity     mgl.Vec3 // should be normalized
	Speed        float32
	Acceleration mgl.Vec3
	EndTime      float64
	StartTime    float64
}

// NewSystem creates a new particle system.
func NewSystem(gfx graphics.GraphicsProvider) *System {
	s := new(System)
	s.gfx = gfx
	return s
}

// GetTransform returns the transform matrix for the system as a whole.
func (s *System) GetTransform() mgl.Mat4 {
	return mgl.Translate3D(s.Origin[0], s.Origin[1], s.Origin[2])
}

// NewEmitter is it creates a new particle emitter and keeps track of it for
// updating. An optional set of emitter properties can be specified.
func (s *System) NewEmitter(optProps *EmitterProperties) *Emitter {
	e := new(Emitter)
	e.Owner = s

	// setup the rng for the emitter with a default seed of 1
	e.rng = rand.New(rand.NewSource(1))

	// for now, create a default spawner
	e.Spawner = NewConeSpawner(e, 0.5, 1.0, 2.0)
	//e.Spawner = NewCubeSpawner(e, mgl.Vec3{-1, 0, -1}, mgl.Vec3{1, 0.01, 1})

	// set the emitter properties if specified
	if optProps != nil {
		e.Properties = *optProps
	} else {
		// plug in some defaults
		e.Properties.Size = 32.0
		e.Properties.Color = mgl.Vec4{1, 1, 1, 1}
		e.Properties.Speed = 1.0
		e.Properties.Velocity = mgl.Vec3{0, 1, 0}
		e.Properties.Rotation = mgl.QuatIdent()
	}

	// construct the objects needed for rendering
	e.vao = s.gfx.GenVertexArray()
	e.comboVBO = s.gfx.GenBuffer()

	// keep track of it
	s.Emitters = append(s.Emitters, e)

	return e
}

// Update will update all of the emitters currently tracked by the system.
func (s *System) Update(frameDelta float64) {
	s.runtime += frameDelta
	for _, emitter := range s.Emitters {
		emitter.Update(frameDelta)
	}
}

// Draw renders all particle emitters.
func (s *System) Draw(projection mgl.Mat4, view mgl.Mat4) {
	for _, emitter := range s.Emitters {
		emitter.Draw(projection, view)
	}
}

// GetLocation returns the emitter location in world space.
func (e *Emitter) GetLocation() mgl.Vec3 {
	return e.Owner.Origin.Add(e.Properties.Origin)
}

// Update will update all of the particles for the emitter and then
// update the graphics buffers.
func (e *Emitter) Update(frameDelta float64) {
	// filter out all of the dead particles
	stillAlive := e.Particles[:0]
	for _, particle := range e.Particles {
		if e.Owner.runtime <= particle.EndTime {
			stillAlive = append(stillAlive, particle)
		}
	}
	e.Particles = stillAlive

	// how many particle to spawn?
	var spawnInterval = float64(1.0)
	e.timeSinceSpawn += frameDelta
	if e.Properties.SpawnRate != 0.0 {
		spawnInterval = 1.0 / float64(e.Properties.SpawnRate)
	}
	spawnCount := math.Floor(e.timeSinceSpawn / spawnInterval)

	// update the timers
	e.timeSinceSpawn -= spawnCount * spawnInterval

	// update the particles
	for i, particle := range e.Particles {
		dV := particle.Velocity.Mul(float32(frameDelta) * particle.Speed)
		//dA := particle.Acceleration.Mul(float32(frameDelta))
		e.Particles[i].Location = particle.Location.Add(dV)
		//e.Particles[i].Velocity = particle.Velocity.Add(dA)
	}

	// add the particles
	var newParticle Particle
	for spawnCount > 0 && len(e.Particles) < int(e.Properties.MaxParticles) {
		newParticle = e.Spawner.NewParticle()
		e.Particles = append(e.Particles, newParticle)
		spawnCount--
	}
}

const (
	floatSize = 4
)

func (e *Emitter) renderToVBO() {
	buffer := e.comboBuffer[:0]

	for _, p := range e.Particles {
		// 3f = vertex
		buffer = append(buffer, p.Location[0])
		buffer = append(buffer, p.Location[1])
		buffer = append(buffer, p.Location[2])

		// 4f = color
		buffer = append(buffer, p.Color[0])
		buffer = append(buffer, p.Color[1])
		buffer = append(buffer, p.Color[2])
		buffer = append(buffer, p.Color[3])

		// 1f = size
		buffer = append(buffer, p.Size)
	}

	// we didn't buffer anything
	if len(buffer) <= 0 {
		return
	}

	// buffer the data
	e.Owner.gfx.BindBuffer(graphics.ARRAY_BUFFER, e.comboVBO)
	e.Owner.gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(buffer), e.Owner.gfx.Ptr(&buffer[0]), graphics.STREAM_DRAW)
}

// Draw renders the particle emitter.
func (e *Emitter) Draw(projection mgl.Mat4, view mgl.Mat4) {
	if e.Particles == nil || len(e.Particles) <= 0 {
		return
	}

	gfx := e.Owner.gfx
	gfx.BindVertexArray(e.vao)

	// update the graphics buffers
	e.renderToVBO()

	gfx.UseProgram(e.Shader)

	parentTransform := e.Owner.GetTransform()
	modelTransform := mgl.Translate3D(e.Properties.Origin[0], e.Properties.Origin[1], e.Properties.Origin[2])
	model := parentTransform.Mul4(modelTransform)
	mvp := projection.Mul4(view).Mul4(model)

	// bind the uniforms and attributes
	mvpMatrix := gfx.GetUniformLocation(e.Shader, "MVP")
	if mvpMatrix >= 0 {
		gfx.UniformMatrix4fv(mvpMatrix, 1, false, mvp)
	}

	shaderTex0 := gfx.GetUniformLocation(e.Shader, "TEX")
	if shaderTex0 >= 0 {
		gfx.ActiveTexture(graphics.TEXTURE0)
		gfx.BindTexture(graphics.TEXTURE_2D, e.Billboard)
		gfx.Uniform1i(shaderTex0, 0)
	}

	const posOffset = 0
	const colorOffset = floatSize * 3
	const sizeOffset = floatSize * 7
	const Stride = floatSize * (3 + 4 + 1) // vert / color / size

	shaderPosition := gfx.GetAttribLocation(e.Shader, "POSITION")
	gfx.BindBuffer(graphics.ARRAY_BUFFER, e.comboVBO)
	gfx.EnableVertexAttribArray(uint32(shaderPosition))
	gfx.VertexAttribPointer(uint32(shaderPosition), 3, graphics.FLOAT, false, Stride, gfx.PtrOffset(posOffset))

	shaderColor := gfx.GetAttribLocation(e.Shader, "COLOR")
	gfx.EnableVertexAttribArray(uint32(shaderColor))
	gfx.VertexAttribPointer(uint32(shaderColor), 4, graphics.FLOAT, false, Stride, gfx.PtrOffset(colorOffset))

	shaderSize := gfx.GetAttribLocation(e.Shader, "SIZE")
	gfx.EnableVertexAttribArray(uint32(shaderSize))
	gfx.VertexAttribPointer(uint32(shaderSize), 1, graphics.FLOAT, false, Stride, gfx.PtrOffset(sizeOffset))

	gfx.DrawArrays(graphics.POINTS, 0, int32(len(e.Particles)))

	gfx.BindVertexArray(0)
}
