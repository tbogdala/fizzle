// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package glfwinput

import (
	"math"

	"github.com/go-gl/glfw/v3.1/glfw"
)

// JoystickButtonCallback is the type of the function that gets called for the
// joystick button callback events.
type JoystickButtonCallback func()

// JoystickAxisCallback is the type of the function that gets called for the
// joystick axis callback events.
type JoystickAxisCallback func(axis float32)

// JoystickAxisBinding determines what values on an axis are mapped to
// what callback.
type JoystickAxisBinding struct {
	// ID is the axis id to check for input
	ID int

	// Min is the minimum value on the axis to trigger the callback
	Min float32

	// Max is the maximum value on the axis to trigger the callback
	Max float32

	// NegativeMapping indicates whether or not the values from
	// the axis should be run through the Abs function before
	// getting passed to the callback.
	NegativeMapping bool

	// Callback is the function to call if the axis id has input
	// withing the [Min..Max] range. Delta will be scaled to match
	// the strength of input.
	Callback JoystickAxisCallback
}

// JoystickModel is the way to bind joystick buttons and axes
// to movement.
type JoystickModel struct {
	// ButtonBindings maps a button id to a callback function
	ButtonBindings map[int]JoystickButtonCallback

	// AxisBindings maps an axis id to an axis binding structure
	// that determines what input ranges trigger the callback.
	//
	// Note: this is a slice because there can be multiple bindings
	// for a given axis ID.
	AxisBindings []JoystickAxisBinding

	// window is the GLFW window to poll for joystick input
	window *glfw.Window

	// joystickID is the joystick to check in GLFW for this model
	joystickID glfw.Joystick
}

// NewJoystickModel returns a newly created joystic model object
func NewJoystickModel(w *glfw.Window, j glfw.Joystick) *JoystickModel {
	js := new(JoystickModel)
	js.ButtonBindings = make(map[int]JoystickButtonCallback)
	js.AxisBindings = make([]JoystickAxisBinding, 0)
	js.window = w
	js.joystickID = j
	return js
}

// NewJoystickAxisBinding creates a new joystick binding with the given settings.
func NewJoystickAxisBinding(id int, min, max float32, negativeMapping bool, cb JoystickAxisCallback) (jb JoystickAxisBinding) {
	jb.ID = id
	jb.Min = min
	jb.Max = max
	jb.NegativeMapping = negativeMapping
	jb.Callback = cb
	return jb
}

// IsJoystickPresent returns true if the joystick is detected.
func IsJoystickPresent(j glfw.Joystick) bool {
	return glfw.JoystickPresent(j)
}

// IsActive returns true if the bound joystick id is present.
func (jm *JoystickModel) IsActive() bool {
	return IsJoystickPresent(jm.joystickID)
}

// BindButton binds an event handler for a given button id on a joystick.
func (jm *JoystickModel) BindButton(button int, f JoystickButtonCallback) {
	jm.ButtonBindings[button] = f
}

// BindAxis binds an axis mapping for an axis id over a range of values on
// a joystick.
func (jm *JoystickModel) BindAxis(binding JoystickAxisBinding) {
	jm.AxisBindings = append(jm.AxisBindings, binding)
}

// CheckInput checks the joystick input against the bindings and invokes
// any matched callbacks.
func (jm *JoystickModel) CheckInput() {
	// if the joystick is still connected, then we do the joystick polling
	if !glfw.JoystickPresent(jm.joystickID) {
		return
	}

	// poll the joystick for the current state
	buttons := glfw.GetJoystickButtons(jm.joystickID)
	axes := glfw.GetJoystickAxes(jm.joystickID)

	// process the buttons
	for buttonID, cb := range jm.ButtonBindings {
		if buttons[buttonID] > 0 && cb != nil {
			cb()
		}
	}

	// process the axis values
	for _, mapping := range jm.AxisBindings {
		// if there's no callback, then there's no point in trying
		if mapping.Callback == nil {
			continue
		}

		v := axes[mapping.ID]
		if v >= mapping.Min && v <= mapping.Max {
			scale := mapping.Max - mapping.Min
			if mapping.NegativeMapping {
				// use the Max value here since NegativeMapping implies a negative ranage for the mapping
				v = (float32(math.Abs(float64(v))) - float32(math.Abs(float64(mapping.Max)))) / scale
			} else {
				v = (v - mapping.Min) / scale
			}
			mapping.Callback(v)
		}
	}
}
