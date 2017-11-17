// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"github.com/golang-ui/nuklear/nk"
	"github.com/tbogdala/fizzle/component"
)

// enumeration defining the results of the Render() function indicating if buttons were pressed
const (
	MaterialPropsWindowNoPress = 0
	MaterialPropsWindowClose   = 1
)

// MaterialPropsWindow is a nuklear window for editing material properties of a mesh
type MaterialPropsWindow struct {
	Material *component.Material
	Bounds   nk.Rect
	Title    string

	// -------
	// NOTE: property shadow values to use for the nuklear api and
	// need to get synced to the Material
	genMipmaps int32
}

// NewMaterialPropsWindow creates a new MaterialPropsWindow window with the initial settings provided.
func NewMaterialPropsWindow(title string, material *component.Material, customBounds *nk.Rect) *MaterialPropsWindow {
	w := new(MaterialPropsWindow)
	w.Title = title
	w.Material = material
	if customBounds != nil {
		w.Bounds = *customBounds
	} else {
		w.Bounds = nk.NkRect(200, 200, 350, 510)
	}

	// populate the shadowed fields correctly
	if material.GenerateMipmaps {
		w.genMipmaps = nk.True
	} else {
		w.genMipmaps = nk.False
	}

	return w
}

// Render draws the file dialog box.
func (w *MaterialPropsWindow) Render(ctx *nk.Context) int {
	result := MaterialPropsWindowNoPress

	update := nk.NkBegin(ctx, w.Title, w.Bounds, nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowTitle)
	if update > 0 {
		m := w.Material
		// shader name
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 40)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "Shader:", nk.TextLeft)
			m.ShaderName, _ = editString(ctx, nk.EditField, m.ShaderName, nk.NkFilterDefault)
		}

		// diffuse
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkLabel(ctx, "Diffuse color:", nk.TextLeft)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#r:", 0.0, &m.Diffuse[0], 1.0, 0.01, 0.01)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#g:", 0.0, &m.Diffuse[1], 1.0, 0.01, 0.01)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#b:", 0.0, &m.Diffuse[2], 1.0, 0.01, 0.01)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#a:", 0.0, &m.Diffuse[2], 1.0, 0.01, 0.01)

		// specular
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkLabel(ctx, "Specular color:", nk.TextLeft)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#r:", 0.0, &m.Specular[0], 1.0, 0.01, 0.01)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#g:", 0.0, &m.Specular[1], 1.0, 0.01, 0.01)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#b:", 0.0, &m.Specular[2], 1.0, 0.01, 0.01)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#a:", 0.0, &m.Specular[2], 1.0, 0.01, 0.01)

		// shininess
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkPropertyFloat(ctx, "#shininess:", 0.0, &m.Shininess, 100.0, 0.1, 0.1)

		// mipmap
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		nk.NkCheckboxLabel(ctx, "Generate mipmaps", &w.genMipmaps)
		if w.genMipmaps == nk.True {
			m.GenerateMipmaps = true
		} else {
			m.GenerateMipmaps = false
		}

		// diffuse texture filepath
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 80)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "Diffuse tex:", nk.TextLeft)
			m.DiffuseTexture, _ = editString(ctx, nk.EditField, m.DiffuseTexture, nk.NkFilterDefault)
		}

		// normals texture filepath
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 80)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "Normals tex:", nk.TextLeft)
			m.NormalsTexture, _ = editString(ctx, nk.EditField, m.NormalsTexture, nk.NkFilterDefault)
		}

		// specular texture filepath
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 80)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "Specular tex:", nk.TextLeft)
			m.SpecularTexture, _ = editString(ctx, nk.EditField, m.SpecularTexture, nk.NkFilterDefault)
		}

		// close button
		nk.NkLayoutRowDynamic(ctx, 30, 1)
		if nk.NkButtonLabel(ctx, "Close") > 0 {
			result = MaterialPropsWindowClose
		}
	}
	nk.NkEnd(ctx)

	return result
}
