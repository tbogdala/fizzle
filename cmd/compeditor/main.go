// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	assimp "github.com/tbogdala/assimp-go"
	gui "github.com/tbogdala/eweygewey"
	guiinput "github.com/tbogdala/eweygewey/glfwinput"
	gombz "github.com/tbogdala/gombz"
	groggy "github.com/tbogdala/groggy"

	fizzle "github.com/tbogdala/fizzle"
	component "github.com/tbogdala/fizzle/component"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

var (
	windowWidth                = 1280
	windowHeight               = 720
	mainWindow                 *glfw.Window
	camera                     *fizzle.OrbitCamera
	uiman                      *gui.Manager
	renderer                   *forward.ForwardRenderer
	textureMan                 *fizzle.TextureManager
	colorShaderFilepath        = "../../examples/assets/forwardshaders/color"
	basicShaderFilepath        = "../../examples/assets/forwardshaders/basic"
	basicSkinnedShaderFilepath = "../../examples/assets/forwardshaders/basicSkinned"

	clearColor = gui.ColorIToV(32, 32, 32, 32)

	shaders      map[string]*fizzle.RenderShader
	componentMan *component.Manager

	visibleMeshes    map[string]*meshRenderable
	visibleColliders []*colliderRenderable
	theComponent     component.Component
	childComponents  []*component.Component

	// childRefFilenames is a map of child reference filename to component name
	childRefFilenames map[string]string

	appStartTime time.Time
	totalTime    float64
)

const (
	fontScale    = 14
	fontFilepath = "../../examples/assets/Oswald-Heavy.ttf"
	fontGlyphs   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890., :[]{}\\|<>;\"'~`?/-+_=()*&^%$#@!"

	compMeshWindowID = "ComponentMesh"

	segsInSphereWire = 32

	// ui layout constants
	textWidth = 0.2
	width3Col = 0.8 / 3.0
	width4Col = 0.2
	meshWndX  = 0.65
	meshWndY  = 0.99
)

// block of flags set on the command line
var (
	flagDesktopNumber int
	flagComponentFile string
)

// meshRenderable is used to tie together state for the component mesh,
// the renderable for this component mesh and any other state information relating.
type meshRenderable struct {
	ComponentMesh     *component.Mesh
	Renderable        *fizzle.Renderable
	AnimationsEnabled []bool
}

// colliderRenderable is used to tie together state for the component collider
// and the renderable for it.
type colliderRenderable struct {
	// Collider is a copy of the CollisionRef data used to make the renderable
	Collider component.CollisionRef

	// Renderable is the drawable wireframe representing the rendreable
	Renderable *fizzle.Renderable
}

// GLFW event handling must run on the main OS thread. If this doesn't get
// locked down, you will likely see random crashes on memory access while
// running the application after a few seconds.
//
// So on initialization of the module, lock the OS thread for this goroutine.
func init() {
	runtime.LockOSThread()
	flag.IntVar(&flagDesktopNumber, "desktop", -1, "the index of the desktop to create the main window on")
	flag.StringVar(&flagComponentFile, "cf", "component.json", "the name of the component file to load and save")
}

// guiAddDragSliderVec3 adds drag slider floats for a Vec3.
func guiAddDragSliderVec3(wnd *gui.Window, widthS float32, idPrefix string, index int, speed float32, v *mgl.Vec3) {
	wnd.RequestItemWidthMax(widthS)
	wnd.DragSliderFloat(fmt.Sprintf("%s%d_0", idPrefix, index), speed, &v[0])
	wnd.RequestItemWidthMax(widthS)
	wnd.DragSliderFloat(fmt.Sprintf("%s%d_1", idPrefix, index), speed, &v[1])
	wnd.RequestItemWidthMax(widthS)
	wnd.DragSliderFloat(fmt.Sprintf("%s%d_2", idPrefix, index), speed, &v[2])
}

// guiAddSliderVec4 adds slider floats for a Vec4.
func guiAddSliderVec4(wnd *gui.Window, widthS float32, idPrefix string, index int, v *mgl.Vec4, min, max float32) {
	wnd.RequestItemWidthMax(widthS)
	wnd.SliderFloat(fmt.Sprintf("%s%d_0", idPrefix, index), &v[0], min, max)
	wnd.RequestItemWidthMax(widthS)
	wnd.SliderFloat(fmt.Sprintf("%s%d_1", idPrefix, index), &v[1], min, max)
	wnd.RequestItemWidthMax(widthS)
	wnd.SliderFloat(fmt.Sprintf("%s%d_2", idPrefix, index), &v[2], min, max)
	wnd.RequestItemWidthMax(widthS)
	wnd.SliderFloat(fmt.Sprintf("%s%d_3", idPrefix, index), &v[3], min, max)
}

// getLoadedChildComponent uses the global childRefFilenames map to look up a
// component name for a given child reference name and then find that component
// in the loaded child components slice. Returns nil if no match is found.
func getLoadedChildComponent(loadedChildComponents []*component.Component, refFileName string) *component.Component {
	compNameToFind, okay := childRefFilenames[refFileName]
	if !okay {
		return nil
	}

	for _, childComp := range loadedChildComponents {
		if childComp.Name == compNameToFind {
			return childComp
		}
	}

	return nil
}

