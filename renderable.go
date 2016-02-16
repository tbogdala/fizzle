// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/gombz"
)

// RenderableCore contains data that is needed to draw an object on the screen.
// Further, data here can be shared between multiple Renderable instances.
type RenderableCore struct {
	Shader   *RenderShader
	Skeleton *Skeleton

	Tex0          graphics.Texture
	Tex1          graphics.Texture
	DiffuseColor  mgl.Vec4
	SpecularColor mgl.Vec4

	// Shininess is the exponent used while calculating specular highlights
	Shininess float32

	Vao            uint32
	VaoInitialized bool

	VertVBO        graphics.Buffer
	UvVBO          graphics.Buffer
	NormsVBO       graphics.Buffer
	TangentsVBO    graphics.Buffer
	ElementsVBO    graphics.Buffer
	BoneFidsVBO    graphics.Buffer
	BoneWeightsVBO graphics.Buffer
	ComboVBO1      graphics.Buffer
	ComboVBO2      graphics.Buffer

	IsDestroyed bool
}

// Rectangle3D defines a rectangular 3d structure by two points
type Rectangle3D struct {
	Bottom mgl.Vec3
	Top    mgl.Vec3
}

// DeltaX is the change of the X-axis component of Rectangle3D
func (rect *Rectangle3D) DeltaX() float32 {
	return rect.Top[0] - rect.Bottom[0]
}

// DeltaY is the change of the Y-axis component of Rectangle3D
func (rect *Rectangle3D) DeltaY() float32 {
	return rect.Top[1] - rect.Bottom[1]
}

// DeltaZ is the change of the Z-axis component of Rectangle3D
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

// NewRenderable creates a new Renderable object and a new RenderableCore
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

// NewRenderableCore creates a new RenderableCore object
func NewRenderableCore() *RenderableCore {
	rc := new(RenderableCore)
	rc.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	rc.SpecularColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	rc.Shininess = 0.01
	rc.Vao = gfx.GenVertexArray()
	return rc
}

// Destroy releases the RenderableCore data
func (r *Renderable) Destroy() {
	r.Core.DestroyCore()
}

