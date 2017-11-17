// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package component

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/gombz"
)

// Mesh defines a mesh reference for a component and everything
// needed to draw it.
type Mesh struct {
	// Name is the user identifier for the mesh in the component.
	Name string

	// The material describes visual attributes of the component.
	Material Material

	// SrcFile is a filepath, relative to the component file,
	// for the source binary of the model to load.
	SrcFile string

	// BinFile is a filepath should be relative to component file
	// for the Gombz binary of the model to load.
	BinFile string

	// Offset is the location offset of the mesh in the component
	// specified in local coordinates.
	Offset mgl.Vec3

	// Scale is the scaling vector for the mesh in the component.
	Scale mgl.Vec3

	// RotationAxis specifies the amount of degrees to rotate along
	// an axis -- the order of which is determined in RotationOrder.
	RotationAxis mgl.Vec3

	// RotationOrder determines the order in which to apply the rotations
	// from RotationAxis.
	// See: mgl32.AnglesToQuat() for the valid values (e.g. mgl32.XYZ)
	RotationOrder mgl.RotationOrder

	// Parent is the owning Component object, if any.
	Parent *Component `json:"-"`

	// SrcMesh is the cached mesh data either from BinFile.
	SrcMesh *gombz.Mesh `json:"-"`
}

// NewMesh creates a new Mesh object with sane defaults.
func NewMesh() *Mesh {
	cm := new(Mesh)
	cm.RotationOrder = mgl.XYZ
	cm.Scale = mgl.Vec3{1, 1, 1}
	cm.Material.Diffuse = mgl.Vec4{1, 1, 1, 1}
	cm.Material.Specular = mgl.Vec4{1, 1, 1, 1}
	cm.Material.GenerateMipmaps = true
	return cm
}

// ChildRef defines a reference to another component JSON file
// so that Components can be built from other Component parts.
type ChildRef struct {
	// File is the component file to load as a child object and the location should
	// be relative to the main component's file path.
	File string

	// Location is the location of the child object in the component.
	Location mgl.Vec3

	// RotationAxis is the axis by which to rotate the child component around; this
	// is only valid if RotationDegrees is non-zero.
	RotationAxis mgl.Vec3

	// RotationDegrees is the amount of rotation to apply to this child component along
	// the axis specified by RotationAxis.
	RotationDegrees float32

	// Scale is the scaling vector for the child component in the component.
	Scale mgl.Vec3
}

// Material defines the visual appearance of the component.
type Material struct {
	// ShaderName is the name of the shader program to use for rendering.
	ShaderName string

	// Diffuse is the base color for the material.
	Diffuse mgl.Vec4

	// Specular is the highlight color for the material.
	Specular mgl.Vec4

	// Shininess is how shiny the material is.
	// Setting to 0 removes the specular effect.
	Shininess float32

	// GenerateMipmaps indicates if mipmaps should be generated for the textures getting loaded.
	GenerateMipmaps bool

	// DiffuseTexture is the relative file path for the diffuse texture.
	DiffuseTexture string

	// NormalsTexture is the relative file path for the normal map texture.
	NormalsTexture string

	// SpecularTexture is the relative file path for the specular map texture.
	SpecularTexture string

	// Textures specifies the texture files to load for mesh, relative
	// to the component file. They will be found to RenderableCore
	// Tex* properties in order defined.
	Textures []string
}

const (
	// ColliderTypeAABB is for axis aligned bounding box colliders.
	ColliderTypeAABB = 0

	// ColliderTypeSphere is for sphere colliders.
	ColliderTypeSphere = 1

	// ColliderTypeCount is the number of collider types supported.
	ColliderTypeCount = 2
)

// CollisionRef specifies a collision object within the component
// (e.g. a collision cube for a wall). It acts as a kind of union
// structure for different collider properties.
type CollisionRef struct {
	// Type is the type of collider from the enum above (e.g. ColliderTypeAABB, etc...).
	Type int8

	// Min is the minimum point for AABB type colliders.
	Min mgl.Vec3

	// Max is the maximum point for AABB type colliders.
	Max mgl.Vec3

	// Radius is the size of the Sphere type of collider.
	Radius float32

	// Offset is used as the offset for Sphere and AABB types of colliders.
	Offset mgl.Vec3

	// Tags is a way to create 'layers' of colliders so that client code
	// can select whether or not to attempt collision against this object.
	Tags []string
}