func makeRenderableForMesh(compMesh *component.Mesh) *fizzle.Renderable {
	prefixDir := getComponentPrefix()

	// attempt to load the mesh from the source file if one is specified
	if compMesh.SrcFile != "" {
		meshFilepath := prefixDir + compMesh.SrcFile
		srcMeshes, parseErr := assimp.ParseFile(meshFilepath)
		if parseErr != nil {
			fmt.Printf("Failed to load source mesh %s: %v\n", meshFilepath, parseErr)
		} else {
			if len(srcMeshes) > 0 {
				compMesh.SrcMesh = srcMeshes[0]
				fmt.Printf("Loaded source mesh: %s\n", compMesh.SrcFile)
			}
		}
	} else if compMesh.BinFile != "" {
		gombzFilepath := prefixDir + compMesh.BinFile
		gombzBytes, err := ioutil.ReadFile(gombzFilepath)
		if err != nil {
			fmt.Printf("Failed to load Gombz bytes from %s: %v\n", gombzFilepath, err)
		} else {
			compMesh.SrcMesh, err = gombz.DecodeMesh(gombzBytes)
			if err != nil {
				fmt.Printf("Failed to decode Gombz mesh from %s: %v\n", gombzFilepath, err)
			} else {
				fmt.Printf("Loaded gombz mesh: %s\n", compMesh.SrcFile)
			}
		}
	}

	// if we haven't loaded something by now, then return a nil renderable
	if compMesh.SrcMesh == nil {
		return nil
	}

	compRenderable := new(meshRenderable)
	r := fizzle.CreateFromGombz(compMesh.SrcMesh)
	r.Material = fizzle.NewMaterial()
	r.Material.Shader = shaders["BasicSkinned"]
	r.Location = compMesh.Offset
	r.Scale = compMesh.Scale

	// Create a quaternion if rotation parameters are set
	if compMesh.RotationDegrees != 0.0 {
		r.LocalRotation = mgl.QuatRotate(mgl.DegToRad(compMesh.RotationDegrees), compMesh.RotationAxis)
	}

	// store the new renderable with the component mesh it belongs to
	compRenderable.ComponentMesh = compMesh
	compRenderable.Renderable = r

	// setup the animation enable flag slice
	compRenderable.AnimationsEnabled = []bool{}
	for i := 0; i < len(compMesh.SrcMesh.Animations); i++ {
		compRenderable.AnimationsEnabled = append(compRenderable.AnimationsEnabled, false)
	}

	visibleMeshes[compMesh.Name] = compRenderable
	return r
}

// doSaveGombz saves a component Mesh out to a gombz file at the
// location specified by BinFile.
func doSaveGombz(compMesh *component.Mesh) error {
	gombzBytes, err := compMesh.SrcMesh.Encode()
	if err != nil {
		return fmt.Errorf("Error while serializing Gombz mesh: %v", err)
	}

	prefixDir := getComponentPrefix()
	gombzFilepath := prefixDir + compMesh.BinFile
	err = ioutil.WriteFile(gombzFilepath, gombzBytes, 0744)
	if err != nil {
		return fmt.Errorf("Error while writing Gombz file: %v", err)
	}

	fmt.Printf("Wrote Gombz file: %s\n", gombzFilepath)
	return nil
}

// doLoadTexture loads a relative filepath texture into the
// texture manager.
func doLoadTexture(texFile string) error {
	prefixDir := getComponentPrefix()
	texFilepath := prefixDir + texFile
	_, err := textureMan.LoadTexture(texFile, texFilepath)
	if err != nil {
		return fmt.Errorf("Failed to load texture %s: %v", texFile, err)
	}

	fmt.Printf("Loaded texture: %s\n", texFile)
	return nil
}

func doDeleteTexture(texIndex int, matTextures []string) []string {
	if texIndex == 0 && len(matTextures) == 1 {
		return []string{}
	} else if texIndex == 0 {
		return matTextures[1:]
	} else if texIndex == len(matTextures)-1 {
		return matTextures[:texIndex]
	}
	return append(matTextures[:texIndex], matTextures[texIndex+1:]...)
}

func doAnimation(animation *gombz.Animation, renderable *fizzle.Renderable, totalTime float64) {
	aniTime := float32(math.Mod(totalTime*float64(animation.TicksPerSecond), float64(animation.Duration)))
	renderable.Core.Skeleton.Animate(animation, aniTime)
}

// getComponentPrefix gets the prefix directory for the current component filename.
func getComponentPrefix() string {
	prefixDir := ""
	if len(flagComponentFile) > 0 {
		prefixDir, _ = filepath.Split(flagComponentFile)
	}
	return prefixDir
}

// loadAllReferenceTextures loads any referenced textures in the Mesh's material.
func loadAllReferenceTextures(compMesh *component.Mesh) {
	for _, texFile := range compMesh.Material.Textures {
		doLoadTexture(texFile)
	}
	if len(compMesh.Material.DiffuseTexture) > 0 {
		doLoadTexture(compMesh.Material.DiffuseTexture)
	}
	if len(compMesh.Material.NormalsTexture) > 0 {
		doLoadTexture(compMesh.Material.NormalsTexture)
	}
	if len(compMesh.Material.SpecularTexture) > 0 {
		doLoadTexture(compMesh.Material.SpecularTexture)
	}
}

func doLoadComponentFile(componentFilepath string) {
	existingCompJSON, err := ioutil.ReadFile(componentFilepath)
	if err == nil {
		err := json.Unmarshal(existingCompJSON, &theComponent)
		if err != nil {
			fmt.Printf("Failed to load component %s: %v\n", componentFilepath, err)
		} else {
			fmt.Printf("Loaded component: %s\n", componentFilepath)

			// destroy all existing renderables
			for _, r := range visibleMeshes {
				r.Renderable.Destroy()
				r.Renderable = nil
			}
			for _, vc := range visibleColliders {
				vc.Renderable.Destroy()
				vc.Renderable = nil
			}
			visibleMeshes = make(map[string]*meshRenderable)
			visibleColliders = make([]*colliderRenderable, 0)

			// open windows for all existing meshes
			screenX := float32(meshWndX)
			screenY := float32(meshWndY)
			for _, compMesh := range theComponent.Meshes {
				loadAllReferenceTextures(compMesh)
				createMeshWindow(compMesh, screenX, screenY)
				if compMesh.SrcFile != "" {
					makeRenderableForMesh(compMesh)
				}
				screenX += 0.05
				screenY -= 0.05
			}
		}
	}
}

