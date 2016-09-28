// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	input "github.com/tbogdala/fizzle/input/glfwinput"
)

// GLFW event handling must run on the main OS thread.
func init() {
	runtime.LockOSThread()
}

const (
	windowWidth  = 1280
	windowHeight = 720
)

var (
	sceneMan *TestScene
	kbModel  *input.KeyboardModel
)

func main() {
	var err error

	// create the render system and initialize it
	renderSystem := NewRenderSystem()
	err = renderSystem.Initialize("Test Scene", windowWidth, windowHeight)
	if err != nil {
		fmt.Printf("Failed to initialize the render system! %v", err)
		os.Exit(1)
	}

	// setup the input system
	inputSystem := NewInputSystem()
	inputSystem.Initialize(renderSystem.MainWindow)

	// create a scene manager
	sceneMan = NewTestScene()
	sceneMan.AddSystem(renderSystem)
	sceneMan.AddSystem(inputSystem)

	// setup the components of the scene
	err = sceneMan.SetupScene()
	if err != nil {
		fmt.Printf("Failed to initialize the test scene! %v", err)
		os.Exit(1)
	}

	// loop until something told the mainWindow that it should close
	lastFrame := time.Now()
	for !renderSystem.MainWindow.ShouldClose() {
		// calculate the difference in time to control rotation speed
		thisFrame := time.Now()
		frameDelta := float32(thisFrame.Sub(lastFrame).Seconds())

		// update the scene
		sceneMan.Update(frameDelta)

		// update our last frame time
		lastFrame = thisFrame
	}
}
