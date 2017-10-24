// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

/*

Package component consists of a Manager type that can
load component files defined in JSON so that application assets
can be defined outside of the binary.

Once a Component is loaded it can be used as a prototype to clone
new Renderable instances from so that multiple objects can be
rendered using the same OpenGL buffers to define model data.

*/
package component

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/gombz"
)

// Manager loads and manages access to Component objects.
// Component files are defined in JSON notation which is a serialized
// version of Component.
type Manager struct {
	// storage is the main collection of Component objects indexed
	// by a user-specified name.
	storage map[string]*Component

	// textureManager is a cached reference to the texture manager.
	// This will be used while loading Components and making
	// Renderables for components to load and get references to
	// textures.
	textureManager *fizzle.TextureManager

	// loadedShaders is a collection of shader programs indexed by
	// a user-specified string. Individual Components can reference
	// these shaders by name and upon Renderable construction, the
	// correct shader will be set.
	loadedShaders map[string]*fizzle.RenderShader
}

// NewManager creates a new Manager object using the
// the texture manager and shader collection specified.
func NewManager(tm *fizzle.TextureManager, shaders map[string]*fizzle.RenderShader) *Manager {
	cm := new(Manager)
	cm.storage = make(map[string]*Component)
	cm.textureManager = tm
	cm.loadedShaders = shaders
	return cm
}

// Destroy will destroy all of the contained Component objects and
// reset the component storage map.
func (cm *Manager) Destroy() {
	for _, c := range cm.storage {
		c.Destroy()
	}
	cm.storage = make(map[string]*Component)
}

// AddComponent adds a new component to the collection. If one existed previous using
// the same name, then it is overwritten.
func (cm *Manager) AddComponent(name string, component *Component) {
	cm.storage[name] = component
}

// MapComponents will call the supplied function for each component in the map.
func (cm *Manager) MapComponents(foo func(component *Component)) {
	for _, c := range cm.storage {
		foo(c)
	}
}

// GetComponentCount returns the number of components stored by the manager.
func (cm *Manager) GetComponentCount() int {
	return len(cm.storage)
}

// GetComponent returns a component from storage that matches the name specified.
// A bool is returned as the second value to indicate whether or not the component
// was found in storage.
func (cm *Manager) GetComponent(name string) (*Component, bool) {
	crComponent, okay := cm.storage[name]
	return crComponent, okay
}

// GetComponentByFilepath returns a component from storage that matches the full
// filepath specified. A non-nil value is returned if a component was found.
func (cm *Manager) GetComponentByFilepath(filepath string) *Component {
	for _, comp := range cm.storage {
		if comp.filePath == filepath {
			return comp
		}
	}

	return nil
}

// GetRenderableInstance gets the renderable from the component and clones it to
// a new instance. It then loops over all child references and calls GetRenderableInstance
// for all of them, creating new clones for each, recursively.
func (cm *Manager) GetRenderableInstance(component *Component) (*fizzle.Renderable, error) {
	compRenderable, err := component.GetRenderable(cm.textureManager, cm.loadedShaders)
	if err != nil {
		return nil, err
	}
	r := compRenderable.Clone()

	// clone a renderable for each of the child references
	for _, cref := range component.ChildReferences {
		_, childFileName := filepath.Split(cref.File)
		crComponent, okay := cm.GetComponent(childFileName)
		if !okay {
			return nil, fmt.Errorf("in GetRenderableInstance(), component %s has a ChildInstance (%s) that wasn't loaded", component.Name, cref.File)
		}

		rc, err := cm.GetRenderableInstance(crComponent)
		if err != nil {
			return nil, err
		}

		// override the location for the renderable if location was specified in
		// the child reference
		rc.Location[0] = cref.Location[0]
		rc.Location[1] = cref.Location[1]
		rc.Location[2] = cref.Location[2]

		r.AddChild(rc)
	}

	return r, nil
}

// LoadComponentFromFile loads a component from a JSON file and stores it under
// the name speicified in the component file. This function returns the new component
// and a non-nil error on failure.
func (cm *Manager) LoadComponentFromFile(filename string) (*Component, error) {
	// make sure the component file exists
	jsonBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read the component file specified: %v", err)
	}

	return cm.LoadComponentFromBytes(jsonBytes, filename)
}