// Component is the main structure that defines a component and also defines
// what fields to use in component JSON files.
type Component struct {
	// Name is the name of the component.
	Name string

	// Offset is the location of the component within the context of the
	// parent. If the Component has no parent, this can be the offset in the
	// world. If the Component does have a parent Component, then this is the
	// offset distance fromt he parent location.
	Offset mgl.Vec3

	// Meshes is a slice of the meshes that are parts of this component.
	Meshes []*Mesh

	// ChildReferences can be specified to include other components
	// to be contained in this component.
	ChildReferences []*ChildRef

	// Collision objects for the component.
	Collisions []*CollisionRef

	// Properties is a map for client code's custom properties for the component.
	Properties map[string]string

	// dirPath is the directory path for the component file if it was loaded
	// from JSON.
	dirPath string

	// filename is the filename without the directory path for the component file.
	filename string

	// filePath is the full filepath used to load the component file.
	filePath string

	// cachedRenderable is the cached renerable object for the component that can
	// be used as a prototype.
	cachedRenderable *fizzle.Renderable
}

// GetDirPath returns the directory path for the component if it was loaded by filepath.
func (c *Component) GetDirPath() string {
	return c.dirPath
}

// GetFilename returns the filename only part path for the component if it was loaded by filepath.
func (c *Component) GetFilename() string {
	return c.filename
}

// GetFilepath returns the full filepath the component was loaded from.
func (c *Component) GetFilepath() string {
	return c.filePath
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
	clone.Offset = c.Offset
	clone.Meshes = c.Meshes
	clone.ChildReferences = c.ChildReferences
	clone.Collisions = c.Collisions
	clone.Properties = c.Properties
	clone.dirPath = c.dirPath
	clone.filename = c.filename
	clone.filePath = c.filePath
	clone.cachedRenderable = c.cachedRenderable

	return clone
}

// SetRenderable sets the cached renderable to the one passed in as a parameter,
// calling Destroy() on the already exisiting cached Renderable.
func (c *Component) SetRenderable(newRenderable *fizzle.Renderable) {
	// destroy the old one if it exists
	if c.cachedRenderable != nil {
		c.cachedRenderable.Destroy()
	}

	// all hail the new renderable
	c.cachedRenderable = newRenderable
}

// GetRenderable will return the cached renderable object for the component
// or create one if it hasn't been made yet. The TextureManager is needed
// to resolve texture references and the shaders collection is needed to
// set a RenderShader identified by the name defined in Component.
//
// NOTE: This is not an instance of the renderable, but the main renderable
// object for the component.
func (c *Component) GetRenderable(tm *fizzle.TextureManager, shaders map[string]*fizzle.RenderShader) (*fizzle.Renderable, error) {
	// see if we have a cached renderable already created
	if c.cachedRenderable != nil {
		return c.cachedRenderable, nil
	}

	// start by creating a renderable to hold all of the meshes
	group := fizzle.NewRenderable()
	group.IsGroup = true
	group.Location = c.Offset

	// now create renderables for all of the meshes.
	// comnponents only create new render nodes for the meshs defined and
	// not for referenced components
	for _, compMesh := range c.Meshes {
		cmRenderable, err := CreateRenderableForMesh(tm, shaders, compMesh)
		if err != nil {
			return nil, err
		}
		group.AddChild(cmRenderable)
	}

	// cache it for later
	c.cachedRenderable = group

	return group, nil
}

// GetFullBinFilePath returns the full file path for the mesh binary file (gombz format).
func (cm *Mesh) GetFullBinFilePath() string {
	return cm.Parent.dirPath + cm.BinFile
}