// doSaveComponent saves the component to a file.
func doSaveComponent(comp *component.Component, filepath string) error {
	compJSON, jsonErr := json.MarshalIndent(comp, "", "    ")
	if jsonErr == nil {
		fileErr := ioutil.WriteFile(filepath, compJSON, 0744)
		if fileErr != nil {
			return fmt.Errorf("Failed to write component: %v\n", fileErr)
		}
	} else {
		return fmt.Errorf("Failed to serialize component to JSON: %v\n", jsonErr)
	}

	return nil
}

// doAddChildReference adds a new child component reference.
func doAddChildReference(comp *component.Component) {
	newChildRef := new(component.ChildRef)
	newChildRef.Scale = mgl.Vec3{1, 1, 1}
	comp.ChildReferences = append(comp.ChildReferences, newChildRef)
}

// doAddCollider ends up adding a collider (defaults to sphere).
func doAddCollider(comp *component.Component) {
	newCollider := new(component.CollisionRef)
	newCollider.Type = component.ColliderTypeSphere
	newCollider.Radius = 1.0
	comp.Collisions = append(comp.Collisions, newCollider)
}

// doAddMesh adds a new mesh to the component.
func doAddMesh() {
	newCompMesh := component.NewMesh()
	newCompMesh.Name = fmt.Sprintf("Mesh %d", len(theComponent.Meshes)+1)
	theComponent.Meshes = append(theComponent.Meshes, newCompMesh)
	createMeshWindow(newCompMesh, meshWndX, meshWndY)
}

// doDeleteMesh destroys the renderable for a component mesh and then
// removes the mesh from the map of visibleMeshes.
func doDeleteMesh(componentMeshName string) {
	cr := visibleMeshes[componentMeshName]
	cr.Renderable.Destroy()
	cr.Renderable = nil
	delete(visibleMeshes, componentMeshName)
}

// doShowMeshWindow will show a mesh property window for a given Mesh
func doShowMeshWindow(compMesh *component.Mesh) {
	meshWindow := uiman.GetWindow(fmt.Sprintf("%s%s", compMeshWindowID, compMesh.Name))
	if meshWindow == nil {
		createMeshWindow(compMesh, meshWndX, meshWndY)
	}
}

// doHideMeshWindow will hide a mesh property window for a given Mesh
func doHideMeshWindow(compMesh *component.Mesh) {
	meshWindow := uiman.GetWindow(fmt.Sprintf("%s%s", compMeshWindowID, compMesh.Name))
	if meshWindow != nil {
		uiman.RemoveWindow(meshWindow)
	}
}

// doLoadComponentFile closes all of the windows with an ID that starts
// with compMeshWindowID.
func closeAllMeshWindows() {
	// remove all existing mesh windows
	meshWindows := uiman.GetWindowsByFilter(func(w *gui.Window) bool {
		if strings.HasPrefix(w.ID, compMeshWindowID) {
			return true
		}
		return false
	})

	for _, meshWnd := range meshWindows {
		uiman.RemoveWindow(meshWnd)
	}
}

func doPrevColliderType(collider *component.CollisionRef) {
	collider.Type = collider.Type - 1
	if collider.Type < 0 {
		collider.Type = component.ColliderTypeCount - 1
	}
}

func doNextColliderType(collider *component.CollisionRef) {
	collider.Type = collider.Type + 1
	if collider.Type >= component.ColliderTypeCount {
		collider.Type = 0
	}
}

// doUpdateVisibleCollider checks the visibleColliders slice at an index to see
// if the collider's renderable needs to get created or updated.
// returns a potentially new slice of []*colliderRenderable because a new
// renderable may have been added.
func doUpdateVisibleCollider(colliderRenderables []*colliderRenderable, collider *component.CollisionRef, colliderIndex int) []*colliderRenderable {
	// is the collider index within the length of renderables we have? If so, update it.
	if len(colliderRenderables) > colliderIndex {
		visCollider := colliderRenderables[colliderIndex]

		switch collider.Type {
		case component.ColliderTypeAABB:
			if !visCollider.Collider.Min.ApproxEqual(collider.Min) ||
				!visCollider.Collider.Max.ApproxEqual(collider.Max) ||
				visCollider.Collider.Type != collider.Type {
				visCollider.Collider = *collider
				visCollider.Renderable = fizzle.CreateWireframeCube(collider.Min[0], collider.Min[1], collider.Min[2],
					collider.Max[0], collider.Max[1], collider.Max[2])
			}
		case component.ColliderTypeSphere:
			if !visCollider.Collider.Offset.ApproxEqual(collider.Offset) ||
				math.Abs(float64(visCollider.Collider.Radius-collider.Radius)) > 0.01 ||
				visCollider.Collider.Type != collider.Type {
				visCollider.Collider = *collider
				visCollider.Renderable = fizzle.CreateWireframeCircle(
					collider.Offset[0], collider.Offset[1], collider.Offset[2], collider.Radius, segsInSphereWire, fizzle.X|fizzle.Y)
				visCollider.Renderable.AddChild(fizzle.CreateWireframeCircle(
					collider.Offset[0], collider.Offset[1], collider.Offset[2], collider.Radius, segsInSphereWire, fizzle.Y|fizzle.Z))
				visCollider.Renderable.AddChild(fizzle.CreateWireframeCircle(
					collider.Offset[0], collider.Offset[1], collider.Offset[2], collider.Radius, segsInSphereWire, fizzle.X|fizzle.Z))
			}
		}
	} else {
		// append a new visible collider
		visCollider := new(colliderRenderable)
		visCollider.Collider = *collider

		switch collider.Type {
		case component.ColliderTypeAABB:
			visCollider.Renderable = fizzle.CreateWireframeCube(collider.Min[0], collider.Min[1], collider.Min[2],
				collider.Max[0], collider.Max[1], collider.Max[2])
		case component.ColliderTypeSphere:
			visCollider.Renderable = fizzle.CreateWireframeCircle(
				collider.Offset[0], collider.Offset[1], collider.Offset[2], collider.Radius, segsInSphereWire, fizzle.X|fizzle.Y)
			visCollider.Renderable.AddChild(fizzle.CreateWireframeCircle(
				collider.Offset[0], collider.Offset[1], collider.Offset[2], collider.Radius, segsInSphereWire, fizzle.Y|fizzle.Z))
			visCollider.Renderable.AddChild(fizzle.CreateWireframeCircle(
				collider.Offset[0], collider.Offset[1], collider.Offset[2], collider.Radius, segsInSphereWire, fizzle.X|fizzle.Z))
		}

		colliderRenderables = append(colliderRenderables, visCollider)
	}

	return colliderRenderables
}

