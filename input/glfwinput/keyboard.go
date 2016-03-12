// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package glfwinput

import glfw "github.com/go-gl/glfw/v3.1/glfw"

// KeyCallback is the type of the function that gets called for the key
// callback events.
type KeyCallback func()

// KeyboardModel is the way to bind keys to events.
type KeyboardModel struct {
	// KeyTriggerBindings are the functions to call when the given key is pressed
	KeyTriggerBindings map[glfw.Key]KeyCallback

	// KeyBindings are the functions to call when the given key is considered
	// 'pressed' when the KeyboardModel runs CheckKeyPresses.
	KeyBindings map[glfw.Key]KeyCallback

	// window is the GLFW window to poll for key input
	window *glfw.Window

	// keyCallback is the function that will get passed to GLFW as the
	// callback handler for key presses
	KeyCallback glfw.KeyCallback
}

// NewKeyboardModel returns a newly created keyboard model object
func NewKeyboardModel(w *glfw.Window) *KeyboardModel {
	kb := new(KeyboardModel)
	kb.KeyTriggerBindings = make(map[glfw.Key]KeyCallback)
	kb.KeyBindings = make(map[glfw.Key]KeyCallback)
	kb.window = w

	// use some default callbacks
	kb.KeyCallback = func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		// for now, we're only interested in press actions
		if action != glfw.Press {
			return
		}

		cb, okay := kb.KeyTriggerBindings[key]
		if okay && cb != nil {
			cb()
		}
	}

	return kb
}

// SetupCallbacks sets the callback handlers for the window.
func (kb *KeyboardModel) SetupCallbacks() {
	kb.window.SetKeyCallback(kb.KeyCallback)
}

// CheckKeyPresses runs through all of the KeyBindings and checks to see if that
// key is held down -- if it is, then the callback is invoked.
func (kb *KeyboardModel) CheckKeyPresses() {
	for key, cb := range kb.KeyBindings {
		if kb.window.GetKey(key) == glfw.Press && cb != nil {
			cb()
		}
	}
}

// Bind binds a key press event with a callback that will get called when
// CheckKeyPresses finds the key to be pressed.
func (kb *KeyboardModel) Bind(key glfw.Key, f KeyCallback) {
	kb.KeyBindings[key] = f
}

// BindTrigger binds a key event that gets called when the key is pressed once.
func (kb *KeyboardModel) BindTrigger(key glfw.Key, f KeyCallback) {
	kb.KeyTriggerBindings[key] = f
}
