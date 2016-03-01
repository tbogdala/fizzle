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

// Camera is an interface defining a common interface between different styles of cameras.
type Camera interface {
	GetViewMatrix() mgl.Mat4
	GetPosition() mgl.Vec3
}

// OrbitCamera makes a camera orbit at a given angle away with the distance controlled by a parameter.
// This poor ASCII art illustrates the relation of the target position, the angle between the
// camera and the up vector and where the camera ends up getting positioned.
//
//  Camera   up
//   \       |
//    \      |
//     \-ang-|
//      \    |
//       \   |
//        \  |
//       {target}
//
// After that's calculated, Camera->Up is used as a radius for a circle to then orbit the
// camera around the target based on the rotation parameter.
type OrbitCamera struct {
	// angle is the angle between the up vector and the camera
	vertAngle float32

	// distance controlls how far away the camera should be from the point of focus
	distance float32

	// target is the origin point of the camera
	target mgl.Vec3

	// rotation controlls the angle of the camera along the circle based on radius
	// being calculted from target->up with a vertAngle.
	rotation float32

	// position is the calculated position of the camera based on the target, the
	// angle and the distance desired.
	position mgl.Vec3
}

// NewOrbitCamera that looks at a target at a given vertAngle and at a given distance.
// See the OrbitCamera struct for ascii art on this.
func NewOrbitCamera(target mgl.Vec3, vertAngle float32, distance float32, rotation float32) *OrbitCamera {
	cam := new(OrbitCamera)
	cam.target = target
	cam.vertAngle = vertAngle
	cam.distance = distance
	cam.rotation = rotation
	cam.generatePosition()
	return cam
}

// generatePosition calculates the position based on the data members in the camera.
func (c *OrbitCamera) generatePosition() {
	cVert := float32(math.Cos(float64(c.vertAngle)))
	sVert := float32(math.Sin(float64(c.vertAngle)))
	height := cVert * c.distance
	c.position[1] = height + c.target[1]

	radius := sVert * c.distance
	cos := float32(math.Cos(float64(c.rotation)))
	sin := float32(math.Sin(float64(c.rotation)))

	c.position[0] = c.target[0] + radius*cos
	c.position[2] = c.target[2] + radius*sin
}

// GetPosition returns the eye position of the camera.
func (c *OrbitCamera) GetPosition() mgl.Vec3 {
	return c.position
}

// GetTarget returns the target position of the camera.
func (c *OrbitCamera) GetTarget() mgl.Vec3 {
	return c.target
}

// SetTarget changes the target position of the camera.
func (c *OrbitCamera) SetTarget(t mgl.Vec3) {
	c.target = t
	c.generatePosition()
}

// Rotate updates the rotation of the camera orbiting around the target.
func (c *OrbitCamera) Rotate(delta float32) {
	c.rotation += delta
	c.generatePosition()
}

// RotateVertical updates the vertical rotation of the camera orbiting
// around the target.
func (c *OrbitCamera) RotateVertical(delta float32) {
	newVal := c.vertAngle + delta

	// only update if we're not flipping the camera over the center axis.
	if newVal > math.Pi || newVal < 0.0 {
		return
	}

	c.vertAngle += delta
	c.generatePosition()
}

// AddDistance adds a value to the distance of the camera away from the target
// and then updates the internal data.
func (c *OrbitCamera) AddDistance(delta float32) {
	c.distance += delta
	c.generatePosition()
}

// GetDistance returns the distance of the camera away from the target.
func (c *OrbitCamera) GetDistance() float32 {
	return c.distance
}

// SetDistance sets the distance of the camera from the target and updates
// the internal data.
func (c *OrbitCamera) SetDistance(d float32) {
	// make sure it's not negative, for sanity purposes
	if d < 0 {
		return
	}

	c.distance = d
	c.generatePosition()
}

// GetViewMatrix returns a 4x4 matrix for the view rot/trans/scale.
func (c *OrbitCamera) GetViewMatrix() mgl.Mat4 {
	view := mgl.LookAtV(c.position, c.target, upVector)
	return view
}

// YawPitchCamera keeps track of the view rotation and position and provides
// utility methods to generate a view matrix.
// It provides a free-moving camera that is adjusted by yaw and pitch which,
// at default, looks down -Z with +Y as the up vector.
type YawPitchCamera struct {
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

// NewYawPitchCamera will create a new camera at a given position with no rotations applied.
func NewYawPitchCamera(eyePosition mgl.Vec3) *YawPitchCamera {
	const yaw float32 = 0.0

	cam := new(YawPitchCamera)
	cam.position = eyePosition
	cam.rotation = mgl.QuatRotate(yaw, mgl.Vec3{0.0, 1.0, 0.0})
	return cam
}

// GetViewMatrix returns a 4x4 matrix for the view rot/trans/scale.
func (c *YawPitchCamera) GetViewMatrix() mgl.Mat4 {
	view := c.rotation.Mat4()
	view = view.Mul4(mgl.Translate3D(-c.position[0], -c.position[1], -c.position[2]))
	return view
}

// GetPosition returns the eye position of the camera
func (c *YawPitchCamera) GetPosition() mgl.Vec3 {
	return c.position
}

// UpdatePosition adds delta values to the eye position vector.
func (c *YawPitchCamera) UpdatePosition(dX, dY, dZ float32) {
	c.position[0] += dX
	c.position[1] += dY
	c.position[2] += dZ
}

// SetPosition sets the position of the camera with an absolute coordinate.
func (c *YawPitchCamera) SetPosition(x, y, z float32) {
	c.position[0] = x
	c.position[1] = y
	c.position[2] = z
}

// SetYawAndPitch sets the yaw and pitch radians directly for the camera
func (c *YawPitchCamera) SetYawAndPitch(yaw, pitch float32) {
	c.yaw = yaw
	c.pitch = pitch
	c.generateRotation()
}

// UpdateYaw adds a delta to the camera yaw and regenerates the rotation quaternion.
func (c *YawPitchCamera) UpdateYaw(delta float32) {
	c.yaw += delta
	c.generateRotation()
}

// GetYaw returns the yaw of the camera in radians
func (c *YawPitchCamera) GetYaw() float32 {
	return c.yaw
}

// GetPitch returns the pitch of the camera in radians
func (c *YawPitchCamera) GetPitch() float32 {
	return c.pitch
}

// UpdatePitch adds a delta to the camera pitch and regenerates the rotation quaternion.
func (c *YawPitchCamera) UpdatePitch(delta float32) {
	c.pitch += delta
	c.generateRotation()
}

// LookAtDirect calculates a view rotation using the current Camera
// position so that it will look at the target coordinate.
// Uses standard up axis of {0,1,0}.
func (c *YawPitchCamera) LookAtDirect(target mgl.Vec3) {
	c.rotation = mgl.QuatLookAtV(c.position, target, upVector)
}

// LookAt adjusts the position of the camera based on the camera yaw/pitch
// and the target location passed in. It does automatically adjust
// the camera's internal rotation quaternion.
func (c *YawPitchCamera) LookAt(target mgl.Vec3, distance float32) {
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
func (c *YawPitchCamera) generateRotation() {
	camYawQ := mgl.QuatRotate(c.yaw, upVector)
	camPitchQ := mgl.QuatRotate(c.pitch, sideVector)
	c.rotation = camPitchQ.Mul(camYawQ)
}
