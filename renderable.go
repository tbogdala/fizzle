// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"math"

	gl "github.com/go-gl/gl/v3.3-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
)

// RenderableCore contains data that is needed to draw an object on the screen.
// Further, data here can be shared between multiple Renderable instances.
type RenderableCore struct {
	Shader   *RenderShader
	Skeleton *Skeleton

	Tex0          uint32
	DiffuseColor  mgl.Vec4
	SpecularColor mgl.Vec4

	// Shininess is the exponent used while calculating specular highlights
	Shininess float32

	Vao            uint32
	VaoInitialized bool

	VertVBO        uint32
	UvVBO          uint32
	NormsVBO       uint32
	TangentsVBO    uint32
	ElementsVBO    uint32
	BoneFidsVBO    uint32
	BoneWeightsVBO uint32
	ComboVBO1      uint32
	ComboVBO2      uint32

	IsDestroyed bool
}

// Rectangle3D defines a rectangular 3d structure by two points
type Rectangle3D struct {
	Bottom mgl.Vec3
	Top    mgl.Vec3
}

func (rect *Rectangle3D) DeltaX() float32 {
	return rect.Top[0] - rect.Bottom[0]
}
func (rect *Rectangle3D) DeltaY() float32 {
	return rect.Top[1] - rect.Bottom[1]
}
func (rect *Rectangle3D) DeltaZ() float32 {
	return rect.Top[2] - rect.Bottom[2]
}

// Renderable defines the data necessary to draw an object in OpenGL.
type Renderable struct {
	ShaderName string

	FaceCount     uint32
	Scale         mgl.Vec3
	Location      mgl.Vec3
	Rotation      mgl.Quat
	LocalRotation mgl.Quat

	// AnimationTime keeps track of the time value to use for the animation
	// currently applied (if any) to the Renderable.
	AnimationTime float32

	// BoundingRect is the unscaled, unrotated bounding rectangle for the renderable.
	BoundingRect Rectangle3D

	IsVisible bool
	IsGroup   bool

	Core     *RenderableCore
	Parent   *Renderable
	Children []*Renderable
}

func NewRenderable() *Renderable {
	r := new(Renderable)
	r.Location = mgl.Vec3{0.0, 0.0, 0.0}
	r.Scale = mgl.Vec3{1.0, 1.0, 1.0}
	r.Rotation = mgl.QuatIdent()
	r.LocalRotation = mgl.QuatIdent()
	r.IsVisible = true
	r.IsGroup = false
	r.Children = make([]*Renderable, 0, 4)
	r.ShaderName = ""
	r.Core = NewRenderableCore()
	return r
}

func NewRenderableCore() *RenderableCore {
	rc := new(RenderableCore)
	rc.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	rc.SpecularColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	rc.Shininess = 0.01
	gl.GenVertexArrays(1, &rc.Vao)
	return rc
}

// Destroy releases the RenderableCore data
func (r *Renderable) Destroy() {
	r.Core.DestroyCore()
}

// DestroyCore releases the OpenGL VBO and VAO objects but does not release
// things that could be shared like Tex0.
func (r *RenderableCore) DestroyCore() {
	gl.DeleteBuffers(1, &r.VertVBO)
	gl.DeleteBuffers(1, &r.UvVBO)
	gl.DeleteBuffers(1, &r.ElementsVBO)
	gl.DeleteBuffers(1, &r.TangentsVBO)
	gl.DeleteBuffers(1, &r.NormsVBO)
	gl.DeleteBuffers(1, &r.BoneFidsVBO)
	gl.DeleteBuffers(1, &r.BoneWeightsVBO)
	gl.DeleteBuffers(1, &r.ComboVBO1)
	gl.DeleteBuffers(1, &r.ComboVBO2)
	gl.DeleteBuffers(1, &r.Vao)
	r.IsDestroyed = true
}

