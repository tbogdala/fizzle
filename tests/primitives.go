package main

import (
	"flag"
	"reflect"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
)

var (
	timeout  = flag.Duration("t", 5*time.Second, "timeout for test")
	prevTest time.Time

	tests   = reflect.ValueOf(Tests{})
	testNum int
)

func main() {
	newWindow()
	renderLoop()
}

func startTests() {
	// if time.Now().After(prevTest) {
	for i := 0; i < tests.NumMethod(); i++ {
		tests.Method(i).Call([]reflect.Value{})
	}
	// if testNum >= tests.NumMethod() {
	// 	os.Exit(0)
	// }

	// clear renderer objects
	// objects = make(map[string]*fizzle.Renderable)
	// shapes = make(map[string]*fizzle.Renderable)

	// run next test

	// set timeout
	// prevTest = time.Now().Add(*timeout)
	// testNum++
	// }

	time.Sleep(*timeout)
}

type Tests struct{}

//Cube test
func (Tests) Cube() {
	cube := fizzle.CreateCube(-1, -1, -1, 1, 1, 1)
	setMaterial(cube)

	objects["cube"] = cube
	cube.Location = mgl32.Vec3{0, 4, 0}
}

func setMaterial(o *fizzle.Renderable) {
	o.Core.Shader = shader
	o.Core.DiffuseColor = mgl32.Vec4{0.5, 0.5, 0.5, 1.0}
	o.Core.SpecularColor = mgl32.Vec4{0.2, 0.2, 0.2, 1.0}
	o.Core.Shininess = 4.8
}

//Circles test shape XY circle
func (Tests) Circles() {
	shapes["circlexy"] = fizzle.CreateWireframeCircle(-6, 0, 0, 1, 32, fizzle.X|fizzle.Y)
	shapes["circlexz"] = fizzle.CreateWireframeCircle(-3, 0, 0, 1, 32, fizzle.X|fizzle.Z)
	shapes["circleyz"] = fizzle.CreateWireframeCircle(0, 0, 0, 1, 32, fizzle.Y|fizzle.Z)
	shapes["circlexyz"] = fizzle.CreateWireframeCircle(3, 0, 0, 1, 32, fizzle.X|fizzle.Y|fizzle.Z)
}

//Cube test
func (Tests) Planes() {
	//normal plane
	plane0 := fizzle.CreatePlaneV(mgl32.Vec3{-1, -1, 0}, mgl32.Vec3{1, 1, 0})
	setMaterial(plane0)
	objects["plane0"] = plane0
	plane0.Location = mgl32.Vec3{-3, -4, 0}

	//backward oblique plane
	plane1 := fizzle.CreatePlaneV(mgl32.Vec3{1, 1, 1}, mgl32.Vec3{-1, -1, -1})
	setMaterial(plane1)
	objects["plane1"] = plane1
	plane1.Location = mgl32.Vec3{0, -4, 0}
}
