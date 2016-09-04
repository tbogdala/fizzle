// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/gombz"
)

const (
	// MaxRenderableTextures is the maximum number of textures that can get assigned
	// to a renderable.
	MaxRenderableTextures = 8
)

// RenderableCore contains data that is needed to draw an object on the screen.
// Further, data here can be shared between multiple Renderable instances.
type RenderableCore struct {
	// Shader is the program used to render this object; This can be overridden
	// by using a DrawWithShader* function.
	Shader *RenderShader

	// Skeleton is the animatable skeleton object for the renderable.
	Skeleton *Skeleton

	// Tex is the array of textures for the renderable.
	Tex [MaxRenderableTextures]graphics.Texture

	// DiffuseColor is the material color for the renderable. This is
	// displayed outright by the shader or often blended with the
	// diffuse texture.
	DiffuseColor mgl.Vec4

	// SpecularColor is the material specular color for the renderable
	// and is used to control the color of the specular highlight.
	//
	// It can be thought of the topcoat layer color to the DiffuseColor's
	// base paint layer color.
	SpecularColor mgl.Vec4

	// Shininess is the specular coefficient used to control the tightness
	// of the specular highlight. It represents the power the specular factor will
	// be raised to -- therefore values between (0.0 - 1.0) will yield different
	// results than values >= 1.0.
	Shininess float32

	// Vao is the OpenGL vertex array object for the renderable.
	Vao uint32

	// VaoInitialized indicates whether or not the Vao has been bound yet.
	VaoInitialized bool

	// VertVBO indicates the VBO that contains the vertex data.
	VertVBO graphics.Buffer

	// UvVBO indicates the VBO that contains the UV data.
	UvVBO graphics.Buffer

	// NormsVBO indicates the VBO that contains the normal vector data.
	NormsVBO graphics.Buffer

	// TangentsVBO indicates the VBO that contains the tangent vector data.
	TangentsVBO graphics.Buffer

	// ElementsVBO indicates the VBO that defines the element indices data.
	ElementsVBO graphics.Buffer

	// BoneFidsVBO indicates the VBO that defines what bones affect a given vertex.
	// These are passed to the shader as floats due to better compatability with GLSL.
	BoneFidsVBO graphics.Buffer

	// BoneWeightsVBO indicates the VBO that defines how strong a bone affects a given vertex.
	BoneWeightsVBO graphics.Buffer

	// ComboVBO1 is a user-customizable VBO object to be bound to the RenderShader.
	ComboVBO1 graphics.Buffer

	// ComboVBO2 is a user-customizable VBO object to be bound to the RenderShader.
	ComboVBO2 graphics.Buffer

	// VBOStride should be set to the total number of bytes that exist in a VBO for
	// a single object.
	//
	// This is set to non-zero when members like VertVBO, UvVBO,
	// NormsVBO, etc... are set tot the same VBO object but the data is interleaved.
	// When this is the case, the corresponding Offset members (e.g. VertVBOOffset)
	// should be set as well.
	VBOStride int32

	// VertVBOOffset is the offset in bytes from the start of a vertex definition needed
	// to read the vertex information.
	VertVBOOffset int

	// UvVBOOffset is the offset in bytes from the start of a vertex definition needed
	// to read the UV information.
	UvVBOOffset int

	// NormsVBOOffset is the offset in bytes from the start of a vertex definition needed
	// to read the normal vector information.
	NormsVBOOffset int

	// TangentsVBOOffset is the offset in bytes from the start of a vertex definition needed
	// to read the tangent vector information.
	TangentsVBOOffset int

	// BoneFidsVBOOffset is the offset in bytes from the start of a vertex definition needed
	// to read the bone float id information.
	BoneFidsVBOOffset int

	// BoneWeightsVBOOffset is the offset in bytes from the start of a vertex definition needed
	// to read the bone weighting information.
	BoneWeightsVBOOffset int

	// ComboVBO1Offset is the offset in bytes from the start of a vertex definition needed
	// to read the customizable information.
	ComboVBO1Offset int

	// ComboVBO2Offset is the offset in bytes from the start of a vertex definition needed
	// to read the customizable information.
	ComboVBO2Offset int

	// IsDestroyed should be set to true if the Renderable has been Destroy()'d.
	IsDestroyed bool
}

// Rectangle3D defines a rectangular 3d structure by two points.
type Rectangle3D struct {
	// Bottom defines the bottom corner opposite of Top.
	Bottom mgl.Vec3

	// TOp defines the top corner opposite of Bottom.
	Top mgl.Vec3
}

