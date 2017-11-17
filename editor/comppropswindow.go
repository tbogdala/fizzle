// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/tbogdala/fizzle/component"
)

// ComponentPropsWindow is a nuklear window for editing component properties
type ComponentPropsWindow struct {
	Component *component.Component
	Bounds    nk.Rect
	Title     string

	// the component mesh property editor window
	meshPropsWindow *ComponentMeshPropsWindow
}

// NewComponentPropsWindow creates a new ComponentPropsWindow window with the initial settings provided.
func NewComponentPropsWindow(title string, comp *component.Component, customBounds *nk.Rect) *ComponentPropsWindow {
	w := new(ComponentPropsWindow)
	w.Title = title
	w.Component = comp
	if customBounds != nil {
		w.Bounds = *customBounds
	} else {
		w.Bounds = nk.NkRect(200, 200, 350, 500)
	}

	return w
}

// Render draws the window listing the properties for the selected component mesh.
func (w *ComponentPropsWindow) Render(ctx *nk.Context, window *glfw.Window) {
	winWidth, _ := window.GetSize()
	update := nk.NkBegin(ctx, "Component Properties", w.Bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowMinimizable|nk.WindowScalable)
	if update > 0 {
		active := w.Component

		// put in the component name
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 40)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "Name:", nk.TextLeft)
			newString, _ := editString(ctx, nk.EditField, active.Name, nk.NkFilterDefault)
			active.Name = newString
		}

		// put in the component offset
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkLabel(ctx, "Offset:", nk.TextLeft)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#x:", -100000.0, &active.Offset[0], 100000.0, 0.01, 0.1)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#y:", -100000.0, &active.Offset[1], 100000.0, 0.01, 0.1)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#z:", -100000.0, &active.Offset[2], 100000.0, 0.01, 0.1)

		// put in the collapsable mesh list
		_, fileName, fileLine, _ := runtime.Caller(0)
		hashStr := fmt.Sprintf("%s:%d", fileName, fileLine)
		if nk.NkTreePushHashed(ctx, nk.TreeTab, "Meshes", nk.Minimized, hashStr, int32(len(hashStr)), int32(fileLine)) != 0 {
			nk.NkLayoutRowDynamic(ctx, 120, 1)
			{
				nk.NkGroupBegin(ctx, "Mesh List", nk.WindowBorder)
				{
					// put a label in for each mesh the component has
					if len(active.Meshes) > 0 {
						for _, compMesh := range active.Meshes {
							nk.NkLayoutRowTemplateBegin(ctx, 30)
							nk.NkLayoutRowTemplatePushVariable(ctx, 80)
							nk.NkLayoutRowTemplatePushStatic(ctx, 40)
							nk.NkLayoutRowTemplateEnd(ctx)

							nk.NkLabel(ctx, compMesh.Name, nk.TextLeft)
							if nk.NkButtonLabel(ctx, "Edit") > 0 {
								log.Println("[DEBUG] comp:mesh:edit pressed!")
								bounds := nk.NkRect(float32(winWidth)-310.0, 75, 300.0, 600)
								w.meshPropsWindow = NewComponentMeshPropsWindow(
									fmt.Sprintf("%s Properties", compMesh.Name),
									compMesh, &bounds)
							}
						}
					}
				}
			}
			nk.NkGroupEnd(ctx)
			nk.NkTreePop(ctx)
		}

		// put in the collapsable collisions list
		_, fileName, fileLine, _ = runtime.Caller(0)
		hashStr = fmt.Sprintf("%s:%d", fileName, fileLine)
		if nk.NkTreePushHashed(ctx, nk.TreeTab, "Colliders", nk.Minimized, hashStr, int32(len(hashStr)), int32(fileLine)) != 0 {
			nk.NkLayoutRowDynamic(ctx, 120, 1)
			{
				nk.NkGroupBegin(ctx, "Collider List", nk.WindowBorder)
				{
					// put a label in for each collider the component has
					if len(active.Collisions) > 0 {
						for i := range active.Collisions {
							nk.NkLayoutRowTemplateBegin(ctx, 30)
							nk.NkLayoutRowTemplatePushVariable(ctx, 80)
							nk.NkLayoutRowTemplatePushStatic(ctx, 40)
							nk.NkLayoutRowTemplateEnd(ctx)

							nk.NkLabel(ctx, fmt.Sprintf("Collider %d", i), nk.TextLeft)
							if nk.NkButtonLabel(ctx, "E") > 0 {
								log.Println("[DEBUG] comp:collider:edit pressed!")
							}
						}
					}
				}
			}
			nk.NkGroupEnd(ctx)
			nk.NkTreePop(ctx)
		}

		// put in the collapsable child component reference list
		_, fileName, fileLine, _ = runtime.Caller(0)
		hashStr = fmt.Sprintf("%s:%d", fileName, fileLine)
		if nk.NkTreePushHashed(ctx, nk.TreeTab, "Child Components", nk.Minimized, hashStr, int32(len(hashStr)), int32(fileLine)) != 0 {
			nk.NkLayoutRowDynamic(ctx, 120, 1)
			{
				nk.NkGroupBegin(ctx, "Child Components List", nk.WindowBorder)
				{
					// put a label in for each child component the component has
					if len(active.ChildReferences) > 0 {
						for _, childRef := range active.ChildReferences {
							nk.NkLayoutRowTemplateBegin(ctx, 30)
							nk.NkLayoutRowTemplatePushVariable(ctx, 80)
							nk.NkLayoutRowTemplatePushStatic(ctx, 40)
							nk.NkLayoutRowTemplateEnd(ctx)

							nk.NkLabel(ctx, childRef.File, nk.TextLeft)
							if nk.NkButtonLabel(ctx, "E") > 0 {
								log.Println("[DEBUG] comp:childref:edit pressed!")
							}
						}
					}
				}
			}
			nk.NkGroupEnd(ctx)
			nk.NkTreePop(ctx)
		}

		// properties
	}
	nk.NkEnd(ctx)

	if w.meshPropsWindow != nil {
		w.meshPropsWindow.Render(ctx)
	}

}