// DestroyCore releases the OpenGL VBO and VAO objects but does not release
// things that could be shared like Tex0.
func (r *RenderableCore) DestroyCore() {
	gfx.DeleteBuffer(r.VertVBO)
	gfx.DeleteBuffer(r.UvVBO)
	gfx.DeleteBuffer(r.ElementsVBO)
	gfx.DeleteBuffer(r.TangentsVBO)
	gfx.DeleteBuffer(r.NormsVBO)
	gfx.DeleteBuffer(r.BoneFidsVBO)
	gfx.DeleteBuffer(r.BoneWeightsVBO)
	gfx.DeleteBuffer(r.ComboVBO1)
	gfx.DeleteBuffer(r.ComboVBO2)
	gfx.DeleteVertexArray(r.Vao)
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

func CreateFromGombz(srcMesh *gombz.Mesh) *Renderable {
	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4

	// create the new renderable
	r := NewRenderable()
	r.Core = NewRenderableCore()

	// setup a skeleton if the mesh has bones associated with it
	if srcMesh.BoneCount > 0 {
		r.Core.Skeleton = NewSkeleton(srcMesh.Bones, srcMesh.Animations)
	}

	// set some basic properties up
	r.FaceCount = srcMesh.FaceCount

	// create a buffer to hold all the data that is the same size as VertexCount
	vertBuffer := make([]float32, srcMesh.VertexCount*3)

	// setup verts and track the bounding rectangle
	for i, v := range srcMesh.Vertices {
		offset := i * 3
		vertBuffer[offset] = v[0]
		vertBuffer[offset+1] = v[1]
		vertBuffer[offset+2] = v[2]
	}
	r.Core.VertVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(vertBuffer), gfx.Ptr(&vertBuffer[0]), graphics.STATIC_DRAW)

	// calculate the bounding rectangle for the mesh
	r.BoundingRect = GetBoundingRect(vertBuffer)

	// setup normals
	if len(srcMesh.Normals) > 0 {
		for i, n := range srcMesh.Normals {
			offset := i * 3
			vertBuffer[offset] = n[0]
			vertBuffer[offset+1] = n[1]
			vertBuffer[offset+2] = n[2]
		}
		r.Core.NormsVBO = gfx.GenBuffer()
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.NormsVBO)
		gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(vertBuffer), gfx.Ptr(&vertBuffer[0]), graphics.STATIC_DRAW)
	}

	// setup tangents
	if len(srcMesh.Tangents) > 0 {
		for i, t := range srcMesh.Tangents {
			offset := i * 3
			vertBuffer[offset] = t[0]
			vertBuffer[offset+1] = t[1]
			vertBuffer[offset+2] = t[2]
		}
		r.Core.TangentsVBO = gfx.GenBuffer()
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.TangentsVBO)
		gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(vertBuffer), gfx.Ptr(&vertBuffer[0]), graphics.STATIC_DRAW)
	}

	// setup UVs
	if len(srcMesh.UVChannels[0]) > 0 {
		uvChan := srcMesh.UVChannels[0]
		for i := uint32(0); i < srcMesh.VertexCount; i++ {
			uv := uvChan[i]
			offset := i * 2
			vertBuffer[offset] = uv[0]
			vertBuffer[offset+1] = uv[1]
		}
		r.Core.UvVBO = gfx.GenBuffer()
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.UvVBO)
		gfx.BufferData(graphics.ARRAY_BUFFER, int(floatSize*srcMesh.VertexCount*2), gfx.Ptr(&vertBuffer[0]), graphics.STATIC_DRAW)
	}

	// setup vertex weight Ids for bones
	var weightBuffer []float32
	if len(srcMesh.VertexWeightIds) > 0 {
		if weightBuffer == nil {
			weightBuffer = make([]float32, srcMesh.VertexCount*4)
		}
		for i, v := range srcMesh.VertexWeightIds {
			offset := i * 4
			weightBuffer[offset] = v[0]
			weightBuffer[offset+1] = v[1]
			weightBuffer[offset+2] = v[2]
			weightBuffer[offset+3] = v[3]
		}
		r.Core.BoneFidsVBO = gfx.GenBuffer()
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.BoneFidsVBO)
		gfx.BufferData(graphics.ARRAY_BUFFER, int(floatSize*srcMesh.VertexCount*4), gfx.Ptr(&weightBuffer[0]), graphics.STATIC_DRAW)
	}

	// setup the vertex weights
	if len(srcMesh.VertexWeights) > 0 {
		if weightBuffer == nil {
			weightBuffer = make([]float32, srcMesh.VertexCount*4)
		}
		for i, v := range srcMesh.VertexWeights {
			offset := i * 4
			weightBuffer[offset] = v[0]
			weightBuffer[offset+1] = v[1]
			weightBuffer[offset+2] = v[2]
			weightBuffer[offset+3] = v[3]
		}
		r.Core.BoneWeightsVBO = gfx.GenBuffer()
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.BoneWeightsVBO)
		gfx.BufferData(graphics.ARRAY_BUFFER, int(floatSize*srcMesh.VertexCount*4), gfx.Ptr(&weightBuffer[0]), graphics.STATIC_DRAW)
	}

	// setup the face indices
	indexBuffer := make([]uint32, len(srcMesh.Faces)*3)
	for i, f := range srcMesh.Faces {
		offset := i * 3
		indexBuffer[offset] = f[0]
		indexBuffer[offset+1] = f[1]
		indexBuffer[offset+2] = f[2]
	}
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexBuffer), gfx.Ptr(&indexBuffer[0]), graphics.STATIC_DRAW)

	gfx.BindBuffer(graphics.ARRAY_BUFFER, 0)
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, 0)

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

// createTangents constructs the tangents for the faces.
// NOTE: this is a general implementation that assumes there's no shared
// vertices between faces.
func createTangents(verts []float32, indexes []uint32, uvs []float32) []float32 {
	tangents := make([]float32, len(verts))

	for i := 0; i < len(indexes); i += 3 {
		index0 := indexes[i+0]
		index1 := indexes[i+1]
		index2 := indexes[i+2]

		v0 := verts[(index0 * 3) : (index0*3)+3]
		v1 := verts[(index1 * 3) : (index1*3)+3]
		v2 := verts[(index2 * 3) : (index2*3)+3]

		uv0 := uvs[(index0 * 2) : (index0*2)+2]
		uv1 := uvs[(index1 * 2) : (index1*2)+2]
		uv2 := uvs[(index2 * 2) : (index2*2)+2]

		deltaPos1 := mgl.Vec3{v1[0] - v0[0], v1[1] - v0[1], v1[2] - v0[2]}
		deltaPos2 := mgl.Vec3{v2[0] - v0[0], v2[1] - v0[1], v2[2] - v0[2]}
		deltaUv1 := mgl.Vec2{uv1[0] - uv0[0], uv1[1] - uv0[1]}
		deltaUv2 := mgl.Vec2{uv2[0] - uv0[0], uv2[1] - uv0[1]}

		r := float32(1.0) / (deltaUv1[0]*deltaUv2[1] - deltaUv1[1]*deltaUv2[0])
		d1 := deltaPos1.Mul(deltaUv2[1])
		d2 := deltaPos2.Mul(deltaUv1[1])
		tangent := d1.Sub(d2)
		tangent = tangent.Mul(r).Normalize()

		// set the tangent array data for each vertex's tangent
		for f := 0; f < 3; f++ {
			index := indexes[i+f]

			tangents[index*3+0] = tangent[0]
			tangents[index*3+1] = tangent[1]
			tangents[index*3+2] = tangent[2]
		}
	}

	return tangents
}

