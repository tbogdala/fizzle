// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"

	fizzle "github.com/tbogdala/fizzle"
	forward "github.com/tbogdala/fizzle/renderer/forward"
	scene "github.com/tbogdala/fizzle/scene"
)

// TestScene is the main scene manager for the example.
type TestScene struct {
	// embed the basic scene manager
	*scene.BasicSceneManager

	shaders map[string]*fizzle.RenderShader
}

// NewTestScene creates a new test scene
func NewTestScene() *TestScene {
	ts := new(TestScene)
	ts.BasicSceneManager = scene.NewBasicSceneManager()
	ts.shaders = make(map[string]*fizzle.RenderShader)
	return ts
}

// SetupScene initializes the scene's assets and sets up the initial entities.
func (s *TestScene) SetupScene() error {
	// pull a reference to the render system
	system := s.BasicSceneManager.GetSystemByName(renderSystemName)
	renderSystem := system.(*RenderSystem)

	// setup the camera to look at the cube
	renderSystem.Camera = fizzle.NewOrbitCamera(mgl.Vec3{0, 0, 0}, math.Pi/2.0, 5.0, math.Pi/2.0)

	// put a light in there
	light := renderSystem.Renderer.NewDirectionalLight(mgl.Vec3{1.0, -0.5, -1.0})
	light.AmbientIntensity = 0.3
	light.DiffuseIntensity = 0.5
	light.SpecularIntensity = 0.3
	renderSystem.SetLight(0, light)

	// load the basic shader
	basicShader, err := forward.CreateBasicShader()
	if err != nil {
		return fmt.Errorf("Failed to compile and link the basic shader program! %v", err)
	}
	s.shaders["Basic"] = basicShader

	// setup a shaderd material to use for the objects
	redMaterial := fizzle.NewMaterial()
	redMaterial.Shader = basicShader
	redMaterial.DiffuseColor = mgl.Vec4{0.9, 0.05, 0.05, 1.0}
	redMaterial.Shininess = 4.8

	// create the test cube
	cubeEntity := NewVisibleEntity()
	cubeEntity.ID = s.GetNextID()
	cubeEntity.Renderable = fizzle.CreateCube(-1, -1, -1, 1, 1, 1)
	cubeEntity.Renderable.Material = redMaterial
	s.AddEntity(cubeEntity)

	return nil
}