// doLoadChildComponent loads a component through the global component manager.
// It returns a new slice of child components since a new one may be added if
// there is no error.
func doLoadChildComponent(childComps []*component.Component, childRef *component.ChildRef) ([]*component.Component, error) {
	prefixDir := getComponentPrefix()
	fullFilepath := prefixDir + childRef.File
	newChildComponent, err := componentMan.LoadComponentFromFile(fullFilepath, childRef.File)
	if err != nil {
		return childComps, fmt.Errorf("Failed to load child component: %s\n%v\n", fullFilepath, err)
	}

	fmt.Printf("Loaded child component: %s\n", childRef.File)
	childComps = append(childComps, newChildComponent)
	childRefFilenames[childRef.File] = newChildComponent.Name
	return childComps, nil
}

// removeStaleChildComponents remove any visible child components that no longer have a reference
func removeStaleChildComponents(childComps []*component.Component, parentComp *component.Component, refFilenames map[string]string) []*component.Component {
	childComponentsThatSurvive := []*component.Component{}
	for _, ref := range parentComp.ChildReferences {
		compNameToFind, okay := refFilenames[ref.File]
		if !okay {
			continue
		}

		for _, childCompToTest := range childComps {
			if compNameToFind == childCompToTest.Name {
				childComponentsThatSurvive = append(childComponentsThatSurvive, childCompToTest)
			}
		}
	}

	return childComponentsThatSurvive
}

var (
	meshWindowCount = 0
)

func createMeshWindow(newCompMesh *component.Mesh, screenX, screenY float32) {
	meshWindowCount++
	wndCount := meshWindowCount
	// FIXME: find a better spot to spawn potentially
	meshWnd := uiman.NewWindow(compMeshWindowID, screenX, screenY, 0.30, 0.75, func(wnd *gui.Window) {
		compRenderable := visibleMeshes[newCompMesh.Name]
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Name")
		wnd.Editbox(fmt.Sprintf("meshNameEditbox%d", wndCount), &newCompMesh.Name)

		// force the window id to be the mesh name plus a prefix
		wnd.ID = fmt.Sprintf("%s%s", compMeshWindowID, newCompMesh.Name)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Source")
		loadSource, _ := wnd.Button(fmt.Sprintf("meshLoadSrcButton%d", wndCount), "L")
		wnd.Editbox(fmt.Sprintf("meshSourceFileEditbox%d", wndCount), &newCompMesh.SrcFile)
		if loadSource {
			makeRenderableForMesh(newCompMesh)
		}

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Gombz")
		saveGombz, _ := wnd.Button(fmt.Sprintf("meshSaveBinButton%d", wndCount), "S")
		wnd.Editbox(fmt.Sprintf("meshBinaryFileEditbox%d", wndCount), &newCompMesh.BinFile)
		if saveGombz && newCompMesh.SrcMesh != nil {
			doSaveGombz(newCompMesh)
		}

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Offset")
		guiAddDragSliderVec3(wnd, width3Col, "MeshOffset", wndCount, 0.1, &newCompMesh.Offset)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Scale")
		guiAddDragSliderVec3(wnd, width3Col, "MeshScale", wndCount, 0.1, &newCompMesh.Scale)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Rotation Axis")
		guiAddDragSliderVec3(wnd, width3Col, "MeshRotationAxis", wndCount, 0.01, &newCompMesh.RotationAxis)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Rotation Degrees")
		wnd.DragSliderFloat(fmt.Sprintf("MeshRotationDegrees%d", wndCount), 0.1, &newCompMesh.RotationDegrees)

		// ------------------------------------------------
		// material settings
		wnd.Separator()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Shader")
		wnd.Editbox(fmt.Sprintf("materialShaderNameEditbox%d", wndCount), &newCompMesh.Material.ShaderName)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Diffuse")
		guiAddSliderVec4(wnd, width4Col, "MaterialDiffuse", wndCount, &newCompMesh.Material.Diffuse, 0.0, 1.0)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Specular")
		guiAddSliderVec4(wnd, width4Col, "MaterialSpecular", wndCount, &newCompMesh.Material.Specular, 0.0, 1.0)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Shininess")
		wnd.DragSliderUFloat(fmt.Sprintf("MaterialShininess%d", wndCount), 0.1, &newCompMesh.Material.Shininess)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("DiffuseTex")
		loadDiffuseTexture, _ := wnd.Button(fmt.Sprintf("materialDiffuseTexLoad%d", wndCount), "L")
		wnd.Editbox(fmt.Sprintf("materialDiffuseTexEditbox%d", wndCount), &newCompMesh.Material.DiffuseTexture)
		if loadDiffuseTexture {
			doLoadTexture(newCompMesh.Material.DiffuseTexture)
		}

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("NormalsTex")
		loadNormalsTexture, _ := wnd.Button(fmt.Sprintf("materialNormalsTexLoad%d", wndCount), "L")
		wnd.Editbox(fmt.Sprintf("materialNormalsTexEditbox%d", wndCount), &newCompMesh.Material.NormalsTexture)
		if loadNormalsTexture {
			doLoadTexture(newCompMesh.Material.NormalsTexture)
		}

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("SpecularTex")
		loadSpecularTexture, _ := wnd.Button(fmt.Sprintf("materialSpecularTexLoad%d", wndCount), "L")
		wnd.Editbox(fmt.Sprintf("materialSpecularTexEditbox%d", wndCount), &newCompMesh.Material.SpecularTexture)
		if loadSpecularTexture {
			doLoadTexture(newCompMesh.Material.SpecularTexture)
		}
		// add in the custom textures
		var textureToDelete = -1
		for i := range newCompMesh.Material.Textures {
			wnd.StartRow()
			wnd.RequestItemWidthMin(textWidth)
			wnd.Text(fmt.Sprintf("Texture %d", i))
			deleteTexture, _ := wnd.Button(fmt.Sprintf("materialTexture%dDelete%d", i, wndCount), "X")
			loadTexture, _ := wnd.Button(fmt.Sprintf("materialTexture%dLoad%d", i, wndCount), "L")
			wnd.Editbox(fmt.Sprintf("materialTexture%dEditbox%d", i, wndCount), &newCompMesh.Material.Textures[i])

			if deleteTexture {
				textureToDelete = i
			}
			if loadTexture {
				doLoadTexture(newCompMesh.Material.Textures[i])
			}
		}

		// did we try to delete a texture
		if textureToDelete != -1 {
			newCompMesh.Material.Textures = doDeleteTexture(textureToDelete, newCompMesh.Material.Textures)
		}

		wnd.StartRow()
		wnd.Space(textWidth)
		nextTexIndex := len(newCompMesh.Material.Textures)
		addNewTexture, _ := wnd.Button(fmt.Sprintf("materialAddTex%d_%d", nextTexIndex, wndCount), "Add Texture")
		if addNewTexture {
			newCompMesh.Material.Textures = append(newCompMesh.Material.Textures, "")
		}

		wnd.StartRow()
		wnd.Space(textWidth)
		wnd.Checkbox(fmt.Sprintf("MaterialGenerateMips%d", wndCount), &newCompMesh.Material.GenerateMipmaps)
		wnd.Text("Generate Mipmaps")

		// do the user interface for animations
		if newCompMesh.SrcMesh != nil && compRenderable != nil && len(newCompMesh.SrcMesh.Animations) > 0 {
			for aniIndex, animation := range newCompMesh.SrcMesh.Animations {
				if aniIndex == 0 {
					wnd.Separator()
					wnd.RequestItemWidthMin(textWidth)
					wnd.Text("Animation: ")
				} else {
					wnd.StartRow()
					wnd.Space(textWidth)
				}
				wnd.Checkbox(fmt.Sprintf("RunAnimations %d %d", aniIndex, wndCount), &compRenderable.AnimationsEnabled[0])
				wnd.Text(animation.Name)
				if compRenderable.AnimationsEnabled[0] {
					doAnimation(&animation, compRenderable.Renderable, totalTime)
				}
			}

			wnd.Separator()
		}
	})
	meshWnd.Title = "Mesh Properties"
	meshWnd.AutoAdjustHeight = false
	meshWnd.IsScrollable = true
	meshWnd.ShowScrollBar = true
}