// Clone makes a new Renderable object but shares the Core member between
// the two. This allows for a different location, scale, rotation, etc ...
func (r *Renderable) Clone() *Renderable {
	clone := NewRenderable()
	clone.FaceCount = r.FaceCount
	clone.Location = r.Location
	clone.Scale = r.Scale
	clone.Rotation = r.Rotation
	clone.LocalRotation = r.LocalRotation
	clone.IsVisible = r.IsVisible
	clone.IsGroup = r.IsGroup
	clone.ShaderName = r.ShaderName
	clone.BoundingRect = r.BoundingRect

	// The render core is shared in the clone
	clone.Core = r.Core

	// Deep clone the child renderables
	for _, rc := range r.Children {
		cloneChild := rc.Clone()
		clone.AddChild(cloneChild)
	}

	return clone
}

// HasSkeleton returns true if the renderable has bones associated with it.
func (r *Renderable) HasSkeleton() bool {
	if r.Core.Skeleton != nil {
		return true
	}
	return false
}

// HasSkeletonDeep returns true if the renderable, or any child, has bones associated with it.
func (r *Renderable) HasSkeletonDeep() bool {
	if r.Core.Skeleton != nil {
		return true
	}

	for _, cn := range r.Children {
		if cn.HasSkeletonDeep() == true {
			return true
		}
	}

	return false
}

// RenderableMapF defines the type of a function that can be passed to Renderable.Map().
type RenderableMapF func(r *Renderable)

// Map takes a function as a parameter that will be called for the renderable and all
// child Renderable objects (as well as their children, etc...)
func (r *Renderable) Map(f RenderableMapF) {
	// call the function for the renderable first
	f(r)

	// loop through all of the children and recurse
	for _, cn := range r.Children {
		cn.Map(f)
	}
}

// GetTransformMat4 creates a transform matrix: scale * transform
func (r *Renderable) GetTransformMat4() mgl.Mat4 {
	scaleMat := mgl.Scale3D(r.Scale[0], r.Scale[1], r.Scale[2])
	transMat := mgl.Translate3D(r.Location[0], r.Location[1], r.Location[2])
	localRotMat := r.LocalRotation.Mat4()
	rotMat := r.Rotation.Mat4()
	modelTransform := rotMat.Mul4(transMat).Mul4(localRotMat).Mul4(scaleMat)
	if r.Parent == nil {
		return modelTransform
	}

	// if there's a parent, apply the transform as well
	parentTransform := r.Parent.GetTransformMat4()
	return parentTransform.Mul4(modelTransform)
}

// AddChild sets the Renderable to be a child of the parent renderable.
func (r *Renderable) AddChild(child *Renderable) {
	r.Children = append(r.Children, child)
	child.Parent = r
}

// GetBoundingRect returns a bounding Rectangle3D for all of the vertices
// passed in.
func GetBoundingRect(verts []float32) (r Rectangle3D) {
	var minx, miny, minz float32 = math.MaxFloat32, math.MaxFloat32, math.MaxFloat32
	var maxx, maxy, maxz float32 = math.MaxFloat32 * -1, math.MaxFloat32 * -1, math.MaxFloat32 * -1

	vertCount := len(verts) / 3
	for i := 0; i < vertCount; i++ {
		offset := i * 3
		x := verts[offset]
		y := verts[offset+1]
		z := verts[offset+2]

		if x < minx {
			minx = x
		}
		if x > maxx {
			maxx = x
		}
		if y < miny {
			miny = y
		}
		if y > maxy {
			maxy = y
		}
		if z < minz {
			minz = z
		}
		if z > maxz {
			maxz = z
		}
	}

	r.Bottom = mgl.Vec3{minx, miny, minz}
	r.Top = mgl.Vec3{maxx, maxy, maxz}
	return r
}