// LoadComponentFromBytes loads the component from a JSON byte slice and stores it
// under the component name. If the component has already been loaded, then it will
// return the previously existing component. This function returns the new component and
// a non-nil error value on failure.
func (cm *Manager) LoadComponentFromBytes(jsonBytes []byte, filename string) (*Component, error) {
	// attempt to decode the json
	component := new(Component)
	err := json.Unmarshal(jsonBytes, component)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the JSON in the component file specified: %v", err)
	}

	// now that we have it decoded, make sure that it doesn't already exist
	// in the collection of loaded components; if it does, just return the
	// already created Component.
	foundChild, found := cm.storage[component.Name]
	if found {
		return foundChild, nil
	}

	// split the directory path to the component file
	componentDirPath, componentFilename := filepath.Split(filename)
	component.dirPath = componentDirPath
	component.filename = componentFilename
	component.filePath = filename

	// load all of the meshes in the component
	for _, compMesh := range component.Meshes {
		err = loadMeshForComponent(component, compMesh)
		if err != nil {
			return nil, err
		}
	}

	// load the associated textures
	for meshIndex, compMesh := range component.Meshes {
		for i := range compMesh.Material.Textures {
			_, err = cm.textureManager.LoadTexture(compMesh.Material.Textures[i], compMesh.GetFullTexturePath(i))
			if err != nil {
				return nil, fmt.Errorf("mesh #%d failed to load texture: %s", meshIndex, compMesh.Material.Textures[i])
			}
		}
		if len(compMesh.Material.DiffuseTexture) > 0 {
			_, err = cm.textureManager.LoadTexture(compMesh.Material.DiffuseTexture, compMesh.Parent.dirPath+compMesh.Material.DiffuseTexture)
			if err != nil {
				return nil, fmt.Errorf("mesh #%d failed to load diffuse texture: %s", meshIndex, compMesh.Material.DiffuseTexture)
			}
		}
		if len(compMesh.Material.NormalsTexture) > 0 {
			_, err = cm.textureManager.LoadTexture(compMesh.Material.NormalsTexture, compMesh.Parent.dirPath+compMesh.Material.NormalsTexture)
			if err != nil {
				return nil, fmt.Errorf("mesh #%d failed to load normal map texture: %s", meshIndex, compMesh.Material.NormalsTexture)
			}
		}
		if len(compMesh.Material.SpecularTexture) > 0 {
			_, err = cm.textureManager.LoadTexture(compMesh.Material.SpecularTexture, compMesh.Parent.dirPath+compMesh.Material.SpecularTexture)
			if err != nil {
				return nil, fmt.Errorf("mesh #%d failed to load specular map texture: %s", meshIndex, compMesh.Material.SpecularTexture)
			}
		}
	}

	// place the new component into storage before parsing children
	// to avoid a possible infinite loop
	cm.storage[component.Name] = component

	// For all of the child references, see if we have a component loaded
	// for it already. If not, then load those components too.
	for _, childRef := range component.ChildReferences {
		foundChild := cm.GetComponentByFilepath(childRef.File)
		if foundChild != nil {
			continue
		}

		_, err := cm.LoadComponentFromFile(componentDirPath + childRef.File)
		if err != nil {
			return nil, fmt.Errorf("component %s has a ChildInstance (%s) could not be loaded: %v", component.Name, childRef.File, err)
		}
	}

	return component, nil
}

func loadMeshForComponent(component *Component, compMesh *Mesh) error {
	// setup a pointer back to the parent
	compMesh.Parent = component

	if len(compMesh.BinFile) > 0 {
		binBytes, err := ioutil.ReadFile(compMesh.GetFullBinFilePath())
		if err != nil {
			return fmt.Errorf("failed to load the binary file (%s) for the ComponentMesh: %v", compMesh.BinFile, err)
		}

		// load the mesh from the binary file
		compMesh.SrcMesh, err = gombz.DecodeMesh(binBytes)
		if err != nil {
			return fmt.Errorf("failed to deocde the binary file (%s) for the ComponentMesh: %v", compMesh.BinFile, err)
		}
	}

	return nil
}
