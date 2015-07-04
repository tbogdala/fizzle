// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"fmt"

	gl "github.com/go-gl/gl/v3.3-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/gombz"
	"github.com/tbogdala/groggy"
)

// ComponentMesh defines a mesh reference for a component and everything
// needed to draw it.
type ComponentMesh struct {
	// SrcFile is a filepath should be relative to component file
	SrcFile string

	// BinFile is a filepath should be relative to component file
	BinFile string

	// Textures specifies the texture files to load for mesh, relative
	// to the component file
	Textures []string

	// Offset is the location offset of the mesh in the component.
	Offset mgl.Vec3

	// Parent is the owning Component object
	Parent *Component

	// SrcMesh is the cached mesh data either from SrcFile or BinFile
	SrcMesh *gombz.Mesh
}

// ComponentChildRef defines a reference to another component JSON file
// so that Components can be built from other Component parts
type ComponentChildRef struct {
	File     string
	Location mgl.Vec3
}

// ComponentMaterial defines the material appearance of the component.
type ComponentMaterial struct {
	// Diffuse color for the material
	Diffuse mgl.Vec4
}

// CollisionRef specifies a collision object within the component
// (e.g. a collision cube for a wall).
// Note: right now it only supports AABB collisions.
type CollisionRef struct {
	Min  mgl.Vec3
	Max  mgl.Vec3
	Tags []string
}

// Component is the main structure for component JSON files.
type Component struct {
	// The name of the component
	Name string

	// Location is the location of the component.
	Location mgl.Vec3

	// All of the meshes that are part of this component
	Meshes []*ComponentMesh

	// The material description of the component
	Material *ComponentMaterial

	// ChildReferences can be specified to include other components
	// to be contained in this component.
	ChildReferences []*ComponentChildRef

	// Collision objects for the component
	Collisions []*CollisionRef

	// Properties is a map for client software's custom properties for the component.
	Properties map[string]string

	// this is the directory path for the component file if it was loaded
	// from JSON.
	componentDirPath string

	// this is the cached renerable object for the component that can
	// be used as a prototype.
	cachedRenderable *Renderable
}

// Destroy will destroy the cached Renderable object if it exists.
func (c *Component) Destroy() {
	if c.cachedRenderable != nil {
		c.cachedRenderable.Destroy()
	}
}

// Clone makes a new component and then copies the members over
// to the new object. This means that Meshes, Collisions, ChildReferences, etc...
// are shared between the clones.
func (c *Component) Clone() *Component {
	clone := new(Component)

	// copy over all of the fields
	clone.Name = c.Name
	clone.Location = c.Location
	clone.Meshes = c.Meshes
	clone.ChildReferences = c.ChildReferences
	clone.Collisions = c.Collisions
	clone.Properties = c.Properties
	clone.Material = c.Material
	clone.componentDirPath = c.componentDirPath
	clone.cachedRenderable = c.cachedRenderable

	return clone
}

// GetRenderable will return the cached renderable object for the component
// or create one if it hasn't been made yet. The TextureManager is needed
// to resolve texture references.
func (c *Component) GetRenderable(tm *TextureManager) *Renderable {
	// see if we have a cached renderable already created
	if c.cachedRenderable != nil {
		return c.cachedRenderable
	}

	// start by creating a renderable to hold all of the meshes
	group := NewRenderable()
	group.IsGroup = true
	group.Location = c.Location

	// now create renderables for all of the meshes.
	// comnponents only create new render nodes for the meshs defined and
	// not for referenced components
	for _, compMesh := range c.Meshes {
		cmRenderable := createRenderableForMesh(tm, compMesh)
		group.AddChild(cmRenderable)

		// assign material properties if specified
		if c.Material != nil {
			cmRenderable.Core.DiffuseColor = c.Material.Diffuse
		}

		// cache it for later
		c.cachedRenderable = cmRenderable
	}

	return group
}

// GetFullSrcFilePath returns the full file path for the mesh source file.
func (cm *ComponentMesh) GetFullSrcFilePath() string {
	return cm.Parent.componentDirPath + cm.SrcFile
}

// GetFullBinFilePath returns the full file path for the mesh binary file (gombz format).
func (cm *ComponentMesh) GetFullBinFilePath() string {
	return cm.Parent.componentDirPath + cm.BinFile
}

// GetFullTexturePath returns the full file path for the mesh texture.
func (cm *ComponentMesh) GetFullTexturePath(textureIndex int) string {
	return cm.Parent.componentDirPath + cm.Textures[textureIndex]
}

// GetVertices returns the vector slice containing the vertices for the mesh.
func (cm *ComponentMesh) GetVertices() ([]mgl.Vec3, error) {
	if cm.SrcMesh == nil {
		return nil, fmt.Errorf("No internal data present for component mesh to get vertices from.")
	}
	return cm.SrcMesh.Vertices, nil
}

