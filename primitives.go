// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

// CreatePlaneXY makes a 2d Renderable object on the XY plane for the given size,
// where (x0,y0) is the lower left and (x1, y1) is the upper right coordinate.
func CreatePlaneXY(shader string, x0, y0, x1, y1 float32) *Renderable {
	verts := [12]float32{
		x0, y0, 0.0,
		x1, y0, 0.0,
		x0, y1, 0.0,
		x1, y1, 0.0,
	}
	indexes := [6]uint32{
		0, 1, 2,
		1, 3, 2,
	}
	uvs := [8]float32{
		0.0, 0.0,
		1.0, 0.0,
		0.0, 1.0,
		1.0, 1.0,
	}
	normals := [12]float32{
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
		0.0, 0.0, 1.0,
	}

	return createPlane(shader, x0, y0, x1, y1, verts, indexes, uvs, normals)
}

// CreatePlaneXZ makes a 2d Renderable object on the XZ plane for the given size,
// where (x0,z0) is the lower left and (x1, z1) is the upper right coordinate.
func CreatePlaneXZ(shader string, x0, z0, x1, z1 float32) *Renderable {
	verts := [12]float32{
		x0, 0.0, z0,
		x1, 0.0, z0,
		x0, 0.0, z1,
		x1, 0.0, z1,
	}
	indexes := [6]uint32{
		0, 1, 2,
		1, 3, 2,
	}
	uvs := [8]float32{
		0.0, 0.0,
		1.0, 0.0,
		0.0, 1.0,
		1.0, 1.0,
	}
	normals := [12]float32{
		0.0, 1.0, 0.0,
		0.0, 1.0, 0.0,
		0.0, 1.0, 0.0,
		0.0, 1.0, 0.0,
	}

	return createPlane(shader, x0, z0, x1, z1, verts, indexes, uvs, normals)
}

// createTangents constructs the tangents for the faces.
// NOTE: this is a general implementation that assumes there's no shared
// vertices between faces.
func createTangents(verts []float32, indexes []uint32, uvs []float32) []float32 {
	tangents := make([]float32, len(verts))

	for i := 0; i < len(indexes); i += 3 {
		index0 := indexes[i+0]
		index1 := indexes[i+1]
		index2 := indexes[i+2]

		v0 := verts[(index0 * 3) : (index0*3)+3]
		v1 := verts[(index1 * 3) : (index1*3)+3]
		v2 := verts[(index2 * 3) : (index2*3)+3]

		uv0 := uvs[(index0 * 2) : (index0*2)+2]
		uv1 := uvs[(index1 * 2) : (index1*2)+2]
		uv2 := uvs[(index2 * 2) : (index2*2)+2]

		deltaPos1 := mgl.Vec3{v1[0] - v0[0], v1[1] - v0[1], v1[2] - v0[2]}
		deltaPos2 := mgl.Vec3{v2[0] - v0[0], v2[1] - v0[1], v2[2] - v0[2]}
		deltaUv1 := mgl.Vec2{uv1[0] - uv0[0], uv1[1] - uv0[1]}
		deltaUv2 := mgl.Vec2{uv2[0] - uv0[0], uv2[1] - uv0[1]}

		r := float32(1.0) / (deltaUv1[0]*deltaUv2[1] - deltaUv1[1]*deltaUv2[0])
		d1 := deltaPos1.Mul(deltaUv2[1])
		d2 := deltaPos2.Mul(deltaUv1[1])
		tangent := d1.Sub(d2)
		tangent = tangent.Mul(r).Normalize()

		// set the tangent array data for each vertex's tangent
		for f := 0; f < 3; f++ {
			index := indexes[i+f]

			tangents[index*3+0] = tangent[0]
			tangents[index*3+1] = tangent[1]
			tangents[index*3+2] = tangent[2]
		}
	}

	return tangents
}