// createComponentWindow creates the main component window GUI.
func createComponentWindow(sX, sY, sW, sH float32) *gui.Window {
	// create a window for operating on the component file
	componentWindow := uiman.NewWindow("Component", sX, sY, sW, sH, func(wnd *gui.Window) {
		loadComponent, _ := wnd.Button("componentFileLoadButton", "Load")
		saveComponent, _ := wnd.Button("componentFileSaveButton", "Save")
		wnd.Editbox("componentFileEditbox", &flagComponentFile)
		if saveComponent {
			err := doSaveComponent(&theComponent, flagComponentFile)
			if err != nil {
				fmt.Printf("Failed to save the component.\n%v\n", err)
			} else {
				fmt.Printf("Saved the component file: %s\n", flagComponentFile)
			}
		}

		if loadComponent {
			// remove all existing mesh windows
			closeAllMeshWindows()
			// load the component file again and create mesh windows / renderables
			doLoadComponentFile(flagComponentFile)
		}

		wnd.Separator()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Name")
		wnd.Editbox("componentNameEditbox", &theComponent.Name)

		// do the user interface for mesh windows
		wnd.Separator()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Meshes:")
		addMesh, _ := wnd.Button("componentFileAddMeshButton", "Add Mesh")
		if addMesh {
			doAddMesh()
		}

		meshesThatSurvive := theComponent.Meshes[:0]
		for compMeshIndex, compMesh := range theComponent.Meshes {
			wnd.StartRow()
			wnd.RequestItemWidthMin(textWidth)
			wnd.Text(fmt.Sprintf("%s", compMesh.Name))
			showMeshWnd, _ := wnd.Button(fmt.Sprintf("buttonShowMesh%d", compMeshIndex), "Show")
			hideMeshWnd, _ := wnd.Button(fmt.Sprintf("buttonHideMesh%d", compMeshIndex), "Hide")
			deleteMesh, _ := wnd.Button(fmt.Sprintf("buttonDeleteMesh%d", compMeshIndex), "Delete")
			if showMeshWnd {
				doShowMeshWindow(compMesh)
			}
			if hideMeshWnd || deleteMesh {
				doHideMeshWindow(compMesh)
			}
			if !deleteMesh {
				meshesThatSurvive = append(meshesThatSurvive, compMesh)
			} else {
				doDeleteMesh(compMesh.Name)
			}

		}
		// FIXME: not Destroying renderables for meshes that don't survive
		theComponent.Meshes = meshesThatSurvive

		// do the user interface for colliders
		wnd.Separator()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Colliders: ")
		addNewCollider, _ := wnd.Button("buttonAddCollider", "Add Collider")
		if addNewCollider {
			doAddCollider(&theComponent)
		}

		collidersThatSurvive := theComponent.Collisions[:0]
		visibleCollidersThatSurvive := visibleColliders[:0]
		for colliderIndex, collider := range theComponent.Collisions {
			wnd.StartRow()
			wnd.RequestItemWidthMin(textWidth)
			wnd.Text(fmt.Sprintf("Collider %d:", colliderIndex))

			delCollider, _ := wnd.Button(fmt.Sprintf("buttonDeleteCollider%d", colliderIndex), "X")
			prevColliderType, _ := wnd.Button(fmt.Sprintf("buttonPrevColliderType%d", colliderIndex), "<")
			nextColliderType, _ := wnd.Button(fmt.Sprintf("buttonNextColliderType%d", colliderIndex), ">")

			if !delCollider {
				collidersThatSurvive = append(collidersThatSurvive, collider)

				if prevColliderType {
					doPrevColliderType(collider)
				}
				if nextColliderType {
					doNextColliderType(collider)
				}

				switch collider.Type {
				case component.ColliderTypeAABB:
					wnd.Text("Axis Aligned Bounding Box")
					wnd.StartRow()
					wnd.Space(textWidth)
					wnd.RequestItemWidthMin(width4Col)
					wnd.Text("Min")
					guiAddDragSliderVec3(wnd, width4Col, "ColliderMin", colliderIndex, 0.01, &collider.Min)

					wnd.StartRow()
					wnd.Space(textWidth)
					wnd.RequestItemWidthMin(width4Col)
					wnd.Text("Max")
					guiAddDragSliderVec3(wnd, width4Col, "ColliderMax", colliderIndex, 0.01, &collider.Max)

				case component.ColliderTypeSphere:
					wnd.Text("Sphere")
					wnd.StartRow()
					wnd.Space(textWidth)
					wnd.RequestItemWidthMin(width4Col)
					wnd.Text("Offset")
					guiAddDragSliderVec3(wnd, width4Col, "ColliderOffset", colliderIndex, 0.01, &collider.Offset)

					wnd.StartRow()
					wnd.Space(textWidth)
					wnd.RequestItemWidthMin(width4Col)
					wnd.Text("Radius")
					wnd.DragSliderFloat(fmt.Sprintf("ColliderRadius%d", colliderIndex), 0.01, &collider.Radius)
				default:
					wnd.Text(fmt.Sprintf("Unknown collider (%d)!", collider.Type))
				}

				// see if we need to update the renderable if it exists already
				visibleColliders = doUpdateVisibleCollider(visibleColliders, collider, colliderIndex)
				visibleCollidersThatSurvive = append(visibleCollidersThatSurvive, visibleColliders[colliderIndex])
			}
		}
		theComponent.Collisions = collidersThatSurvive
		visibleColliders = visibleCollidersThatSurvive

		wnd.Separator()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Child Components:")
		addChildComponent, _ := wnd.Button("addChildComponent", "Add Child")
		if addChildComponent {
			doAddChildReference(&theComponent)
		}

		childRefsThatSurvive := theComponent.ChildReferences[:0]
		for childRefIndex, childRef := range theComponent.ChildReferences {
			wnd.StartRow()
			wnd.RequestItemWidthMin(textWidth)
			wnd.Text("File:")
			removeReference, _ := wnd.Button(fmt.Sprintf("childRefRemove%d", childRefIndex), "X")
			loadChildReference, _ := wnd.Button(fmt.Sprintf("childRefLoad%d", childRefIndex), "L")
			wnd.Editbox(fmt.Sprintf("childRefFileEditbox%d", childRefIndex), &childRef.File)

			wnd.StartRow()
			wnd.Space(textWidth)
			wnd.RequestItemWidthMin(width4Col)
			wnd.Text("Offset")
			guiAddDragSliderVec3(wnd, width4Col, "childRefLocation", childRefIndex, 0.01, &childRef.Location)

			wnd.StartRow()
			wnd.Space(textWidth)
			wnd.RequestItemWidthMin(width4Col)
			wnd.Text("Scale")
			guiAddDragSliderVec3(wnd, width4Col, "childRefScale", childRefIndex, 0.01, &childRef.Scale)

			wnd.StartRow()
			wnd.Space(textWidth)
			wnd.RequestItemWidthMin(width4Col)
			wnd.Text("Rot Axis")
			guiAddDragSliderVec3(wnd, width4Col, "childRefRotAxis", childRefIndex, 0.01, &childRef.RotationAxis)

			wnd.StartRow()
			wnd.Space(textWidth)
			wnd.RequestItemWidthMin(width4Col)
			wnd.Text("Rot Deg")
			wnd.DragSliderFloat(fmt.Sprintf("childRefRotDeg%d", childRefIndex), 0.1, &childRef.RotationDegrees)

			if !removeReference {
				childRefsThatSurvive = append(childRefsThatSurvive, childRef)
			}
			if loadChildReference {
				var err error
				childComponents, err = doLoadChildComponent(childComponents, childRef)
				if err != nil {
					fmt.Printf("Failed to load child component.\n%v\n", err)
				}
			}
		}
		theComponent.ChildReferences = childRefsThatSurvive

		// remove any visible child components that no longer have a reference
		childComponents = removeStaleChildComponents(childComponents, &theComponent, childRefFilenames)
	})
	return componentWindow
}