// createRenderableForMesh does the work of creating the Renderable and putting all of
// the mesh data into VBOs.
func createRenderableForMesh(tm *TextureManager, compMesh *ComponentMesh) *Renderable {
	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4

	// create the new renderable
	r := NewRenderable()
	r.Core = NewRenderableCore()

	// setup a skeleton if the mesh has bones associated with it
	if compMesh.SrcMesh.BoneCount > 0 {
		r.Core.Skeleton = NewSkeleton(compMesh.SrcMesh.Bones, compMesh.SrcMesh.Animations)
	}

	// assign the texture
	if len(compMesh.Textures) > 0 {
		var okay bool
		r.Core.Tex0, okay = tm.GetTexture(compMesh.Textures[0])
		if !okay {
			groggy.Log("ERROR", "createRenderableForMesh failed to assign a texture gl id for %s.", compMesh.Textures[0])
		}
	}

	// set some basic properties up
	r.FaceCount = compMesh.SrcMesh.FaceCount
	r.Location = compMesh.Offset

	// create a buffer to hold all the data that is the same size as VertexCount
	vertBuffer := make([]float32, compMesh.SrcMesh.VertexCount*3)

	// setup verts and track the bounding rectangle
	for i, v := range compMesh.SrcMesh.Vertices {
		offset := i * 3
		vertBuffer[offset] = v[0]
		vertBuffer[offset+1] = v[1]
		vertBuffer[offset+2] = v[2]
	}
	gl.GenBuffers(1, &r.Core.VertVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.VertVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(vertBuffer), gl.Ptr(&vertBuffer[0]), gl.STATIC_DRAW)

	// calculate the bounding rectangle for the mesh
	r.BoundingRect = GetBoundingRect(vertBuffer)

	// setup normals
	if len(compMesh.SrcMesh.Normals) > 0 {
		for i, n := range compMesh.SrcMesh.Normals {
			offset := i * 3
			vertBuffer[offset] = n[0]
			vertBuffer[offset+1] = n[1]
			vertBuffer[offset+2] = n[2]
		}
		gl.GenBuffers(1, &r.Core.NormsVBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.NormsVBO)
		gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(vertBuffer), gl.Ptr(&vertBuffer[0]), gl.STATIC_DRAW)
	}

	// setup tangents
	if len(compMesh.SrcMesh.Tangents) > 0 {
		for i, t := range compMesh.SrcMesh.Tangents {
			offset := i * 3
			vertBuffer[offset] = t[0]
			vertBuffer[offset+1] = t[1]
			vertBuffer[offset+2] = t[2]
		}
		gl.GenBuffers(1, &r.Core.TangentsVBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.TangentsVBO)
		gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(vertBuffer), gl.Ptr(&vertBuffer[0]), gl.STATIC_DRAW)
	}

	// setup UVs
	if len(compMesh.SrcMesh.UVChannels[0]) > 0 {
		uvChan := compMesh.SrcMesh.UVChannels[0]
		for i := uint32(0); i < compMesh.SrcMesh.VertexCount; i++ {
			uv := uvChan[i]
			offset := i * 2
			vertBuffer[offset] = uv[0]
			vertBuffer[offset+1] = uv[1]
		}
		gl.GenBuffers(1, &r.Core.UvVBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.UvVBO)
		gl.BufferData(gl.ARRAY_BUFFER, int(floatSize*compMesh.SrcMesh.VertexCount*2), gl.Ptr(&vertBuffer[0]), gl.STATIC_DRAW)
	}

	// setup vertex weight Ids for bones
	var weightBuffer []float32
	if len(compMesh.SrcMesh.VertexWeightIds) > 0 {
		if weightBuffer == nil {
			weightBuffer = make([]float32, compMesh.SrcMesh.VertexCount*4)
		}
		for i, v := range compMesh.SrcMesh.VertexWeightIds {
			offset := i * 4
			weightBuffer[offset] = v[0]
			weightBuffer[offset+1] = v[1]
			weightBuffer[offset+2] = v[2]
			weightBuffer[offset+3] = v[3]
		}
		gl.GenBuffers(1, &r.Core.BoneFidsVBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.BoneFidsVBO)
		gl.BufferData(gl.ARRAY_BUFFER, int(floatSize*compMesh.SrcMesh.VertexCount*4), gl.Ptr(&weightBuffer[0]), gl.STATIC_DRAW)
	}

	// setup the vertex weights
	if len(compMesh.SrcMesh.VertexWeights) > 0 {
		if weightBuffer == nil {
			weightBuffer = make([]float32, compMesh.SrcMesh.VertexCount*4)
		}
		for i, v := range compMesh.SrcMesh.VertexWeights {
			offset := i * 4
			weightBuffer[offset] = v[0]
			weightBuffer[offset+1] = v[1]
			weightBuffer[offset+2] = v[2]
			weightBuffer[offset+3] = v[3]
		}
		gl.GenBuffers(1, &r.Core.BoneWeightsVBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.BoneWeightsVBO)
		gl.BufferData(gl.ARRAY_BUFFER, int(floatSize*compMesh.SrcMesh.VertexCount*4), gl.Ptr(&weightBuffer[0]), gl.STATIC_DRAW)
	}

	// setup the face indices
	indexBuffer := make([]uint32, len(compMesh.SrcMesh.Faces)*3)
	for i, f := range compMesh.SrcMesh.Faces {
		offset := i * 3
		indexBuffer[offset] = f[0]
		indexBuffer[offset+1] = f[1]
		indexBuffer[offset+2] = f[2]
	}
	gl.GenBuffers(1, &r.Core.ElementsVBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, uintSize*len(indexBuffer), gl.Ptr(&indexBuffer[0]), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	return r
}