func createPlane(shader string, x0, y0, x1, y1 float32, verts [12]float32, indexes [6]uint32, uvs [8]float32, normals [12]float32) *Renderable {
	const floatSize = 4
	const uintSize = 4

	// calculate the tangents based on the vertices and UVs.
	tangents := createTangents(verts[:], indexes[:], uvs[:])

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.ShaderName = shader
	r.FaceCount = 2
	r.BoundingRect.Bottom = mgl.Vec3{x0, y0, 0.0}
	r.BoundingRect.Top = mgl.Vec3{x1, y1, 0.0}

	// create a VBO to hold the vertex data
	r.Core.VertVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(verts), gfx.Ptr(&verts[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the uv data
	r.Core.UvVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.UvVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(uvs), gfx.Ptr(&uvs[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the normals data
	r.Core.NormsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.NormsVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(normals), gfx.Ptr(&normals[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the tangent data
	r.Core.TangentsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.TangentsVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(tangents), gfx.Ptr(&tangents[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gfx.Ptr(&indexes[0]), graphics.STATIC_DRAW)

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

	// calculate the tangents based on the vertices and UVs.
	tangents := createTangents(verts[:], indexes[:], uvs[:])

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.ShaderName = shader
	r.FaceCount = 12
	r.BoundingRect.Bottom = mgl.Vec3{xmin, ymin, zmin}
	r.BoundingRect.Top = mgl.Vec3{xmax, ymax, zmax}

	const floatSize = 4
	const uintSize = 4

	// create a VBO to hold the vertex data
	r.Core.VertVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(verts), gfx.Ptr(&verts[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the uv data
	r.Core.UvVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.UvVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(uvs), gfx.Ptr(&uvs[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the normals data
	r.Core.NormsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.NormsVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(normals), gfx.Ptr(&normals[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the tangent data
	r.Core.TangentsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.TangentsVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(tangents), gfx.Ptr(&tangents[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gfx.Ptr(&indexes[0]), graphics.STATIC_DRAW)

	return r
}

// CreateWireframeCube makes a cube with vertex and element VBO objects designed to be
// rendered as graphics.LINES.
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
	r.Core.VertVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(verts), gfx.Ptr(&verts[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gfx.Ptr(&indexes[0]), graphics.STATIC_DRAW)

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

	// calculate the tangents based on the vertices and UVs.
	// FIXME: disabled for now, there's an error in this code somewhere
	// where indexes end up eclipsing the number of vertices
	//tangents := createTangents(verts[:], indexes[:], uvs[:])

	r := NewRenderable()
	r.ShaderName = shader
	r.FaceCount = uint32(rings * sectors * 2)
	r.BoundingRect.Bottom = mgl.Vec3{-radius, -radius, -radius}
	r.BoundingRect.Top = mgl.Vec3{radius, radius, radius}

	const floatSize = 4
	const uintSize = 4

	// create a VBO to hold the vertex data
	r.Core.VertVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(verts), gfx.Ptr(&verts[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the uv data
	r.Core.UvVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.UvVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(uvs), gfx.Ptr(&uvs[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the normals data
	r.Core.NormsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.NormsVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(normals), gfx.Ptr(&normals[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the tangent data
	/*
		r.Core.TangentsVBO = gfx.GenBuffer()
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.TangentsVBO)
		gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(tangents), gfx.Ptr(&tangents[0]), graphics.STATIC_DRAW)
	*/

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gfx.Ptr(&indexes[0]), graphics.STATIC_DRAW)

	return r
}