// DeltaX is the change of the X-axis component of a Rectangle3D.
func (rect *Rectangle3D) DeltaX() float32 {
	return rect.Top[0] - rect.Bottom[0]
}

// DeltaY is the change of the Y-axis component of a Rectangle3D.
func (rect *Rectangle3D) DeltaY() float32 {
	return rect.Top[1] - rect.Bottom[1]
}

// DeltaZ is the change of the Z-axis component of a Rectangle3D.
func (rect *Rectangle3D) DeltaZ() float32 {
	return rect.Top[2] - rect.Bottom[2]
}

// Renderable defines the data necessary to draw an object in OpenGL.
// This structure focuses more on 'instance' type of data which is
// typically not sharable between multiple Renderable instances.
type Renderable struct {
	// FaceCount specifies how many elements to draw when rendered.
	FaceCount uint32

	// Scale is the scaling vector for the Renderable used to modify
	// the size of the object.
	Scale mgl.Vec3

	// Location is the world-space location of Renderable.
	Location mgl.Vec3

	// Rotation is the world-space rotation quaternion of the Renderable, which is
	// different than LocalRotation and the pivot is {0,0,0} in world-space.
	Rotation mgl.Quat

	// LocalRotation is the local rotation quaternion for the Renderable, which is
	// different than the Rotation property and pivots around the model's origin.
	LocalRotation mgl.Quat

	// AnimationTime keeps track of the time value to use for the animation
	// currently applied (if any) to the Renderable.
	AnimationTime float32

	// BoundingRect is the unscaled, unrotated bounding rectangle for the renderable.
	BoundingRect Rectangle3D

	// IsVisible should be set to true if the object is to be rendered.
	IsVisible bool

	// IsGroup should be set to true if the renderable should only render its children objects
	// and that this Renderable itself should not be drawn.
	IsGroup bool

	// Core is the RenderableCore object that contains the renderable data that can
	// be shadered between multiple Renderable objects if needed.
	Core *RenderableCore

	// Parent can be set to a Renderable that should be considered this Renderable's
	// 'Parent' which will make some properties relative to this parent object (e.g.
	// Location).
	Parent *Renderable

	// Children is a slice of Renderables that are the Renderable's children objects
	// that should be drawn with this renderable.
	Children []*Renderable
}

// NewRenderable creates a new Renderable object and a new RenderableCore.
func NewRenderable() *Renderable {
	r := new(Renderable)
	r.Location = mgl.Vec3{0.0, 0.0, 0.0}
	r.Scale = mgl.Vec3{1.0, 1.0, 1.0}
	r.Rotation = mgl.QuatIdent()
	r.LocalRotation = mgl.QuatIdent()
	r.IsVisible = true
	r.IsGroup = false
	r.Children = make([]*Renderable, 0, 4)

	r.Core = NewRenderableCore()
	return r
}

// NewRenderableCore creates a new RenderableCore object.
func NewRenderableCore() *RenderableCore {
	rc := new(RenderableCore)
	rc.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	rc.SpecularColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	rc.Shininess = 1.00
	rc.Vao = gfx.GenVertexArray()
	return rc
}

// Destroy releases the RenderableCore data.
func (r *Renderable) Destroy() {
	r.Core.DestroyCore()
}

// DestroyCore releases the OpenGL VBO and VAO objects but does not release
// things that could be shared like Tex0 and then marks the object as destroyed.
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

// HasSkeleton returns true if the Renderable has bones associated with it.
func (r *Renderable) HasSkeleton() bool {
	if r.Core.Skeleton != nil {
		return true
	}
	return false
}

// HasSkeletonDeep returns true if the Renderable, or any child, has bones associated with it.
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

// Map takes a function as a parameter that will be called for the Renderable and all
// child Renderable objects recursively.
func (r *Renderable) Map(f RenderableMapF) {
	// call the function for the renderable first
	f(r)

	// loop through all of the children and recurse
	for _, cn := range r.Children {
		cn.Map(f)
	}
}

// GetTransformMat4 creates a transform matrix that can be used to transform
// a vertex of the Renderable into world space.
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

// AddChild sets the Renderable passed in as a child of the renderable.
func (r *Renderable) AddChild(child *Renderable) {
	r.Children = append(r.Children, child)
	child.Parent = r
}

// GetBoundingRect calculates a bounding Rectangle3D for all of the vertices pssed in.
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

// CreateFromGombz creates a new Renderable based on model data from
// a GOMBZ file. (http://www.github.com/tbogdala/gombz)
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
