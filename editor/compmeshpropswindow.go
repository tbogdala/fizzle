// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/golang-ui/nuklear/nk"
	"github.com/tbogdala/fizzle/component"
)

// ComponentMeshPropsWindow is a nuklear window for editing component mesh properties
type ComponentMeshPropsWindow struct {
	Mesh   *component.Mesh
	Bounds nk.Rect
	Title  string

	materialWindow *MaterialPropsWindow
}

// NewComponentMeshPropsWindow creates a new ComponentMeshPropsWindow window with the initial settings provided.
func NewComponentMeshPropsWindow(title string, mesh *component.Mesh, customBounds *nk.Rect) *ComponentMeshPropsWindow {
	w := new(ComponentMeshPropsWindow)
	w.Title = title
	w.Mesh = mesh
	if customBounds != nil {
		w.Bounds = *customBounds
	} else {
		w.Bounds = nk.NkRect(200, 200, 350, 500)
	}

	return w
}

// Render draws the window listing the properties for the selected component mesh.
func (w *ComponentMeshPropsWindow) Render(ctx *nk.Context) {
	update := nk.NkBegin(ctx, "Mesh Properties", w.Bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowMinimizable|nk.WindowScalable)
	if update > 0 {
		active := w.Mesh

		// put in the mesh name
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 40)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "Name:", nk.TextLeft)
			newString, _ := editString(ctx, nk.EditField, active.Name, nk.NkFilterDefault)
			active.Name = newString
		}

		// put in the mesh source file
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 40)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "Source:", nk.TextLeft)
			newString, _ := editString(ctx, nk.EditField, active.SrcFile, nk.NkFilterDefault)
			active.SrcFile = newString
		}

		// put in the mesh binary file
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 40)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "Binary:", nk.TextLeft)
			newString, _ := editString(ctx, nk.EditField, active.BinFile, nk.NkFilterDefault)
			active.BinFile = newString
		}

		// put in the mesh offset
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkLabel(ctx, "Offset:", nk.TextLeft)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#x:", -100000.0, &active.Offset[0], 100000.0, 0.01, 0.1)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#y:", -100000.0, &active.Offset[1], 100000.0, 0.01, 0.1)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#z:", -100000.0, &active.Offset[2], 100000.0, 0.01, 0.1)

		// put in the mesh rotation
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkLabel(ctx, "Rotation:", nk.TextLeft)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#x:", -100000.0, &active.RotationAxis[0], 100000.0, 0.01, 0.1)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#y:", -100000.0, &active.RotationAxis[1], 100000.0, 0.01, 0.1)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#z:", -100000.0, &active.RotationAxis[2], 100000.0, 0.01, 0.1)

		// put in the mesh rotation order
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		sel := nk.NkPropertyi(ctx, "Order: "+rotationOrderStrings[active.RotationOrder], 0, int32(active.RotationOrder), 11, 1, 1)
		active.RotationOrder = mgl.RotationOrder(sel)

		// put in the mesh scale
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkLabel(ctx, "Scale:", nk.TextLeft)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#x:", -100000.0, &active.Scale[0], 100000.0, 0.01, 0.1)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#y:", -100000.0, &active.Scale[1], 100000.0, 0.01, 0.1)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#z:", -100000.0, &active.Scale[2], 100000.0, 0.01, 0.1)

		nk.NkLayoutRowDynamic(ctx, 20, 1)
		if nk.NkButtonLabel(ctx, "Edit Material") > 0 {
			w.materialWindow = NewMaterialPropsWindow(fmt.Sprintf("%s Material Settings", active.Name), &active.Material, nil)
		}
	}
	nk.NkEnd(ctx)

	if w.materialWindow != nil {
		button := w.materialWindow.Render(ctx)
		if button == MaterialPropsWindowClose {
			w.materialWindow = nil
		}
	}
}
