// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

// Gizmo is the transform gizmo that can be drawn in the editor.
type Gizmo struct {
	// Renderable is the drawable gizmo object for the current operation.
	Renderable *fizzle.Renderable

	translate *fizzle.Renderable
}

// CreateGizmo allocates a new gizmo and builds the renderable with the shader specified.
// (Shader should support Vert & VertColor)
func CreateGizmo(shader *fizzle.RenderShader) *Gizmo {
	g := new(Gizmo)
	g.buildRenderables(shader)
	return g
}

func addAxisToVBO(xmin, xmax, ymin, ymax, zmin, zmax, r, g, b, a float32, verts []float32) []float32 {
	/* Cube vertices are layed out like this:

	  +--------+           6          5
	/ |       /|
	+--------+ |        1          0        +Y
	| |      | |                            |___ +X
	| +------|-+           7          4    /
	|/       |/                           +Z
	+--------+          2          3

	*/
	return append(verts,
		xmax, ymax, zmax, r, g, b, a, xmin, ymax, zmax, r, g, b, a, xmin, ymin, zmax, r, g, b, a, xmax, ymin, zmax, r, g, b, a, // v0,v1,v2,v3 (front)
		xmax, ymax, zmin, r, g, b, a, xmax, ymax, zmax, r, g, b, a, xmax, ymin, zmax, r, g, b, a, xmax, ymin, zmin, r, g, b, a, // v5,v0,v3,v4 (right)
		xmax, ymax, zmin, r, g, b, a, xmin, ymax, zmin, r, g, b, a, xmin, ymax, zmax, r, g, b, a, xmax, ymax, zmax, r, g, b, a, // v5,v6,v1,v0 (top)
		xmin, ymax, zmax, r, g, b, a, xmin, ymax, zmin, r, g, b, a, xmin, ymin, zmin, r, g, b, a, xmin, ymin, zmax, r, g, b, a, // v1,v6,v7,v2 (left)
		xmax, ymin, zmax, r, g, b, a, xmin, ymin, zmax, r, g, b, a, xmin, ymin, zmin, r, g, b, a, xmax, ymin, zmin, r, g, b, a, // v3,v2,v7,v4 (bottom)
		xmin, ymax, zmin, r, g, b, a, xmax, ymax, zmin, r, g, b, a, xmax, ymin, zmin, r, g, b, a, xmin, ymin, zmin, r, g, b, a, // v6,v5,v4,v7 (back)
	)
}

func buildAxisSet() (verts []float32, indexes []uint32) {
	verts = make([]float32, 0, (4 * (3 + 4) * 6 * 3))                                   // 4 verts, 3+4 floats for pos&color, 6 faces, 3 total rectangles
	verts = addAxisToVBO(0.1, 0.9, -0.01, 0.01, -0.01, 0.01, 1.0, 0.0, 0.0, 0.5, verts) // x-axis / red
	verts = addAxisToVBO(-0.01, 0.01, 0.1, 0.9, -0.01, 0.01, 0.0, 1.0, 0.0, 0.5, verts) // y-axis / green
	verts = addAxisToVBO(-0.01, 0.01, -0.01, 0.01, 0.1, 0.9, 0.0, 0.0, 1.0, 0.5, verts) // z-axis / blue

	// make the indexes for the three axis rectangles
	indexPattern := [...]uint32{
		0, 1, 2, 2, 3, 0,
		4, 5, 6, 6, 7, 4,
		8, 9, 10, 10, 11, 8,
		12, 13, 14, 14, 15, 12,
		16, 17, 18, 18, 19, 16,
		20, 21, 22, 22, 23, 20,
	}
	indexes = make([]uint32, 0, 6*6*3)
	for i := 0; i < 3; i++ {
		for _, idx := range indexPattern {
			indexes = append(indexes, idx+uint32(i*24))
		}
	}

	return verts, indexes
}

func assembleIntoRenderable(verts []float32, indexes []uint32, facecount int) *fizzle.Renderable {
	const floatSize = 4
	const uintSize = 4

	robj := fizzle.NewRenderable()
	robj.FaceCount = 12 * 3
	robj.BoundingRect.Bottom = mgl.Vec3{-1, -1, -1}
	robj.BoundingRect.Top = mgl.Vec3{1, 1, 1}

	// create a VBO to hold the vertex data
	gfx := fizzle.GetGraphics()
	robj.Core.VertVBO = gfx.GenBuffer()
	robj.Core.VertColorVBO = robj.Core.VertVBO
	robj.Core.VertVBOOffset = 0
	robj.Core.VertColorVBOOffset = floatSize * 3
	robj.Core.VBOStride = floatSize * (3 + 4) // vert / vertcolor
	gfx.BindBuffer(graphics.ARRAY_BUFFER, robj.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(verts), gfx.Ptr(&verts[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	robj.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, robj.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gfx.Ptr(&indexes[0]), graphics.STATIC_DRAW)

	robj.Material = fizzle.NewMaterial()

	return robj
}

func (g *Gizmo) buildRenderables(shader *fizzle.RenderShader) {
	const axisFaceCount = 12 * 3

	// build the translate gizmo
	verts, indexes := buildAxisSet()
	g.translate = assembleIntoRenderable(verts, indexes, axisFaceCount)
	g.translate.Material.Shader = shader

	// set the current gizmo to translate
	g.Renderable = g.translate
}