// CreatePlaneXY makes a 2d Renderable object on the XY plane for the given size,
// where (x0,y0) is the lower left and (x1, y1) is the upper right coordinate.
func CreatePlaneXY(shader string, x0, y0, x1, y1 float32) *Renderable {
	verts := [12]float32{
		x0, y0, 0.0,
		x1, y0, 0.0,
		x0, y1, 0.0,
		x1, y1, 0.0,
	}
	indexes := [6]uint32{
		0, 1, 2,
		1, 3, 2,
	}
	uvs := [8]float32{
		0.0, 0.0,
		1.0, 0.0,
		0.0, 1.0,
		1.0, 1.0,
	}
	normals := [12]float32{
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
	}

	return createPlane(shader, x0, y0, x1, y1, verts, indexes, uvs, normals)
}

// CreatePlaneXZ makes a 2d Renderable object on the XZ plane for the given size,
// where (x0,z0) is the lower left and (x1, z1) is the upper right coordinate.
func CreatePlaneXZ(shader string, x0, z0, x1, z1 float32) *Renderable {
	verts := [12]float32{
		x0, 0.0, z0,
		x1, 0.0, z0,
		x0, 0.0, z1,
		x1, 0.0, z1,
	}
	indexes := [6]uint32{
		0, 1, 2,
		1, 3, 2,
	}
	uvs := [8]float32{
		0.0, 0.0,
		1.0, 0.0,
		0.0, 1.0,
		1.0, 1.0,
	}
	normals := [12]float32{
		0.0, 1.0, 0.0,
		0.0, 1.0, 0.0,
		0.0, 1.0, 0.0,
		0.0, 1.0, 0.0,
	}

	return createPlane(shader, x0, z0, x1, z1, verts, indexes, uvs, normals)
}

func createPlane(shader string, x0, y0, x1, y1 float32, verts [12]float32, indexes [6]uint32, uvs [8]float32, normals [12]float32) *Renderable {
	const floatSize = 4
	const uintSize = 4

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.ShaderName = shader
	r.FaceCount = 2
	r.BoundingRect.Bottom = mgl.Vec3{x0, y0, 0.0}
	r.BoundingRect.Top = mgl.Vec3{x1, y1, 0.0}

	// create a VBO to hold the vertex data
	gl.GenBuffers(1, &r.Core.VertVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.VertVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(verts), gl.Ptr(&verts[0]), gl.STATIC_DRAW)

	// create a VBO to hold the uv data
	gl.GenBuffers(1, &r.Core.UvVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.UvVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(uvs), gl.Ptr(&uvs[0]), gl.STATIC_DRAW)

	// create a VBO to hold the normals data
	gl.GenBuffers(1, &r.Core.NormsVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.NormsVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(normals), gl.Ptr(&normals[0]), gl.STATIC_DRAW)

	// create a VBO to hold the face indexes
	gl.GenBuffers(1, &r.Core.ElementsVBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gl.Ptr(&indexes[0]), gl.STATIC_DRAW)

	return r
}

