// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	gl "github.com/go-gl/gl/v3.3-core/gl"
	"github.com/tbogdala/groggy"
)

func DegreesToRadians(x float64) float64 {
	return x * 0.017453292519943296
}

func RadiansToDegrees(x float64) float64 {
	return x * 57.2957795130823229
}

// DebugCheckForError repeatedly calls OpenGL's GetError() and prints out the
// error codes with a customized message string.
func DebugCheckForError(msg string) {
	err := gl.GetError()
	for err != gl.NO_ERROR {
		if len(msg) > 0 {
			var errTypeStr string
			switch err {
			case gl.INVALID_ENUM:
				errTypeStr = "INVALID_ENUM"
			case gl.INVALID_VALUE:
				errTypeStr = "INVALID_VALUE"
			case gl.INVALID_OPERATION:
				errTypeStr = "INVALID_OPERATION"
			case gl.OUT_OF_MEMORY:
				errTypeStr = "OUT_OF_MEMORY"
			case gl.STACK_OVERFLOW:
				errTypeStr = "STACK_OVERFLOW"
			case gl.STACK_UNDERFLOW:
				errTypeStr = "STACK_UNDERFLOW"
			default:
				errTypeStr = "Undefined Error"
			}
			groggy.Logsf("DEBUG", "OpenGL error %d(0x%x) detected (%s): %s", int(err), int(err), msg, errTypeStr)
		}
		err = gl.GetError()
	}
}
