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
	colorShader                *fizzle.RenderShader
	colorShaderFilepath        = "../../examples/assets/forwardshaders/color"
	basicShader                *fizzle.RenderShader
	basicShaderFilepath        = "../../examples/assets/forwardshaders/basic"
	basicSkinnedShader         *fizzle.RenderShader
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

// getMeshRenderableByName will return a *meshRenderable for a given
// ComponentMesh name from visibleMeshes
func getMeshRenderableByName(compMeshName string) *meshRenderable {
	for _, cr := range visibleMeshes {
		if cr.ComponentMesh.Name == compMeshName {
			return cr
		}
	}
	return nil
}

func makeRenderableForMesh(compMesh *component.Mesh) *fizzle.Renderable {
	prefixDir := ""
	if len(flagComponentFile) > 0 {
		prefixDir, _ = filepath.Split(flagComponentFile)
	}

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
	r.Core.Shader = basicSkinnedShader
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

func createMeshWindow(newCompMesh *component.Mesh, screenX, screenY float32) {
	// FIXME: find a better spot to spawn potentially
	meshWnd := uiman.NewWindow(compMeshWindowID, screenX, screenY, 0.30, 0.75, func(wnd *gui.Window) {
		compRenderable := getMeshRenderableByName(newCompMesh.Name)
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Name")
		wnd.Editbox("meshNameEditbox", &newCompMesh.Name)

		// force the window id to be the mesh name plus a prefix
		wnd.ID = fmt.Sprintf("%s%s", compMeshWindowID, newCompMesh.Name)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Source")
		loadSource, _ := wnd.Button("meshLoadSrcButton", "L")
		wnd.Editbox("meshSourceFileEditbox", &newCompMesh.SrcFile)
		if loadSource {
			makeRenderableForMesh(newCompMesh)
		}

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Gombz")
		saveGombz, _ := wnd.Button("meshSaveBinButton", "S")
		wnd.Editbox("meshBinaryFileEditbox", &newCompMesh.BinFile)
		if saveGombz && newCompMesh.SrcMesh != nil {
			gombzBytes, err := newCompMesh.SrcMesh.Encode()
			if err != nil {
				fmt.Printf("Error while serializing Gombz mesh: %v\n", err)
			} else {
				prefixDir := ""
				if len(flagComponentFile) > 0 {
					prefixDir, _ = filepath.Split(flagComponentFile)
				}

				gombzFilepath := prefixDir + newCompMesh.BinFile
				err = ioutil.WriteFile(gombzFilepath, gombzBytes, 0744)
				if err != nil {
					fmt.Printf("Error while writing Gombz file: %v\n", err)
				} else {
					fmt.Printf("Wrote Gombz file: %s\n", gombzFilepath)
				}
			}
		}

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Offset")
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("MeshOffsetX", 0.1, &newCompMesh.Offset[0])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("MeshOffsetY", 0.1, &newCompMesh.Offset[1])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("MeshOffsetZ", 0.1, &newCompMesh.Offset[2])

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Scale")
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("MeshScaleX", 0.1, &newCompMesh.Scale[0])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("MeshScaleY", 0.1, &newCompMesh.Scale[1])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("MeshScaleZ", 0.1, &newCompMesh.Scale[2])

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Rotation Axis")
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("MeshRotationAxisX", 0.01, &newCompMesh.RotationAxis[0])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("MeshRotationAxisY", 0.01, &newCompMesh.RotationAxis[1])
		wnd.RequestItemWidthMax(width3Col)
		wnd.DragSliderFloat("MeshRotationAxisZ", 0.01, &newCompMesh.RotationAxis[2])

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Rotation Degrees")
		wnd.DragSliderFloat("MeshRotationDegrees", 0.1, &newCompMesh.RotationDegrees)

		// ------------------------------------------------
		// material settings

		wnd.Separator()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Shader")
		wnd.Editbox("materialShaderNameEditbox", &newCompMesh.Material.ShaderName)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Diffuse")
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("MaterialDiffuseR", &newCompMesh.Material.Diffuse[0], 0.0, 1.0)
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("MaterialDiffuseG", &newCompMesh.Material.Diffuse[1], 0.0, 1.0)
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("MaterialDiffuseB", &newCompMesh.Material.Diffuse[2], 0.0, 1.0)
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("MaterialDiffuseA", &newCompMesh.Material.Diffuse[3], 0.0, 1.0)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Specular")
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("MaterialSpecularR", &newCompMesh.Material.Specular[0], 0.0, 1.0)
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("MaterialSpecularG", &newCompMesh.Material.Specular[1], 0.0, 1.0)
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("MaterialSpecularB", &newCompMesh.Material.Specular[2], 0.0, 1.0)
		wnd.RequestItemWidthMax(width4Col)
		wnd.SliderFloat("MaterialSpecularA", &newCompMesh.Material.Specular[3], 0.0, 1.0)

		wnd.StartRow()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Shininess")
		wnd.DragSliderUFloat("MaterialShininess", 0.1, &newCompMesh.Material.Shininess)

		var textureToDelete = -1
		for i := range newCompMesh.Material.Textures {
			wnd.StartRow()
			wnd.RequestItemWidthMin(textWidth)
			wnd.Text(fmt.Sprintf("Texture %d", i))
			deleteTexture, _ := wnd.Button(fmt.Sprintf("materialTexture%dDelete", i), "X")
			loadTexture, _ := wnd.Button(fmt.Sprintf("materialTexture%dLoad", i), "L")
			wnd.Editbox(fmt.Sprintf("materialTexture%dEditbox", i), &newCompMesh.Material.Textures[i])

			if deleteTexture {
				textureToDelete = i
			}
			if loadTexture {
				prefixDir := ""
				if len(flagComponentFile) > 0 {
					prefixDir, _ = filepath.Split(flagComponentFile)
				}

				texFile := newCompMesh.Material.Textures[i]
				texFilepath := prefixDir + texFile
				_, err := textureMan.LoadTexture(texFile, texFilepath)
				if err != nil {
					fmt.Printf("Failed to load texture %s: %v\n", texFile, err)
				} else {
					fmt.Printf("Loaded texture: %s\n", texFile)
				}
			}
		}

		// did we try to delete a texture
		if textureToDelete != -1 {
			newCompMesh.Material.Textures = append(
				newCompMesh.Material.Textures[:textureToDelete],
				newCompMesh.Material.Textures[textureToDelete+1:]...)
		}

		wnd.StartRow()
		wnd.Space(textWidth)
		nextTexIndex := len(newCompMesh.Material.Textures)
		addNewTexture, _ := wnd.Button(fmt.Sprintf("materialAddTex%d", nextTexIndex), "Add Texture")
		if addNewTexture {
			newCompMesh.Material.Textures = append(newCompMesh.Material.Textures, "")
		}

		wnd.StartRow()
		wnd.Space(textWidth)
		wnd.Checkbox("MaterialGenerateMips", &newCompMesh.Material.GenerateMipmaps)
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
				wnd.Checkbox(fmt.Sprintf("RunAnimations %d", aniIndex), &compRenderable.AnimationsEnabled[0])
				wnd.Text(animation.Name)
				if compRenderable.AnimationsEnabled[0] {
					aniTime := float32(math.Mod(totalTime*float64(animation.TicksPerSecond), float64(animation.Duration)))
					compRenderable.Renderable.Core.Skeleton.Animate(&animation, aniTime)
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
				createMeshWindow(compMesh, screenX, screenY)
				if compMesh.SrcFile != "" {
					makeRenderableForMesh(compMesh)
				}
				screenX += 0.05
				screenY -= 0.05

				// load any referenced textures
				for _, texFile := range compMesh.Material.Textures {
					prefixDir := ""
					if len(flagComponentFile) > 0 {
						prefixDir, _ = filepath.Split(flagComponentFile)
					}

					texFilepath := prefixDir + texFile
					_, err := textureMan.LoadTexture(texFile, texFilepath)
					if err != nil {
						fmt.Printf("Failed to load texture %s: %v\n", texFile, err)
					} else {
						fmt.Printf("Loaded texture: %s\n", texFile)
					}
				}
			}
		}
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
	basicShader, err = fizzle.LoadShaderProgramFromFiles(basicShaderFilepath, nil)
	if err != nil {
		panic("Failed to compile and link the basic shader program! " + err.Error())
	}
	defer basicShader.Destroy()

	// load the basic skinned shader
	basicSkinnedShader, err = fizzle.LoadShaderProgramFromFiles(basicSkinnedShaderFilepath, nil)
	if err != nil {
		panic("Failed to compile and link the basic skinned shader program! " + err.Error())
	}
	defer basicSkinnedShader.Destroy()

	// load the color shader
	colorShader, err = fizzle.LoadShaderProgramFromFiles(colorShaderFilepath, nil)
	if err != nil {
		panic("Failed to compile and link the color shader program! " + err.Error())
	}
	defer colorShader.Destroy()

	// setup the component manager
	shaders = make(map[string]*fizzle.RenderShader)
	shaders["Basic"] = basicShader
	shaders["BasicSkinned"] = basicSkinnedShader
	shaders["Color"] = colorShader
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

	// create a window for operating on the component file
	componentWindow := uiman.NewWindow("Component", 0.01, 0.99, 0.25, 0.5, func(wnd *gui.Window) {
		loadComponent, _ := wnd.Button("componentFileLoadButton", "Load")
		saveComponent, _ := wnd.Button("componentFileSaveButton", "Save")
		wnd.Editbox("componentFileEditbox", &flagComponentFile)
		if saveComponent {
			compJSON, jsonErr := json.MarshalIndent(theComponent, "", "    ")
			if jsonErr == nil {
				fileErr := ioutil.WriteFile(flagComponentFile, compJSON, 0744)
				if fileErr != nil {
					fmt.Printf("Failed to write component: %v\n", fileErr)
				} else {
					fmt.Printf("Saved the component file: %s\n", flagComponentFile)
				}
			} else {
				fmt.Printf("Failed to serialize component to JSON: %v\n", jsonErr)
			}
		} else if loadComponent {
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
			newCompMesh := component.NewMesh()
			newCompMesh.Name = fmt.Sprintf("Mesh %d", len(theComponent.Meshes)+1)
			theComponent.Meshes = append(theComponent.Meshes, newCompMesh)
			createMeshWindow(newCompMesh, meshWndX, meshWndY)
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
				meshWindow := uiman.GetWindow(fmt.Sprintf("%s%s", compMeshWindowID, compMesh.Name))
				if meshWindow == nil {
					createMeshWindow(compMesh, meshWndX, meshWndY)
				}
			}
			if hideMeshWnd || deleteMesh {
				meshWindow := uiman.GetWindow(fmt.Sprintf("%s%s", compMeshWindowID, compMesh.Name))
				if meshWindow != nil {
					uiman.RemoveWindow(meshWindow)
				}
			}
			if !deleteMesh {
				meshesThatSurvive = append(meshesThatSurvive, compMesh)
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
			// create a new sphere collider
			newCollider := new(component.CollisionRef)
			newCollider.Type = component.ColliderTypeSphere
			newCollider.Radius = 1.0
			theComponent.Collisions = append(theComponent.Collisions, newCollider)
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
					collider.Type = collider.Type - 1
					if collider.Type < 0 {
						collider.Type = component.ColliderTypeCount - 1
					}
				}
				if nextColliderType {
					collider.Type = collider.Type + 1
					if collider.Type >= component.ColliderTypeCount {
						collider.Type = 0
					}
				}

				switch collider.Type {
				case component.ColliderTypeAABB:
					wnd.Text("Axis Aligned Bounding Box")
					wnd.StartRow()
					wnd.Space(textWidth)
					wnd.RequestItemWidthMin(width4Col)
					wnd.Text("Min")
					wnd.RequestItemWidthMax(width4Col)
					wnd.DragSliderFloat(fmt.Sprintf("ColliderMinX%d", colliderIndex), 0.01, &collider.Min[0])
					wnd.RequestItemWidthMax(width4Col)
					wnd.DragSliderFloat(fmt.Sprintf("ColliderMinY%d", colliderIndex), 0.01, &collider.Min[1])
					wnd.RequestItemWidthMax(width4Col)
					wnd.DragSliderFloat(fmt.Sprintf("ColliderMinZ%d", colliderIndex), 0.01, &collider.Min[2])

					wnd.StartRow()
					wnd.Space(textWidth)
					wnd.RequestItemWidthMin(width4Col)
					wnd.Text("Max")
					wnd.RequestItemWidthMax(width4Col)
					wnd.DragSliderFloat(fmt.Sprintf("ColliderMaxX%d", colliderIndex), 0.01, &collider.Max[0])
					wnd.RequestItemWidthMax(width4Col)
					wnd.DragSliderFloat(fmt.Sprintf("ColliderMaxY%d", colliderIndex), 0.01, &collider.Max[1])
					wnd.RequestItemWidthMax(width4Col)
					wnd.DragSliderFloat(fmt.Sprintf("ColliderMaxZ%d", colliderIndex), 0.01, &collider.Max[2])

				case component.ColliderTypeSphere:
					wnd.Text("Sphere")
					wnd.StartRow()
					wnd.Space(textWidth)
					wnd.RequestItemWidthMin(width4Col)
					wnd.Text("Offset")
					wnd.RequestItemWidthMax(width4Col)
					wnd.DragSliderFloat(fmt.Sprintf("ColliderOffsetX%d", colliderIndex), 0.01, &collider.Offset[0])
					wnd.RequestItemWidthMax(width4Col)
					wnd.DragSliderFloat(fmt.Sprintf("ColliderOffsetY%d", colliderIndex), 0.01, &collider.Offset[1])
					wnd.RequestItemWidthMax(width4Col)
					wnd.DragSliderFloat(fmt.Sprintf("ColliderOffsetZ%d", colliderIndex), 0.01, &collider.Offset[2])

					wnd.StartRow()
					wnd.Space(textWidth)
					wnd.RequestItemWidthMin(width4Col)
					wnd.Text("Radius")
					wnd.DragSliderFloat(fmt.Sprintf("ColliderRadius%d", colliderIndex), 0.01, &collider.Radius)
				default:
					wnd.Text(fmt.Sprintf("Unknown collider (%d)!", collider.Type))
				}

				// see if we need to update the renderable if it exists already
				if len(visibleColliders) > colliderIndex {
					visCollider := visibleColliders[colliderIndex]

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
					// FIXME: not Destroying renderables for visCol's that don't survive
					visibleCollidersThatSurvive = append(visibleCollidersThatSurvive, visCollider)
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
					visibleCollidersThatSurvive = append(visibleCollidersThatSurvive, visCollider)
				}
			}
		}
		theComponent.Collisions = collidersThatSurvive
		visibleColliders = visibleCollidersThatSurvive

		wnd.Separator()
		wnd.RequestItemWidthMin(textWidth)
		wnd.Text("Child Components:")
		addChildComponent, _ := wnd.Button("addChildComponent", "Add Child")
		if addChildComponent {
			newChildRef := new(component.ChildRef)
			newChildRef.Scale = mgl.Vec3{1, 1, 1}
			theComponent.ChildReferences = append(theComponent.ChildReferences, newChildRef)
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
			wnd.RequestItemWidthMax(width4Col)
			wnd.DragSliderFloat(fmt.Sprintf("childRefLocationX%d", childRefIndex), 0.01, &childRef.Location[0])
			wnd.RequestItemWidthMax(width4Col)
			wnd.DragSliderFloat(fmt.Sprintf("childRefLocationY%d", childRefIndex), 0.01, &childRef.Location[1])
			wnd.RequestItemWidthMax(width4Col)
			wnd.DragSliderFloat(fmt.Sprintf("childRefLocationZ%d", childRefIndex), 0.01, &childRef.Location[2])

			wnd.StartRow()
			wnd.Space(textWidth)
			wnd.RequestItemWidthMin(width4Col)
			wnd.Text("Scale")
			wnd.RequestItemWidthMax(width4Col)
			wnd.DragSliderFloat(fmt.Sprintf("childRefScaleX%d", childRefIndex), 0.01, &childRef.Scale[0])
			wnd.RequestItemWidthMax(width4Col)
			wnd.DragSliderFloat(fmt.Sprintf("childRefScaleY%d", childRefIndex), 0.01, &childRef.Scale[1])
			wnd.RequestItemWidthMax(width4Col)
			wnd.DragSliderFloat(fmt.Sprintf("childRefScaleZ%d", childRefIndex), 0.01, &childRef.Scale[2])

			wnd.StartRow()
			wnd.Space(textWidth)
			wnd.RequestItemWidthMin(width4Col)
			wnd.Text("Rot Axis")
			wnd.RequestItemWidthMax(width4Col)
			wnd.DragSliderFloat(fmt.Sprintf("childRefRotAxisX%d", childRefIndex), 0.01, &childRef.RotationAxis[0])
			wnd.RequestItemWidthMax(width4Col)
			wnd.DragSliderFloat(fmt.Sprintf("childRefRotAxisY%d", childRefIndex), 0.01, &childRef.RotationAxis[1])
			wnd.RequestItemWidthMax(width4Col)
			wnd.DragSliderFloat(fmt.Sprintf("childRefRotAxisZ%d", childRefIndex), 0.01, &childRef.RotationAxis[2])

			wnd.StartRow()
			wnd.Space(textWidth)
			wnd.RequestItemWidthMin(width4Col)
			wnd.Text("Rot Deg")
			wnd.DragSliderFloat(fmt.Sprintf("childRefRotDeg%d", childRefIndex), 0.1, &childRef.RotationDegrees)

			if !removeReference {
				childRefsThatSurvive = append(childRefsThatSurvive, childRef)
			}
			if loadChildReference {
				prefixDir := ""
				if len(flagComponentFile) > 0 {
					prefixDir, _ = filepath.Split(flagComponentFile)
				}
				newChildComponent, err := componentMan.LoadComponentFromFile(prefixDir+childRef.File, childRef.File)
				if err != nil {
					fmt.Printf("Failed to load child component: %s\n%v\n", childRef.File, err)
				} else {
					fmt.Printf("Loaded child component: %s\n", childRef.File)
					childComponents = append(childComponents, newChildComponent)
					childRefFilenames[childRef.File] = newChildComponent.Name
				}
			}
		}
		theComponent.ChildReferences = childRefsThatSurvive

		// remove any visible child components that no longer have a reference
		childComponentsThatSurvive := childComponents[:0]
		for _, ref := range theComponent.ChildReferences {
			compNameToFind, okay := childRefFilenames[ref.File]
			if !okay {
				continue
			}

			for _, childCompToTest := range childComponents {
				if compNameToFind == childCompToTest.Name {
					childComponentsThatSurvive = append(childComponentsThatSurvive, childCompToTest)
					break
				}
			}
		}
		childComponents = childComponentsThatSurvive

	})
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
			compRenderable.Renderable.Location = compRenderable.ComponentMesh.Offset
			compRenderable.Renderable.Scale = compRenderable.ComponentMesh.Scale
			compRenderable.Renderable.Core.DiffuseColor = compRenderable.ComponentMesh.Material.Diffuse
			if compRenderable.ComponentMesh.RotationDegrees != 0.0 {
				compRenderable.Renderable.Rotation = mgl.QuatRotate(
					mgl.DegToRad(compRenderable.ComponentMesh.RotationDegrees),
					compRenderable.ComponentMesh.RotationAxis)
			}
			compRenderable.Renderable.Core.SpecularColor = compRenderable.ComponentMesh.Material.Specular
			compRenderable.Renderable.Core.Shininess = compRenderable.ComponentMesh.Material.Shininess

			// assign textures
			textures := compRenderable.ComponentMesh.Material.Textures
			for i := 0; i < len(textures); i++ {
				glTex, texFound := textureMan.GetTexture(textures[i])
				if texFound {
					compRenderable.Renderable.Core.Tex[i] = glTex
				}
			}

			// draw the thing
			renderer.DrawRenderable(compRenderable.Renderable, nil, perspective, view, camera)
		}

		// draw the child components
		for _, childComponent := range childComponents {
			r := childComponent.GetRenderable(textureMan, shaders)
			renderer.DrawRenderable(r, nil, perspective, view, camera)
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
