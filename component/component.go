// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package component

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/gombz"
	"github.com/tbogdala/groggy"
)

// ComponentMesh defines a mesh reference for a component and everything
// needed to draw it.
type ComponentMesh struct {
	// The material describes visual attirbutes of the component.
	Material ComponentMaterial

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

	// RotationAxis is the axis by which to rotate the mesh around; this
	// is only valid if RotationDegrees is non-zero.
	RotationAxis mgl.Vec3

	// RotationDegrees is the amount of rotation to apply to this mesh along
	// the axis specified by RotationAxis.
	RotationDegrees float32

	// Parent is the owning Component object, if any.
	Parent *Component `json:"-"`

	// SrcMesh is the cached mesh data either from BinFile.
	SrcMesh *gombz.Mesh `json:"-"`
}

// NewComponentMesh creates a new ComponentMesh object with sane defaults.
func NewComponentMesh() *ComponentMesh {
	cm := new(ComponentMesh)
	cm.Scale = mgl.Vec3{1, 1, 1}
	cm.Material.Diffuse = mgl.Vec4{1, 1, 1, 1}
	cm.Material.Specular = mgl.Vec4{1, 1, 1, 1}
	cm.Material.GenerateMipmaps = true
	return cm
}

// ComponentChildRef defines a reference to another component JSON file
// so that Components can be built from other Component parts.
type ComponentChildRef struct {
	File     string
	Location mgl.Vec3
}

// ComponentMaterial defines the visual appearance of the component.
type ComponentMaterial struct {
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
)

// CollisionRef specifies a collision object within the component
// (e.g. a collision cube for a wall). It acts as a kind of union
// structure for different collider properties.
type CollisionRef struct {
	// Type is the type of collider from the enum above (e.g. ColliderTypeAABB, etc...).
	Type uint8

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

	// Location is the location of the component in world-space coordinates.
	// This can be viewed as a kind-of default value.
	Location mgl.Vec3

	// Meshes is a slice of the meshes that are parts of this component.
	Meshes []*ComponentMesh

	// ChildReferences can be specified to include other components
	// to be contained in this component.
	ChildReferences []*ComponentChildRef

	// Collision objects for the component; currently the fizzle library doesn't
	// do anything specific with this and choice of collision library is left to
	// the user.
	Collisions []*CollisionRef

	// Properties is a map for client code's custom properties for the component.
	Properties map[string]string

	// componentDirPath is the directory path for the component file if it was loaded
	// from JSON.
	componentDirPath string

	// cachedRenderable is the cached renerable object for the component that can
	// be used as a prototype.
	cachedRenderable *fizzle.Renderable
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
	clone.componentDirPath = c.componentDirPath
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
func (c *Component) GetRenderable(tm *fizzle.TextureManager, shaders map[string]*fizzle.RenderShader) *fizzle.Renderable {
	// see if we have a cached renderable already created
	if c.cachedRenderable != nil {
		return c.cachedRenderable
	}

	// start by creating a renderable to hold all of the meshes
	group := fizzle.NewRenderable()
	group.IsGroup = true
	group.Location = c.Location

	// now create renderables for all of the meshes.
	// comnponents only create new render nodes for the meshs defined and
	// not for referenced components
	for _, compMesh := range c.Meshes {
		cmRenderable := CreateRenderableForMesh(tm, shaders, compMesh)
		group.AddChild(cmRenderable)

		// cache it for later
		c.cachedRenderable = cmRenderable
	}

	return group
}

// GetFullBinFilePath returns the full file path for the mesh binary file (gombz format).
func (cm *ComponentMesh) GetFullBinFilePath() string {
	return cm.Parent.componentDirPath + cm.BinFile
}

// GetFullTexturePath returns the full file path for the mesh texture. The textureIndex
// is an index into ComponentMesh.Textures to pull the texture name to build the path for.
func (cm *ComponentMesh) GetFullTexturePath(textureIndex int) string {
	return cm.Parent.componentDirPath + cm.Material.Textures[textureIndex]
}

// GetVertices returns the vector slice containing the vertices for the mesh from
// the cached source gombz structure.
func (cm *ComponentMesh) GetVertices() ([]mgl.Vec3, error) {
	if cm.SrcMesh == nil {
		return nil, fmt.Errorf("No internal data present for component mesh to get vertices from.")
	}
	return cm.SrcMesh.Vertices, nil
}

// CreateRenderableForMesh does the work of creating the Renderable and putting all of
// the mesh data into VBOs.
func CreateRenderableForMesh(tm *fizzle.TextureManager, shaders map[string]*fizzle.RenderShader, compMesh *ComponentMesh) *fizzle.Renderable {
	// create the new renderable
	r := fizzle.CreateFromGombz(compMesh.SrcMesh)

	// if a scale is set, copy it over to the renderable
	if compMesh.Scale[0] != 0.0 || compMesh.Scale[1] != 0.0 || compMesh.Scale[2] != 0.0 {
		r.Scale = compMesh.Scale
	}

	// Create a quaternion if rotation parameters are set
	if compMesh.RotationDegrees != 0.0 {
		r.Rotation = mgl.QuatRotate(mgl.DegToRad(compMesh.RotationDegrees), compMesh.RotationAxis)
	}

	// assign the textures
	var okay bool
	textureCount := len(compMesh.Material.Textures)
	for i := 0; i < textureCount; i++ {
		r.Core.Tex[i], okay = tm.GetTexture(compMesh.Material.Textures[i])
		if !okay {
			groggy.Logsf("ERROR", "createRenderableForMesh failed to assign a texture gl id for %s.", compMesh.Material.Textures[i])
		}
		if compMesh.Material.GenerateMipmaps {
			fizzle.GenerateMipmaps(r.Core.Tex[i])
		}
	}

	// assign material properties if specified
	r.Core.DiffuseColor = compMesh.Material.Diffuse
	r.Core.SpecularColor = compMesh.Material.Specular
	r.Core.Shininess = compMesh.Material.Shininess
	loadedShader, okay := shaders[compMesh.Material.ShaderName]
	if okay {
		r.Core.Shader = loadedShader
	}

	return r
}
