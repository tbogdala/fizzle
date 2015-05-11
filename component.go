// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
  "fmt"
  "github.com/tbogdala/gombz"
  "github.com/tbogdala/groggy"
  gl "github.com/go-gl/gl/v3.3-core/gl"
  mgl "github.com/go-gl/mathgl/mgl32"
)

type ComponentMesh struct {
  // this filepath should be relative to component file
	SrcFile string

  // this filepath should be relative to component file
  BinFile string

	Textures []string
  Parent *Component

  srcMesh *gombz.Mesh
}

type ComponentChildRef struct {
  File string
  Location mgl.Vec3
}

type Component struct {
	Name string
  Meshes []*ComponentMesh
  ChildReferences []*ComponentChildRef

  componentDirPath string
  cachedRenderable *Renderable
}

func (c *Component) GetRenderable(tm *TextureManager) *Renderable {
  // see if we have a cached renderable already created
  if c.cachedRenderable != nil {
    return c.cachedRenderable
  }

  // start by creating a renderable to hold all of the meshes
  group := NewRenderable()
  group.IsGroup = true

  // now create renderables for all of the meshes.
  // comnponents only create new render nodes for the meshs defined and
  // not for referenced components
  for _,compMesh := range c.Meshes {
    cmRenderable := createRenderableForMesh(tm, compMesh)
    group.AddChild(cmRenderable)

    // cache it for later
    c.cachedRenderable = cmRenderable
  }

  return group
}

func (cm *ComponentMesh) GetFullSrcFilePath() string {
  return cm.Parent.componentDirPath + cm.SrcFile
}

func (cm *ComponentMesh) GetFullBinFilePath() string {
  return cm.Parent.componentDirPath + cm.BinFile
}

func (cm *ComponentMesh) GetFullTexturePath(textureIndex int) string {
  return cm.Parent.componentDirPath + cm.Textures[textureIndex]
}

func (cm *ComponentMesh) GetVertices() ([]mgl.Vec3, error) {
  if (cm.srcMesh == nil) {
    return nil, fmt.Errorf("No internal data present for component mesh to get vertices from.")
  }
  return cm.srcMesh.Vertices, nil
}


func createRenderableForMesh(tm *TextureManager, compMesh *ComponentMesh) *Renderable {
  // calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4

  // create the new renderable
  r := NewRenderable()
  r.Core = NewRenderableCore()

  // assign the texture
  if len(compMesh.Textures) > 0 {
    var okay bool
    r.Core.Tex0, okay = tm.GetTexture(compMesh.Textures[0])
    if !okay {
      groggy.Log("ERROR", "createRenderableForMesh failed to assign a texture gl id for %s.", compMesh.Textures[0])
    }
  }

  // set some basic properties up
  r.FaceCount = compMesh.srcMesh.FaceCount

  // create a buffer to hold all the data that is the same size as VertexCount
  vertBuffer := make([]float32, compMesh.srcMesh.VertexCount * 3)

  // setup verts and track the bounding rectangle
  for i,v := range compMesh.srcMesh.Vertices {
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
  if len(compMesh.srcMesh.Normals) > 0 {
    for i,n := range compMesh.srcMesh.Normals {
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
  if len(compMesh.srcMesh.Tangents) > 0 {
    for i,t := range compMesh.srcMesh.Tangents {
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
  if len(compMesh.srcMesh.UVChannels[0]) > 0 {
    uvChan := compMesh.srcMesh.UVChannels[0]
    for i:=uint32(0); i<compMesh.srcMesh.VertexCount; i++ {
      uv := uvChan[i]
      offset := i * 2
      vertBuffer[offset] = uv[0]
      vertBuffer[offset+1] = uv[1]
    }
    gl.GenBuffers(1, &r.Core.UvVBO)
    gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.UvVBO)
    gl.BufferData(gl.ARRAY_BUFFER, int(floatSize*compMesh.srcMesh.VertexCount*2), gl.Ptr(&vertBuffer[0]), gl.STATIC_DRAW)
  }

  // setup the face indices
  indexBuffer := make([]uint32, len(compMesh.srcMesh.Faces)*3)
  for i,f := range compMesh.srcMesh.Faces {
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