func CreateCube(shader string, xmin, ymin, zmin, xmax, ymax, zmax float32) *Renderable {
	/* Cube vertices are layed out like this:

	  +--------+           6          5
	/ |       /|
	+--------+ |        1          0        +Y
	| |      | |                            |___ +X
	| +------|-+           7          4    /
	|/       |/                           +Z
	+--------+          2          3

	*/

	verts := [...]float32{
		xmax, ymax, zmax, xmin, ymax, zmax, xmin, ymin, zmax, xmax, ymin, zmax, // v0,v1,v2,v3 (front)
		xmax, ymax, zmin, xmax, ymax, zmax, xmax, ymin, zmax, xmax, ymin, zmin, // v5,v0,v3,v4 (right)
		xmax, ymax, zmin, xmin, ymax, zmin, xmin, ymax, zmax, xmax, ymax, zmax, // v5,v6,v1,v0 (top)
		xmin, ymax, zmax, xmin, ymax, zmin, xmin, ymin, zmin, xmin, ymin, zmax, // v1,v6,v7,v2 (left)
		xmax, ymin, zmax, xmin, ymin, zmax, xmin, ymin, zmin, xmax, ymin, zmin, // v3,v2,v7,v4 (bottom)
		xmin, ymax, zmin, xmax, ymax, zmin, xmax, ymin, zmin, xmin, ymin, zmin, // v6,v5,v4,v7 (back)
	}
	indexes := [...]uint32{
		0, 1, 2, 2, 3, 0,
		4, 5, 6, 6, 7, 4,
		8, 9, 10, 10, 11, 8,
		12, 13, 14, 14, 15, 12,
		16, 17, 18, 18, 19, 16,
		20, 21, 22, 22, 23, 20,
	}
	uvs := [...]float32{
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
	}
	normals := [...]float32{
		0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, // v0,v1,v2,v3 (front)
		1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, // v5,v0,v3,v4 (right)
		0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, // v5,v6,v1,v0 (top)
		-1, 0, 0, -1, 0, 0, -1, 0, 0, -1, 0, 0, // v1,v6,v7,v2 (left)
		0, -1, 0, 0, -1, 0, 0, -1, 0, 0, -1, 0, // v3,v2,v7,v4 (bottom)
		0, 0, -1, 0, 0, -1, 0, 0, -1, 0, 0, -1, // v6,v5,v4,v7 (back)
	}

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.ShaderName = shader
	r.FaceCount = 12
	r.BoundingRect.Bottom = mgl.Vec3{xmin, ymin, zmin}
	r.BoundingRect.Top = mgl.Vec3{xmax, ymax, zmax}

	const floatSize = 4
	const uintSize = 4

	// create a VBO to hold the vertex data
	gl.GenBuffers(1, &r.Core.VertVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.VertVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(verts), gl.Ptr(&verts[0]), gl.STATIC_DRAW)

	// create a VBO to hold the uv data
	gl.GenBuffers(1, &r.Core.UvVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.UvVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(uvs), gl.Ptr(&uvs[0]), gl.STATIC_DRAW)

	// create a VBO to hold the normals data
	gl.GenBuffers(1, &r.Core.NormsVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.NormsVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(normals), gl.Ptr(&normals[0]), gl.STATIC_DRAW)

	// create a VBO to hold the face indexes
	gl.GenBuffers(1, &r.Core.ElementsVBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gl.Ptr(&indexes[0]), gl.STATIC_DRAW)

	return r
}

// CreateWireframeCube makes a cube with vertex and element VBO objects designed to be
// rendered as gl.LINES.
func CreateWireframeCube(shader string, xmin, ymin, zmin, xmax, ymax, zmax float32) *Renderable {
	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4
	const facesPerCollision = 16

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.ShaderName = shader
	r.FaceCount = facesPerCollision
	r.BoundingRect.Bottom = mgl.Vec3{xmin, ymin, zmin}
	r.BoundingRect.Top = mgl.Vec3{xmax, ymax, zmax}

	/* Cube vertices are layed out like this:

	  +--------+           6          5
	/ |       /|
	+--------+ |        1          0        +Y
	| |      | |                            |___ +X
	| +------|-+           7          4    /
	|/       |/                           +Z
	+--------+          2          3

	*/
	verts := [...]float32{
		xmax, ymax, zmax, xmin, ymax, zmax, xmin, ymin, zmax, xmax, ymin, zmax, // v0,v1,v2,v3 (front)
		xmax, ymax, zmin, xmax, ymax, zmax, xmax, ymin, zmax, xmax, ymin, zmin, // v5,v0,v3,v4 (right)
		xmin, ymax, zmax, xmin, ymax, zmin, xmin, ymin, zmin, xmin, ymin, zmax, // v1,v6,v7,v2 (left)
		xmin, ymax, zmin, xmax, ymax, zmin, xmax, ymin, zmin, xmin, ymin, zmin, // v6,v5,v4,v7 (back)
	}
	indexes := [...]uint32{
		0, 1, 1, 2, 2, 3, 3, 0,
		4, 5, 5, 6, 6, 7, 7, 4,
		8, 9, 9, 10, 10, 11, 11, 8,
		12, 13, 13, 14, 14, 15, 15, 12,
	}

	r.BoundingRect.Bottom = mgl.Vec3{xmin, ymin, zmin}
	r.BoundingRect.Top = mgl.Vec3{xmax, ymax, zmax}

	// create a VBO to hold the vertex data
	gl.GenBuffers(1, &r.Core.VertVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.VertVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(verts), gl.Ptr(&verts[0]), gl.STATIC_DRAW)

	// create a VBO to hold the face indexes
	gl.GenBuffers(1, &r.Core.ElementsVBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gl.Ptr(&indexes[0]), gl.STATIC_DRAW)

	return r
}

