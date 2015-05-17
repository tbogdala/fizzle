// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/tbogdala/assimp-go"
	"github.com/tbogdala/gombz"
	"github.com/tbogdala/groggy"
)

type ComponentManager struct {
	storage        map[string]*Component
	textureManager *TextureManager
}

func NewComponentManager(tm *TextureManager) *ComponentManager {
	cm := new(ComponentManager)
	cm.storage = make(map[string]*Component)
	cm.textureManager = tm
	return cm
}

func (cm *ComponentManager) GetComponent(name string) (*Component, bool) {
	crComponent, okay := cm.storage[name]
	return crComponent, okay
}

func (cm *ComponentManager) GetRenderableInstance(component *Component) *Renderable {
	compRenderable := component.GetRenderable(cm.textureManager)
	r := compRenderable.Clone()

	// clone a renderable for each of the child references
	for _, cref := range component.ChildReferences {
		_, childFileName := filepath.Split(cref.File)
		crComponent, okay := cm.GetComponent(childFileName)
		if !okay {
			groggy.Logsf("ERROR", "GetRenderableInstance: Component %s has a ChildInstance (%s) that wasn't loaded.\n", component.Name, cref.File)
			continue
		}

		childRenderable := crComponent.GetRenderable(cm.textureManager)
		rc := childRenderable.Clone()

		// override the location for the renderable if location was specified in
		// the child reference
		rc.Location[0] = cref.Location[0]
		rc.Location[1] = cref.Location[1]
		rc.Location[2] = cref.Location[2]

		r.AddChild(rc)
	}

	return r
}

func (cm *ComponentManager) LoadComponentFromFile(filename string) (*Component, error) {
	// split the directory path to the component file
	componentDirPath, componentFileName := filepath.Split(filename)

	// check to see if it exists in storage already
	if loadedComp, okay := cm.storage[componentFileName]; okay {
		return loadedComp, nil
	}

	// make sure the component file exists
	jsonBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the component file specified.\n%s\n", err)
	}

	// attempt to decode the json
	component := new(Component)
	err = json.Unmarshal(jsonBytes, component)
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
	cm.storage[componentFileName] = component

	// For all of the child references, see if we have a component loaded
	// for it already. If not, then load those components too.
	for _, childRef := range component.ChildReferences {
		_, childFileName := filepath.Split(childRef.File)
		if _, okay := cm.storage[childFileName]; okay {
			continue
		}

		_, err := cm.LoadComponentFromFile(componentDirPath + childRef.File)
		if err != nil {
			groggy.Logsf("ERROR", "Component %s has a ChildInstance (%s) could not be loaded.\n%v", component.Name, childRef.File, err)
		}
	}

	groggy.Logsf("DEBUG", "Component \"%s\" has been loaded from %s", component.Name, filename)
	return component, nil
}

func loadMeshForComponent(component *Component, compMesh *ComponentMesh) error {
	// setup a pointer back to the parent
	compMesh.Parent = component

	// now that we have the component json figured out, see if we can load the src
	// file through assimp for all of the meshes
	var srcMeshes []*gombz.Mesh
	var err error
	if len(compMesh.SrcFile) > 0 {
		srcMeshes, err = assimp.ParseFile(compMesh.GetFullSrcFilePath())
		if err != nil {
			return fmt.Errorf("Failed to load the source file (%s) for the ComponentMesh.\n%v\n", compMesh.SrcFile, err)
		}
	}

	// TODO: Eventually instead of returning here, it will just load the binary
	// version. This logic isn't as useful in the component editor so it doesn't
	// exist yet ...
	// we return here if no meshes were returned in the parsed file
	if len(srcMeshes) < 1 {
		return nil
	}

	// send out a warning if we have more than one mesh returned from the assimp parsing.
	numOfSrcMeshes := len(srcMeshes)
	if numOfSrcMeshes > 1 {
		groggy.Logsf("ERROR", "SrcFile mesh has %d meshes. Only one mesh is supported!", numOfSrcMeshes)
	}
	compMesh.srcMesh = srcMeshes[0]

	// force a write out of the compressed binary form of the model if the BinFile is set
	if len(compMesh.BinFile) > 0 {
		// do the encode
		meshBytes, err := compMesh.srcMesh.Encode()
		if err != nil {
			groggy.Log("ERROR", "Failed to encode BinFile mesh")
		} else {
			// we've encoded, now write the file out
			err = ioutil.WriteFile(compMesh.GetFullBinFilePath(), meshBytes, os.ModePerm)
			if err != nil {
				groggy.Logsf("ERROR", "Failed to write BinFile mesh: %s", compMesh.BinFile)
			} else {
				groggy.Logsf("INFO", "Wrote BinFile mesh: %s", compMesh.BinFile)
			}
		}
	}

	return nil
}