// updateVisibleMesh copies the settings from the ComponentMesh part of meshRenderable
// to the Renderable so that it renders correctly.
// This also gets attempts to get textures from textureMan as well.
func updateVisibleMesh(compRenderable *meshRenderable) {
	// push all settings from the component to the renderable
	compRenderable.Renderable.Location = compRenderable.ComponentMesh.Offset
	compRenderable.Renderable.Scale = compRenderable.ComponentMesh.Scale
	compRenderable.Renderable.Material.DiffuseColor = compRenderable.ComponentMesh.Material.Diffuse
	if compRenderable.ComponentMesh.RotationDegrees != 0.0 {
		compRenderable.Renderable.LocalRotation = mgl.QuatRotate(
			mgl.DegToRad(compRenderable.ComponentMesh.RotationDegrees),
			compRenderable.ComponentMesh.RotationAxis)
	}

	compRenderable.Renderable.Material.SpecularColor = compRenderable.ComponentMesh.Material.Specular
	compRenderable.Renderable.Material.Shininess = compRenderable.ComponentMesh.Material.Shininess

	// assign textures
	textures := compRenderable.ComponentMesh.Material.Textures
	for i := 0; i < len(textures); i++ {
		glTex, texFound := textureMan.GetTexture(textures[i])
		if texFound && i < fizzle.MaxCustomTextures {
			compRenderable.Renderable.Material.CustomTex[i] = glTex
		}
	}
	if len(compRenderable.ComponentMesh.Material.DiffuseTexture) > 0 {
		glTex, texFound := textureMan.GetTexture(compRenderable.ComponentMesh.Material.DiffuseTexture)
		if texFound {
			compRenderable.Renderable.Material.DiffuseTex = glTex
		}
	}
	if len(compRenderable.ComponentMesh.Material.NormalsTexture) > 0 {
		glTex, texFound := textureMan.GetTexture(compRenderable.ComponentMesh.Material.NormalsTexture)
		if texFound {
			compRenderable.Renderable.Material.NormalsTex = glTex
		}
	}
	if len(compRenderable.ComponentMesh.Material.SpecularTexture) > 0 {
		glTex, texFound := textureMan.GetTexture(compRenderable.ComponentMesh.Material.SpecularTexture)
		if texFound {
			compRenderable.Renderable.Material.SpecularTex = glTex
		}
	}

}