// CreateSphere generates a 3d uv-sphere with the given radius and returns a Renderable.
func CreateSphere(shader string, radius float32, rings int, sectors int) *Renderable {
	// nothing to create
	if rings < 2 || sectors < 2 {
		return nil
	}

	const piDiv2 = math.Pi / 2.0

	verts := make([]float32, 0, rings*sectors)
	indexes := make([]uint32, 0, rings*sectors)
	uvs := make([]float32, 0, rings*sectors)
	normals := make([]float32, 0, rings*sectors)

	R := float64(1.0 / float32(rings-1))
	S := float64(1.0 / float32(sectors-1))

	for ri := 0; ri < int(rings); ri++ {
		for si := 0; si < int(sectors); si++ {
			y := float32(math.Sin(-piDiv2 + math.Pi*float64(ri)*R))
			x := float32(math.Cos(2.0*math.Pi*float64(si)*S) * math.Sin(math.Pi*float64(ri)*R))
			z := float32(math.Sin(2.0*math.Pi*float64(si)*S) * math.Sin(math.Pi*float64(ri)*R))

			uvs = append(uvs, float32(si)*float32(S))
			uvs = append(uvs, float32(ri)*float32(R))

			verts = append(verts, x*radius)
			verts = append(verts, y*radius)
			verts = append(verts, z*radius)

			normals = append(normals, x)
			normals = append(normals, y)
			normals = append(normals, z)

			currentRow := ri * sectors
			nextRow := (ri + 1) * sectors

			indexes = append(indexes, uint32(currentRow+si))
			indexes = append(indexes, uint32(nextRow+si))
			indexes = append(indexes, uint32(nextRow+si+1))

			indexes = append(indexes, uint32(currentRow+si))
			indexes = append(indexes, uint32(nextRow+si+1))
			indexes = append(indexes, uint32(currentRow+si+1))
		}
	}

	r := NewRenderable()
	r.ShaderName = shader
	r.FaceCount = uint32(rings * sectors * 2)
	r.BoundingRect.Bottom = mgl.Vec3{-radius, -radius, -radius}
	r.BoundingRect.Top = mgl.Vec3{radius, radius, radius}

	const floatSize = 4
	const uintSize = 4

	// create a VBO to hold the vertex data
	gl.GenBuffers(1, &r.Core.VertVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.VertVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(verts), gl.Ptr(&verts[0]), gl.STATIC_DRAW)

	// create a VBO to hold the uv data
	gl.GenBuffers(1, &r.Core.UvVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.UvVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(uvs), gl.Ptr(&uvs[0]), gl.STATIC_DRAW)

	// create a VBO to hold the normals data
	gl.GenBuffers(1, &r.Core.NormsVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.NormsVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(normals), gl.Ptr(&normals[0]), gl.STATIC_DRAW)

	// create a VBO to hold the face indexes
	gl.GenBuffers(1, &r.Core.ElementsVBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gl.Ptr(&indexes[0]), gl.STATIC_DRAW)

	return r
}