func createPlane(shader string, x0, y0, x1, y1 float32, verts [12]float32, indexes [6]uint32, uvs [8]float32, normals [12]float32) *Renderable {
	const floatSize = 4
	const uintSize = 4

	// calculate the tangents based on the vertices and UVs.
	tangents := createTangents(verts[:], indexes[:], uvs[:])

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.ShaderName = shader
	r.FaceCount = 2
	r.BoundingRect.Bottom = mgl.Vec3{x0, y0, 0.0}
	r.BoundingRect.Top = mgl.Vec3{x1, y1, 0.0}

	// create the buffer to hold all of the interleaved data
	const numOfVerts = 4
	vnutBuffer := make([]float32, 0, len(verts)+len(uvs)+len(normals)+len(tangents))
	for i := 0; i < numOfVerts; i++ {
		// add the vertex
		vnutBuffer = append(vnutBuffer, verts[i*3])
		vnutBuffer = append(vnutBuffer, verts[i*3+1])
		vnutBuffer = append(vnutBuffer, verts[i*3+2])

		// add the normal
		vnutBuffer = append(vnutBuffer, normals[i*3])
		vnutBuffer = append(vnutBuffer, normals[i*3+1])
		vnutBuffer = append(vnutBuffer, normals[i*3+2])

		// add the uv
		vnutBuffer = append(vnutBuffer, uvs[i*2])
		vnutBuffer = append(vnutBuffer, uvs[i*2+1])

		// add the tangents
		vnutBuffer = append(vnutBuffer, tangents[i*3])
		vnutBuffer = append(vnutBuffer, tangents[i*3+1])
		vnutBuffer = append(vnutBuffer, tangents[i*3+2])

	}

	// create a VBO to hold the vertex data
	r.Core.VertVBO = gfx.GenBuffer()
	r.Core.UvVBO = r.Core.VertVBO
	r.Core.NormsVBO = r.Core.VertVBO
	r.Core.TangentsVBO = r.Core.VertVBO

	r.Core.VertVBOOffset = 0
	r.Core.NormsVBOOffset = floatSize * 3
	r.Core.UvVBOOffset = floatSize * 6
	r.Core.TangentsVBOOffset = floatSize * 8
	r.Core.VBOStride = floatSize * (3 + 3 + 2 + 3) // vert / normal / uv / tangent
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(vnutBuffer), gfx.Ptr(&vnutBuffer[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gfx.Ptr(&indexes[0]), graphics.STATIC_DRAW)

	return r
}

// CreateCube creates a cube based on the dimensions specified.
func CreateCube(shader string, xmin, ymin, zmin, xmax, ymax, zmax float32) *Renderable {
	/* Cube vertices are layed out like this:

	  +--------+           6          5
	/ |       /|
	+--------+ |        1          0        +Y
	| |      | |                            |___ +X
	| +------|-+           7          4    /
	|/       |/                           +Z
	+--------+          2          3

	*/

	verts := [...]float32{
		xmax, ymax, zmax, xmin, ymax, zmax, xmin, ymin, zmax, xmax, ymin, zmax, // v0,v1,v2,v3 (front)
		xmax, ymax, zmin, xmax, ymax, zmax, xmax, ymin, zmax, xmax, ymin, zmin, // v5,v0,v3,v4 (right)
		xmax, ymax, zmin, xmin, ymax, zmin, xmin, ymax, zmax, xmax, ymax, zmax, // v5,v6,v1,v0 (top)
		xmin, ymax, zmax, xmin, ymax, zmin, xmin, ymin, zmin, xmin, ymin, zmax, // v1,v6,v7,v2 (left)
		xmax, ymin, zmax, xmin, ymin, zmax, xmin, ymin, zmin, xmax, ymin, zmin, // v3,v2,v7,v4 (bottom)
		xmin, ymax, zmin, xmax, ymax, zmin, xmax, ymin, zmin, xmin, ymin, zmin, // v6,v5,v4,v7 (back)
	}
	indexes := [...]uint32{
		0, 1, 2, 2, 3, 0,
		4, 5, 6, 6, 7, 4,
		8, 9, 10, 10, 11, 8,
		12, 13, 14, 14, 15, 12,
		16, 17, 18, 18, 19, 16,
		20, 21, 22, 22, 23, 20,
	}
	uvs := [...]float32{
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
	}
	normals := [...]float32{
		0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, // v0,v1,v2,v3 (front)
		1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, // v5,v0,v3,v4 (right)
		0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, // v5,v6,v1,v0 (top)
		-1, 0, 0, -1, 0, 0, -1, 0, 0, -1, 0, 0, // v1,v6,v7,v2 (left)
		0, -1, 0, 0, -1, 0, 0, -1, 0, 0, -1, 0, // v3,v2,v7,v4 (bottom)
		0, 0, -1, 0, 0, -1, 0, 0, -1, 0, 0, -1, // v6,v5,v4,v7 (back)
	}

	// calculate the tangents based on the vertices and UVs.
	tangents := createTangents(verts[:], indexes[:], uvs[:])

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.ShaderName = shader
	r.FaceCount = 12
	r.BoundingRect.Bottom = mgl.Vec3{xmin, ymin, zmin}
	r.BoundingRect.Top = mgl.Vec3{xmax, ymax, zmax}

	const floatSize = 4
	const uintSize = 4

	// create the buffer to hold all of the interleaved data
	const numOfVerts = 24
	vnutBuffer := make([]float32, 0, len(verts)+len(uvs)+len(normals)+len(tangents))
	for i := 0; i < numOfVerts; i++ {
		// add the vertex
		vnutBuffer = append(vnutBuffer, verts[i*3])
		vnutBuffer = append(vnutBuffer, verts[i*3+1])
		vnutBuffer = append(vnutBuffer, verts[i*3+2])

		// add the normal
		vnutBuffer = append(vnutBuffer, normals[i*3])
		vnutBuffer = append(vnutBuffer, normals[i*3+1])
		vnutBuffer = append(vnutBuffer, normals[i*3+2])

		// add the uv
		vnutBuffer = append(vnutBuffer, uvs[i*2])
		vnutBuffer = append(vnutBuffer, uvs[i*2+1])

		// add the tangents
		vnutBuffer = append(vnutBuffer, tangents[i*3])
		vnutBuffer = append(vnutBuffer, tangents[i*3+1])
		vnutBuffer = append(vnutBuffer, tangents[i*3+2])

	}

	// create a VBO to hold the vertex data
	r.Core.VertVBO = gfx.GenBuffer()
	r.Core.UvVBO = r.Core.VertVBO
	r.Core.NormsVBO = r.Core.VertVBO
	r.Core.TangentsVBO = r.Core.VertVBO

	r.Core.VertVBOOffset = 0
	r.Core.NormsVBOOffset = floatSize * 3
	r.Core.UvVBOOffset = floatSize * 6
	r.Core.TangentsVBOOffset = floatSize * 8
	r.Core.VBOStride = floatSize * (3 + 3 + 2 + 3) // vert / normal / uv / tangent
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(vnutBuffer), gfx.Ptr(&vnutBuffer[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gfx.Ptr(&indexes[0]), graphics.STATIC_DRAW)

	return r
}

// CreateWireframeCube makes a cube with vertex and element VBO objects designed to be
// rendered as graphics.LINES.
func CreateWireframeCube(shader string, xmin, ymin, zmin, xmax, ymax, zmax float32) *Renderable {
	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4
	const facesPerCollision = 16

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.ShaderName = shader
	r.FaceCount = facesPerCollision
	r.BoundingRect.Bottom = mgl.Vec3{xmin, ymin, zmin}
	r.BoundingRect.Top = mgl.Vec3{xmax, ymax, zmax}

	/* Cube vertices are layed out like this:

	  +--------+           6          5
	/ |       /|
	+--------+ |        1          0        +Y
	| |      | |                            |___ +X
	| +------|-+           7          4    /
	|/       |/                           +Z
	+--------+          2          3

	*/
	verts := [...]float32{
		xmax, ymax, zmax, xmin, ymax, zmax, xmin, ymin, zmax, xmax, ymin, zmax, // v0,v1,v2,v3 (front)
		xmax, ymax, zmin, xmax, ymax, zmax, xmax, ymin, zmax, xmax, ymin, zmin, // v5,v0,v3,v4 (right)
		xmin, ymax, zmax, xmin, ymax, zmin, xmin, ymin, zmin, xmin, ymin, zmax, // v1,v6,v7,v2 (left)
		xmin, ymax, zmin, xmax, ymax, zmin, xmax, ymin, zmin, xmin, ymin, zmin, // v6,v5,v4,v7 (back)
	}
	indexes := [...]uint32{
		0, 1, 1, 2, 2, 3, 3, 0,
		4, 5, 5, 6, 6, 7, 7, 4,
		8, 9, 9, 10, 10, 11, 11, 8,
		12, 13, 13, 14, 14, 15, 15, 12,
	}

	r.BoundingRect.Bottom = mgl.Vec3{xmin, ymin, zmin}
	r.BoundingRect.Top = mgl.Vec3{xmax, ymax, zmax}

	// create a VBO to hold the vertex data
	r.Core.VertVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(verts), gfx.Ptr(&verts[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gfx.Ptr(&indexes[0]), graphics.STATIC_DRAW)

	return r
}

// CreateSphere generates a 3d uv-sphere with the given radius and returns a Renderable.
func CreateSphere(shader string, radius float32, rings int, sectors int) *Renderable {
	// nothing to create
	if rings < 2 || sectors < 2 {
		return nil
	}

	const piDiv2 = math.Pi / 2.0

	indexes := make([]uint32, 0, rings*sectors)

	R := float64(1.0 / float32(rings-1))
	S := float64(1.0 / float32(sectors-1))

	// create the buffer to hold all of the interleaved data
	numOfVerts := rings * sectors
	vnutBuffer := make([]float32, 0, numOfVerts*(3+2+3))

	for ri := 0; ri < int(rings); ri++ {
		for si := 0; si < int(sectors); si++ {
			y := float32(math.Sin(-piDiv2 + math.Pi*float64(ri)*R))
			x := float32(math.Cos(2.0*math.Pi*float64(si)*S) * math.Sin(math.Pi*float64(ri)*R))
			z := float32(math.Sin(2.0*math.Pi*float64(si)*S) * math.Sin(math.Pi*float64(ri)*R))

			// add the vertex
			vnutBuffer = append(vnutBuffer, x*radius)
			vnutBuffer = append(vnutBuffer, y*radius)
			vnutBuffer = append(vnutBuffer, z*radius)

			// add the normal
			vnutBuffer = append(vnutBuffer, x)
			vnutBuffer = append(vnutBuffer, y)
			vnutBuffer = append(vnutBuffer, z)

			// add the uv
			vnutBuffer = append(vnutBuffer, float32(si)*float32(S))
			vnutBuffer = append(vnutBuffer, float32(ri)*float32(R))

			currentRow := ri * sectors
			nextRow := (ri + 1) * sectors

			indexes = append(indexes, uint32(currentRow+si))
			indexes = append(indexes, uint32(nextRow+si))
			indexes = append(indexes, uint32(nextRow+si+1))

			indexes = append(indexes, uint32(currentRow+si))
			indexes = append(indexes, uint32(nextRow+si+1))
			indexes = append(indexes, uint32(currentRow+si+1))
		}
	}

	// calculate the tangents based on the vertices and UVs.
	// FIXME: disabled for now, there's an error in this code somewhere
	// where indexes end up eclipsing the number of vertices
	//tangents := createTangents(verts[:], indexes[:], uvs[:])

	r := NewRenderable()
	r.ShaderName = shader
	r.FaceCount = uint32(rings * sectors * 2)
	r.BoundingRect.Bottom = mgl.Vec3{-radius, -radius, -radius}
	r.BoundingRect.Top = mgl.Vec3{radius, radius, radius}

	const floatSize = 4
	const uintSize = 4

	// create a VBO to hold the vertex data
	r.Core.VertVBO = gfx.GenBuffer()
	r.Core.UvVBO = r.Core.VertVBO
	r.Core.NormsVBO = r.Core.VertVBO

	r.Core.VertVBOOffset = 0
	r.Core.NormsVBOOffset = floatSize * 3
	r.Core.UvVBOOffset = floatSize * 6
	r.Core.VBOStride = floatSize * (3 + 3 + 2) // vert / normal / uv
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(vnutBuffer), gfx.Ptr(&vnutBuffer[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexes), gfx.Ptr(&indexes[0]), graphics.STATIC_DRAW)

	return r
}