// updateChildComponentRenderable copies the location, scale and rotation from the
// child component reference to the renderable object.
func updateChildComponentRenderable(childRenderable *fizzle.Renderable, childComp *component.ChildRef) {
	// push all settings from the child component to the renderable
	childRenderable.Location = childComp.Location
	childRenderable.Scale = childComp.Scale
	if childComp.RotationDegrees != 0.0 {
		childRenderable.LocalRotation = mgl.QuatRotate(mgl.DegToRad(childComp.RotationDegrees), childComp.RotationAxis)
	}
}

func main() {
	// parse the command line options
	flag.Parse()

	// Setup the log handlers
	groggy.Register("INFO", groggy.DefaultSyncHandler)
	groggy.Register("ERROR", groggy.DefaultSyncHandler)
	groggy.Register("DEBUG", groggy.DefaultSyncHandler)

	// start off by initializing the GL and GLFW libraries and creating a window.
	w, gfx := initGraphics("Component Editor", windowWidth, windowHeight)
	mainWindow = w

	/////////////////////////////////////////////////////////////////////////////
	// create and initialize the gui Manager
	uiman = gui.NewManager(gfx)
	err := uiman.Initialize(gui.VertShader330, gui.FragShader330, int32(windowWidth), int32(windowHeight), int32(windowHeight))
	if err != nil {
		panic("Failed to initialize the user interface! " + err.Error())
	}
	guiinput.SetInputHandlers(uiman, mainWindow)

	// load a font
	_, err = uiman.NewFont("Default", fontFilepath, fontScale, fontGlyphs)
	if err != nil {
		panic("Failed to load the font file! " + err.Error())
	}

	/////////////////////////////////////////////////////////////////////////////
	// setup renderer and shaders
	renderer = forward.NewForwardRenderer(gfx)
	renderer.ChangeResolution(int32(windowWidth), int32(windowHeight))
	defer renderer.Destroy()
	textureMan = fizzle.NewTextureManager()

	// load the basic shader
	basicShader, err := fizzle.LoadShaderProgramFromFiles(basicShaderFilepath, nil)
	if err != nil {
		panic("Failed to compile and link the basic shader program! " + err.Error())
	}

	// load the basic skinned shader
	basicSkinnedShader, err := fizzle.LoadShaderProgramFromFiles(basicSkinnedShaderFilepath, nil)
	if err != nil {
		panic("Failed to compile and link the basic skinned shader program! " + err.Error())
	}

	// load the color shader
	colorShader, err := fizzle.LoadShaderProgramFromFiles(colorShaderFilepath, nil)
	if err != nil {
		panic("Failed to compile and link the color shader program! " + err.Error())
	}

	shaders = make(map[string]*fizzle.RenderShader)
	shaders["Basic"] = basicShader
	shaders["BasicSkinned"] = basicSkinnedShader
	shaders["Color"] = colorShader

	// setup the component manager
	componentMan = component.NewManager(textureMan, shaders)

	// setup the camera to look at the component
	camera = fizzle.NewOrbitCamera(mgl.Vec3{0, 0, 0}, math.Pi/2.0, 5.0, math.Pi/2.0)

	// put a light in there
	light := renderer.NewDirectionalLight(mgl.Vec3{1.0, -0.5, -1.0})
	light.AmbientIntensity = 0.5
	light.DiffuseIntensity = 0.5
	light.SpecularIntensity = 0.3
	renderer.ActiveLights[0] = light

	/////////////////////////////////////////////////////////////////////////////
	// setup the component and user interface
	visibleMeshes = make(map[string]*meshRenderable)
	visibleColliders = make([]*colliderRenderable, 0)
	childRefFilenames = make(map[string]string)

	// if the component file passed in as a flag exists, try to load it
	doLoadComponentFile(flagComponentFile)

	// create the main component window
	componentWindow := createComponentWindow(0.01, 0.99, 0.25, 0.5)
	componentWindow.Title = "Component File"
	componentWindow.ShowTitleBar = false
	componentWindow.ShowScrollBar = true
	componentWindow.IsScrollable = true
	componentWindow.IsMoveable = true

	/////////////////////////////////////////////////////////////////////////////
	// loop until something told the mainWindow that it should close
	// set some OpenGL flags
	gfx.Enable(graphics.CULL_FACE)
	gfx.Enable(graphics.DEPTH_TEST)
	gfx.Enable(graphics.PROGRAM_POINT_SIZE)
	gfx.Enable(graphics.BLEND)
	gfx.Enable(graphics.LINES)
	gfx.BlendFunc(graphics.SRC_ALPHA, graphics.ONE_MINUS_SRC_ALPHA)

	lastFrame := time.Now()
	appStartTime := time.Now()
	for !mainWindow.ShouldClose() {
		// calculate the difference in time to control rotation speed
		thisFrame := time.Now()
		totalTime = thisFrame.Sub(appStartTime).Seconds()
		frameDelta := thisFrame.Sub(lastFrame).Seconds()

		// check for input
		handleInput(mainWindow, float32(frameDelta))

		// clear the screen
		width, height := renderer.GetResolution()
		gfx.Viewport(0, 0, int32(width), int32(height))
		gfx.ClearColor(clearColor[0], clearColor[1], clearColor[2], clearColor[3])
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		perspective := mgl.Perspective(mgl.DegToRad(60.0), float32(width)/float32(height), 1.0, 100.0)
		view := camera.GetViewMatrix()

		// draw the meshes that are visible
		for _, compRenderable := range visibleMeshes {
			// push all settings from the component to the renderable
			updateVisibleMesh(compRenderable)

			// draw the thing
			renderer.DrawRenderable(compRenderable.Renderable, nil, perspective, view, camera)
		}

		// draw the child components
		for _, childRef := range theComponent.ChildReferences {
			matchedChild := getLoadedChildComponent(childComponents, childRef.File)
			if matchedChild != nil {
				r := matchedChild.GetRenderable(textureMan, shaders)
				updateChildComponentRenderable(r, childRef)
				renderer.DrawRenderable(r, nil, perspective, view, camera)
			}
		}

		// draw all of the colliders
		gfx.Disable(graphics.DEPTH_TEST)
		for _, visCollider := range visibleColliders {
			renderer.DrawLines(visCollider.Renderable, colorShader, nil, perspective, view, camera)
		}
		gfx.Enable(graphics.DEPTH_TEST)

		// draw the user interface
		uiman.Construct(frameDelta)
		uiman.Draw()

		// draw the screen
		mainWindow.SwapBuffers()

		// advise GLFW to poll for input. without this the window appears to hang.
		glfw.PollEvents()

		// update our last frame time
		lastFrame = thisFrame
	}

	// cleanup
	for _, vc := range visibleColliders {
		vc.Renderable.Destroy()
	}
	for _, vm := range visibleMeshes {
		vm.Renderable.Destroy()
	}
	textureMan.Destroy()
	componentMan.Destroy()
	for _, shader := range shaders {
		shader.Destroy()
	}

	renderer.Destroy()
}

