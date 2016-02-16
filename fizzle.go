// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

/*

Fizzle is a library to make rendering graphics via OpenGL easier.

*/

package fizzle

import (
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/groggy"
)

// gfx is the currently initialized GraphicsProvider. It is accessed
// externally through the GetGraphics() and SetGraphics() functions.
var gfx graphics.GraphicsProvider

// GetGraphics returns the currently initialized GraphicsProvider
// if one has been indeed initizlied.
func GetGraphics() graphics.GraphicsProvider {
	return gfx
}

// SetGraphics sets the GraphicsProvider to use for all operations
// in the fizzle package.
func SetGraphics(g graphics.GraphicsProvider) {
	gfx = g
}

// DegreesToRadians converts degrees to radians
func DegreesToRadians(x float64) float64 {
	return x * 0.017453292519943296
}

// RadiansToDegrees converts radians to degrees
func RadiansToDegrees(x float64) float64 {
	return x * 57.2957795130823229
}

// DebugCheckForError repeatedly calls OpenGL's GetError() and prints out the
// error codes with a customized message string.
func DebugCheckForError(msg string) {
	err := gfx.GetError()
	for err != graphics.NO_ERROR {
		if len(msg) > 0 {
			var errTypeStr string
			switch err {
			case graphics.INVALID_ENUM:
				errTypeStr = "INVALID_ENUM"
			case graphics.INVALID_VALUE:
				errTypeStr = "INVALID_VALUE"
			case graphics.INVALID_OPERATION:
				errTypeStr = "INVALID_OPERATION"
			case graphics.OUT_OF_MEMORY:
				errTypeStr = "OUT_OF_MEMORY"
			default:
				errTypeStr = "Undefined Error"
			}
			groggy.Logsf("DEBUG", "OpenGL error %d(0x%x) detected (%s): %s", int(err), int(err), msg, errTypeStr)
		}
		err = gfx.GetError()
	}
}
