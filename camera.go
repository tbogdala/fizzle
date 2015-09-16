// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
)

var (
	upVector   = mgl.Vec3{0.0, 1.0, 0.0}
	sideVector = mgl.Vec3{1.0, 0.0, 0.0}
)

// Camera keeps track of the view rotation and position and provides
// utility methods to generate a view matrix.
//
// The default mode of the camera is to provide a free-moving camera
// that is adjusted by yaw and pitch.
type Camera struct {
	// store the pitch and yaw as angles to convert into
	// quaternions on change. this allows for movement directed
	// by yaw, but not affected by pitch.
	// NOTE: specified in radians.
	yaw   float32
	pitch float32

	// derived from camYaw and camPitch and is what is used for the camera
	rotation mgl.Quat
	position mgl.Vec3
}

// NewCamera will create a new camera at a given position with no rotations applied.
func NewCamera(eyePosition mgl.Vec3) *Camera {
	const yaw float32 = 0.0

	cam := new(Camera)
	cam.position = eyePosition
	cam.rotation = mgl.QuatRotate(yaw, mgl.Vec3{0.0, 1.0, 0.0})
	return cam
}

// GetViewMatrix returns a 4x4 matrix for the view rot/trans/scale.
func (c *Camera) GetViewMatrix() mgl.Mat4 {
	view := c.rotation.Mat4()
	view = view.Mul4(mgl.Translate3D(-c.position[0], -c.position[1], -c.position[2]))
	return view
}

// GetPosition returns the eye position of the camera
func (c *Camera) GetPosition() mgl.Vec3 {
	return c.position
}

// UpdatePosition adds delta values to the eye position vector.
func (c *Camera) UpdatePosition(dX, dY, dZ float32) {
	c.position[0] += dX
	c.position[1] += dY
	c.position[2] += dZ
}

// SetYawAndPitch sets the yaw and pitch radians directly for the camera
func (c *Camera) SetYawAndPitch(yaw, pitch float32) {
	c.yaw = yaw
	c.pitch = pitch
	c.generateRotation()
}

// UpdateYaw adds a delta to the camera yaw and regenerates the rotation quaternion.
func (c *Camera) UpdateYaw(delta float32) {
	c.yaw += delta
	c.generateRotation()
}

// GetYaw returns the yaw of the camera in radians
func (c *Camera) GetYaw() float32 {
	return c.yaw
}

// GetPitch returns the pitch of the camera in radians
func (c *Camera) GetPitch() float32 {
	return c.pitch
}

// UpdatePitch adds a delta to the camera pitch and regenerates the rotation quaternion.
func (c *Camera) UpdatePitch(delta float32) {
	c.pitch += delta
	c.generateRotation()
}

// LookAtDirect calculates a view rotation using the current Camera
// position so that it will look at the target coordinate.
// Uses standard up axis of {0,1,0}.
var (
	defaultLookAtUp = mgl.Vec3{0.0, 1.0, 0.0}
)

func (c *Camera) LookAtDirect(target mgl.Vec3) {
	c.rotation = mgl.QuatLookAtV(c.position, target, defaultLookAtUp)
}

// LookAt adjusts the position of the camera based on the camera yaw/pitch
// and the target location passed in. It does automatically adjust
// the camera's internal rotation quaternion.
func (c *Camera) LookAt(target mgl.Vec3, distance float32) {
	// use trig to get the camera position scaled by distance
	rotatedX := float32(math.Cos(float64(c.yaw))) * distance
	rotatedZ := float32(math.Sin(float64(c.yaw))) * distance

	// set the camera's location
	c.position[0] = target[0] + rotatedX
	c.position[1] = target[1] + distance
	c.position[2] = target[2] + rotatedZ

	correctedYaw := float32(math.Atan2(float64(c.position[0]-target[0]), float64(c.position[2]-target[2])))

	// update the rotation quaternion
	camYawQ := mgl.QuatRotate(-1*correctedYaw, upVector)
	camPitchQ := mgl.QuatRotate(c.pitch, sideVector)
	c.rotation = camPitchQ.Mul(camYawQ)
}

// generateRotation recalculates the rotation quaternion based on the pitch and yaw radians.
func (c *Camera) generateRotation() {
	camYawQ := mgl.QuatRotate(c.yaw, upVector)
	camPitchQ := mgl.QuatRotate(c.pitch, sideVector)
	c.rotation = camPitchQ.Mul(camYawQ)
}
