// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package component

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/gombz"
	"github.com/tbogdala/groggy"
)

type ComponentManager struct {
	storage        map[string]*Component
	textureManager *fizzle.TextureManager
	loadedShaders  map[string]*fizzle.RenderShader
}

func NewComponentManager(tm *fizzle.TextureManager, shaders map[string]*fizzle.RenderShader) *ComponentManager {
	cm := new(ComponentManager)
	cm.storage = make(map[string]*Component)
	cm.textureManager = tm
	cm.loadedShaders = shaders
	return cm
}

// Destroy will destroy all of the contained Component objects and
// reset the storage map.
func (cm *ComponentManager) Destroy() {
	for _, c := range cm.storage {
		c.Destroy()
	}
	cm.storage = make(map[string]*Component)
}

// AddComponent adds a new component to the collection. If one existed previous using
// the same name, then it is overwritten.
func (cm *ComponentManager) AddComponent(name string, component *Component) {
	cm.storage[name] = component
}

// GetComponent returns a component from storage that matches the name specified.
// A bool is returned as the second value to indicate whether or not the component
// was found in storage.
func (cm *ComponentManager) GetComponent(name string) (*Component, bool) {
	crComponent, okay := cm.storage[name]
	return crComponent, okay
}

// GetRenderableInstance gets the renderable from the component and clones it to
// a new instance. It then loops over all child references and calls GetRenderableInstance
// for all of them, creating new clones for each, recursively.
func (cm *ComponentManager) GetRenderableInstance(component *Component) *fizzle.Renderable {
	compRenderable := component.GetRenderable(cm.textureManager, cm.loadedShaders)
	r := compRenderable.Clone()

	// clone a renderable for each of the child references
	for _, cref := range component.ChildReferences {
		_, childFileName := filepath.Split(cref.File)
		crComponent, okay := cm.GetComponent(childFileName)
		if !okay {
			groggy.Logsf("ERROR", "GetRenderableInstance: Component %s has a ChildInstance (%s) that wasn't loaded.\n", component.Name, cref.File)
			continue
		}

		rc := cm.GetRenderableInstance(crComponent)

		// override the location for the renderable if location was specified in
		// the child reference
		rc.Location[0] = cref.Location[0]
		rc.Location[1] = cref.Location[1]
		rc.Location[2] = cref.Location[2]

		r.AddChild(rc)
	}

	return r
}

// LoadComponentFromFile loads a component from a JSON file and stores it under
// the name speicified. This function returns the new component and a possible
// error value.
func (cm *ComponentManager) LoadComponentFromFile(filename string, storageName string) (*Component, error) {
	// split the directory path to the component file
	componentDirPath, _ := filepath.Split(filename)

	// check to see if it exists in storage already
	if loadedComp, okay := cm.storage[storageName]; okay {
		return loadedComp, nil
	}

	// make sure the component file exists
	jsonBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the component file specified.\n%s\n", err)
	}

	return cm.LoadComponentFromBytes(jsonBytes, storageName, componentDirPath)
}

// LoadComponentFromBytes loads the component from a JSON byte slice and stores it
// under the name specified. componentDirPath can be set to aid the finding of
// parts of the component to load. This function returns the new component and
// a possible error value.
func (cm *ComponentManager) LoadComponentFromBytes(jsonBytes []byte, storageName string, componentDirPath string) (*Component, error) {
	// attempt to decode the json
	component := new(Component)
	err := json.Unmarshal(jsonBytes, component)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode the JSON in the component file specified.\n%s\n", err)
	}

	// store the directory path to the component file
	component.componentDirPath = componentDirPath

	// load all of the meshes in the component
	for _, compMesh := range component.Meshes {
		err = loadMeshForComponent(component, compMesh)
		if err != nil {
			return nil, err
		}
	}

	// load the associated textures
	for meshIndex, compMesh := range component.Meshes {
		for i := range compMesh.Textures {
			_, err = cm.textureManager.LoadTexture(compMesh.Textures[i], compMesh.GetFullTexturePath(i))
			if err != nil {
				groggy.Logsf("ERROR", "Mesh #%d failed to load texture: %s", meshIndex, compMesh.Textures[i])
			} else {
				groggy.Logsf("DEBUG", "Mesh #%d loaded texture: %s", meshIndex, compMesh.Textures[i])
			}
		}
	}

	// place the new component into storage before parsing children
	// to avoid a possible infinite loop
	cm.storage[storageName] = component

	// For all of the child references, see if we have a component loaded
	// for it already. If not, then load those components too.
	for _, childRef := range component.ChildReferences {
		_, childFileName := filepath.Split(childRef.File)
		if _, okay := cm.storage[childFileName]; okay {
			continue
		}

		_, err := cm.LoadComponentFromFile(componentDirPath+childRef.File, storageName)
		if err != nil {
			groggy.Logsf("ERROR", "Component %s has a ChildInstance (%s) could not be loaded.\n%v", component.Name, childRef.File, err)
		}
	}

	groggy.Logsf("DEBUG", "Component \"%s\" has been loaded", component.Name)
	return component, nil
}

func loadMeshForComponent(component *Component, compMesh *ComponentMesh) error {
	// setup a pointer back to the parent
	compMesh.Parent = component

	if len(compMesh.BinFile) > 0 {
		binBytes, err := ioutil.ReadFile(compMesh.GetFullBinFilePath())
		if err != nil {
			return fmt.Errorf("Failed to load the binary file (%s) for the ComponentMesh.\n%v\n", compMesh.BinFile, err)
		}

		// load the mesh from the binary file
		compMesh.SrcMesh, err = gombz.DecodeMesh(binBytes)
		if err != nil {
			return fmt.Errorf("Failed to deocde the binary file (%s) for the ComponentMesh.\n%v\n", compMesh.BinFile, err)
		}
	}

	return nil
}