// initGraphics creates an OpenGL window and initializes the required graphics libraries.
// It will either succeed or panic.
func initGraphics(title string, w int, h int) (*glfw.Window, graphics.GraphicsProvider) {
	// GLFW must be initialized before it's called
	err := glfw.Init()
	if err != nil {
		panic("Can't init glfw! " + err.Error())
	}

	// request a OpenGL 3.3 core context
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// get a list of all the monitors to use and then take the one
	// specified by the command line flag
	monitors := glfw.GetMonitors()
	if flagDesktopNumber >= len(monitors) {
		flagDesktopNumber = -1
	}
	var monitorToUse *glfw.Monitor
	if flagDesktopNumber >= 0 {
		monitorToUse = monitors[flagDesktopNumber]
	}

	// do the actual window creation
	mainWindow, err = glfw.CreateWindow(w, h, title, monitorToUse, nil)
	if err != nil {
		panic("Failed to create the main window! " + err.Error())
	}
	mainWindow.SetSizeCallback(onWindowResize)
	mainWindow.MakeContextCurrent()

	// disable v-sync for max draw rate
	glfw.SwapInterval(0)

	// initialize OpenGL
	gfx, err := opengl.InitOpenGL()
	if err != nil {
		panic("Failed to initialize OpenGL! " + err.Error())
	}
	fizzle.SetGraphics(gfx)

	return mainWindow, gfx
}

// handleInput checks for keys and does some updates.
func handleInput(w *glfw.Window, delta float32) {
	const minDistance float32 = 3.0
	const zoomSpeed float32 = 3.0
	const rotSpeed = math.Pi

	rmbStatus := w.GetMouseButton(glfw.MouseButton2)
	if rmbStatus == glfw.Press {
		if w.GetKey(glfw.KeyA) == glfw.Press {
			camera.Rotate(delta * rotSpeed)
		}
		if w.GetKey(glfw.KeyD) == glfw.Press {
			camera.Rotate(delta * rotSpeed * -1.0)
		}

		if w.GetKey(glfw.KeyW) == glfw.Press {
			camera.RotateVertical(delta * rotSpeed)
		}
		if w.GetKey(glfw.KeyS) == glfw.Press {
			camera.RotateVertical(delta * rotSpeed * -1.0)
		}

		if w.GetKey(glfw.KeyQ) == glfw.Press {
			d := camera.GetDistance()
			newD := d + delta*zoomSpeed
			camera.SetDistance(newD)
		}
		if w.GetKey(glfw.KeyE) == glfw.Press {
			d := camera.GetDistance()
			newD := d - delta*zoomSpeed
			if newD > minDistance {
				camera.SetDistance(newD)
			}
		}
	}
}

// onWindowResize is called when the window changes size
func onWindowResize(w *glfw.Window, width int, height int) {
	uiman.AdviseResolution(int32(width), int32(height))
	renderer.ChangeResolution(int32(width), int32(height))
}
