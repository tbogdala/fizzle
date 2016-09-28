// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	glfw "github.com/go-gl/glfw/v3.1/glfw"

	input "github.com/tbogdala/fizzle/input/glfwinput"
	"github.com/tbogdala/fizzle/scene"
)

const (
	inputSystemPriority = -100.0
	inputSystemName     = "InputSystem"
)

// InputSystem implements the fizzle/scene/System interface and handles the
// player input.
type InputSystem struct {
	kbModel *input.KeyboardModel

	mainWindow *glfw.Window
}

// NewInputSystem creates a new InputSystem object
func NewInputSystem() *InputSystem {
	system := new(InputSystem)
	return system
}

// Initialize sets up the input models for the scene.
func (s *InputSystem) Initialize(w *glfw.Window) {
	s.mainWindow = w

	// set the callback functions for key input
	s.kbModel = input.NewKeyboardModel(s.mainWindow)
	s.kbModel.BindTrigger(glfw.KeyEscape, setShouldClose)
	s.kbModel.SetupCallbacks()

}

// Update should get called to run updates for the system every frame
// by the owning Manager object.
func (s *InputSystem) Update(frameDelta float32) {
	// advise GLFW to poll for input. without this the window appears to hang.
	glfw.PollEvents()

	// handle any keyboard input
	s.kbModel.CheckKeyPresses()
}

// OnAddEntity should get called by the scene Manager each time a new entity
// has been added to the scene.
func (s *InputSystem) OnAddEntity(newEntity scene.Entity) {
	// NOP
}

// OnRemoveEntity should get called by the scene Manager each time an entity
// has been removed from the scene.
func (s *InputSystem) OnRemoveEntity(oldEntity scene.Entity) {
	// NOP
}

// GetRequestedPriority returns the requested priority level for the System
// which may be of significance to a Manager if they want to order Update() calls.
func (s *InputSystem) GetRequestedPriority() float32 {
	return inputSystemPriority
}

// GetName returns the name of the system that can be used to identify
// the System within Manager.
func (s *InputSystem) GetName() string {
	return inputSystemName
}

// setShouldClose should be called to close the window and kill the app.
func setShouldClose() {
	system := sceneMan.GetSystemByName(inputSystemName)
	inputSystem := system.(*InputSystem)
	inputSystem.mainWindow.SetShouldClose(true)
}