// GetFullTexturePath returns the full file path for the mesh texture. The textureIndex
// is an index into Mesh.Textures to pull the texture name to build the path for.
func (cm *Mesh) GetFullTexturePath(textureIndex int) string {
	return cm.Parent.dirPath + cm.Material.Textures[textureIndex]
}

// GetVertices returns the vector slice containing the vertices for the mesh from
// the cached source gombz structure.
func (cm *Mesh) GetVertices() ([]mgl.Vec3, error) {
	if cm.SrcMesh == nil {
		return nil, fmt.Errorf("no internal data present for component mesh to get vertices from")
	}
	return cm.SrcMesh.Vertices, nil
}

// CreateRenderableForMesh does the work of creating the Renderable and putting all of
// the mesh data into VBOs. This also creates a new material for the renderable
// and assigns the textures accordingly.
func CreateRenderableForMesh(tm *fizzle.TextureManager, shaders map[string]*fizzle.RenderShader, compMesh *Mesh) (*fizzle.Renderable, error) {
	// create the new renderable
	r := fizzle.CreateFromGombz(compMesh.SrcMesh)
	r.Material = fizzle.NewMaterial()
	r.Location = compMesh.Offset

	// if a scale is set, copy it over to the renderable
	if compMesh.Scale[0] != 0.0 || compMesh.Scale[1] != 0.0 || compMesh.Scale[2] != 0.0 {
		r.Scale = compMesh.Scale
	}

	// Create a quaternion for the local rotation
	r.LocalRotation = mgl.AnglesToQuat(
		mgl.DegToRad(compMesh.RotationAxis[0]),
		mgl.DegToRad(compMesh.RotationAxis[1]),
		mgl.DegToRad(compMesh.RotationAxis[2]),
		compMesh.RotationOrder)

	// assign the textures
	var okay bool
	textureCount := len(compMesh.Material.Textures)
	for i := 0; i < textureCount; i++ {
		r.Material.CustomTex[i], okay = tm.GetTexture(compMesh.Material.Textures[i])
		if !okay {
			return nil, fmt.Errorf("createRenderableForMesh failed to assign a texture gl id for %s", compMesh.Material.Textures[i])
		}
		if compMesh.Material.GenerateMipmaps {
			fizzle.GenerateMipmaps(r.Material.CustomTex[i])
		}
	}
	if len(compMesh.Material.DiffuseTexture) > 0 {
		r.Material.DiffuseTex, okay = tm.GetTexture(compMesh.Material.DiffuseTexture)
		if !okay {
			return nil, fmt.Errorf("createRenderableForMesh failed to assign a texture gl id for %s", compMesh.Material.DiffuseTexture)
		}
		if compMesh.Material.GenerateMipmaps {
			fizzle.GenerateMipmaps(r.Material.DiffuseTex)
		}
	}
	if len(compMesh.Material.NormalsTexture) > 0 {
		r.Material.NormalsTex, okay = tm.GetTexture(compMesh.Material.NormalsTexture)
		if !okay {
			return nil, fmt.Errorf("createRenderableForMesh failed to assign a texture gl id for %s", compMesh.Material.NormalsTexture)
		}
		if compMesh.Material.GenerateMipmaps {
			fizzle.GenerateMipmaps(r.Material.NormalsTex)
		}
	}
	if len(compMesh.Material.SpecularTexture) > 0 {
		r.Material.SpecularTex, okay = tm.GetTexture(compMesh.Material.SpecularTexture)
		if !okay {
			return nil, fmt.Errorf("createRenderableForMesh failed to assign a texture gl id for %s", compMesh.Material.SpecularTexture)
		}
		if compMesh.Material.GenerateMipmaps {
			fizzle.GenerateMipmaps(r.Material.SpecularTex)
		}
	}

	// assign material properties if specified
	r.Material.DiffuseColor = compMesh.Material.Diffuse
	r.Material.SpecularColor = compMesh.Material.Specular
	r.Material.Shininess = compMesh.Material.Shininess
	loadedShader, okay := shaders[compMesh.Material.ShaderName]
	if okay {
		r.Material.Shader = loadedShader
	}

	return r, nil
}
