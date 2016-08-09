// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"fmt"
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

// CreatePlaneXY makes a 2d Renderable object on the XY plane for the given size,
// where (x0,y0) is the lower left and (x1, y1) is the upper right coordinate.
func CreatePlaneXY(x0, y0, x1, y1 float32) *Renderable {
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

	return createPlane(x0, y0, x1, y1, verts, indexes, uvs, normals)
}

// CreatePlaneXZ makes a 2d Renderable object on the XZ plane for the given size,
// where (x0,z0) is the lower left and (x1, z1) is the upper right coordinate.
func CreatePlaneXZ(x0, z0, x1, z1 float32) *Renderable {
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

	return createPlane(x0, z0, x1, z1, verts, indexes, uvs, normals)
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

func createPlane(x0, y0, x1, y1 float32, verts [12]float32, indexes [6]uint32, uvs [8]float32, normals [12]float32) *Renderable {
	const floatSize = 4
	const uintSize = 4

	// calculate the tangents based on the vertices and UVs.
	tangents := createTangents(verts[:], indexes[:], uvs[:])

	r := NewRenderable()
	r.Core = NewRenderableCore()
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
func CreateCube(xmin, ymin, zmin, xmax, ymax, zmax float32) *Renderable {
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
func CreateWireframeCube(xmin, ymin, zmin, xmax, ymax, zmax float32) *Renderable {
	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4
	const facesPerCollision = 16

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.FaceCount = facesPerCollision

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

// CreateLine makes a line between a two points
// rendered as graphics.LINES.
func CreateLine(x0, y0, z0, x1, y1, z1 float32) *Renderable {
	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.FaceCount = 1 //one line one face

	verts := [...]float32{
		x0, y0, z0, x1, y1, z1,
	}
	indexes := [...]uint32{
		0, 1,
	}

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

//axis for forming planes
const (
	X = 1 << iota
	Y
	Z
)

func genCircleSegData(xmin, ymin, zmin, radius float32, segments int, axis int) ([]float32, []uint32) {
	verts := []float32{}
	indexes := []uint32{}

	// create the lines for the circle
	radsPerSeg := math.Pi * 2.0 / float64(segments)
	for i := 0; i < segments; i++ {

		//XZ plane
		if axis == (X | Z) {
			verts = append(verts, xmin+(radius*float32(math.Cos(radsPerSeg*float64(i)))))
			verts = append(verts, ymin)
			verts = append(verts, zmin+(radius*float32(math.Sin(radsPerSeg*float64(i)))))
		}

		//XY plane
		if axis == (X | Y) {
			verts = append(verts, xmin+(radius*float32(math.Cos(radsPerSeg*float64(i)))))
			verts = append(verts, ymin+(radius*float32(math.Sin(radsPerSeg*float64(i)))))
			verts = append(verts, zmin)
		}

		//ZY plane
		if axis == (Z | Y) {
			verts = append(verts, xmin)
			verts = append(verts, ymin+(radius*float32(math.Cos(radsPerSeg*float64(i)))))
			verts = append(verts, zmin+(radius*float32(math.Sin(radsPerSeg*float64(i)))))
		}

		//XYZ ...
		if axis == (X | Y | Z) {
			verts = append(verts, xmin+(radius*float32(math.Cos(radsPerSeg*float64(i)))))
			verts = append(verts, ymin+(radius*float32(math.Cos(radsPerSeg*float64(i)))))
			verts = append(verts, zmin+(radius*float32(math.Sin(radsPerSeg*float64(i)))))
		}

		// original XZ verts
		// verts = append(verts, xmin+(radius*float32(math.Cos(radsPerSeg*float64(i)))))
		// verts = append(verts, ymin)
		// verts = append(verts, zmin+(radius*float32(math.Sin(radsPerSeg*float64(i)))))

		indexes = append(indexes, uint32(i))
		if i != segments-1 {
			indexes = append(indexes, uint32(i)+1)
		} else {
			indexes = append(indexes, uint32(0))
		}
	}

	return verts, indexes
}

// CreateWireframeCircle makes a cirle with vertex and element VBO objects designed to be
// rendered as graphics.LINES.
func CreateWireframeCircle(xmin, ymin, zmin, radius float32, segments int, axis int) *Renderable {
	// sanity check
	if segments == 0 {
		return nil
	}

	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4

	verts, indexes := genCircleSegData(xmin, ymin, zmin, radius, segments, axis)

	r := NewRenderable()
	r.Core = NewRenderableCore()
	r.FaceCount = uint32(segments)
	r.BoundingRect.Bottom = mgl.Vec3{xmin - radius, ymin, zmin - radius}
	r.BoundingRect.Top = mgl.Vec3{xmin + radius, ymin, zmin + radius}

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

// CreateWireframeConeSegmentXZ makes a cone segment with vertex and element VBO objects designed to be
// rendered as graphics.LINES wtih the default orientation of the cone segment along +Y.
func CreateWireframeConeSegmentXZ(xmin, ymin, zmin, bottomRadius, topRadius, length float32, circleSegments, sideSegments int) *Renderable {
	// sanity check
	if circleSegments == 0 {
		return nil
	}

	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4

	// create the bottom circle for the cone segment
	verts, indexes := genCircleSegData(xmin, ymin, zmin, bottomRadius, circleSegments, X|Z)

	// create the top circle for the cone segment
	topVerts, topIndexes := genCircleSegData(xmin, ymin+length, zmin, topRadius, circleSegments, X|Z)
	verts = append(verts, topVerts...)
	for _, index := range topIndexes {
		indexes = append(indexes, index+uint32(circleSegments))
	}

	// create the side lines that will connect the cone circles
	lineIndexOff := uint32(circleSegments) * 2
	radsPerSideSeg := math.Pi * 2.0 / float64(sideSegments)
	for i := 0; i < sideSegments; i++ {
		verts = append(verts, xmin+(bottomRadius*float32(math.Cos(radsPerSideSeg*float64(i)))))
		verts = append(verts, ymin)
		verts = append(verts, zmin+(bottomRadius*float32(math.Sin(radsPerSideSeg*float64(i)))))

		verts = append(verts, xmin+(topRadius*float32(math.Cos(radsPerSideSeg*float64(i)))))
		verts = append(verts, ymin+length)
		verts = append(verts, zmin+(topRadius*float32(math.Sin(radsPerSideSeg*float64(i)))))

		indexes = append(indexes, lineIndexOff+(uint32(i)*2))
		indexes = append(indexes, lineIndexOff+(uint32(i)*2)+1)
	}

	// figure out the biggest radius for the bounding box
	maxRadius := float32(math.Max(float64(bottomRadius), float64(topRadius)))

	r := NewRenderable()
	r.Core = NewRenderableCore()

	r.FaceCount = uint32(circleSegments)*2 + uint32(sideSegments)
	r.BoundingRect.Bottom = mgl.Vec3{xmin - maxRadius, ymin, zmin - maxRadius}
	r.BoundingRect.Top = mgl.Vec3{xmin + maxRadius, ymin + length, zmin + maxRadius}

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
func CreateSphere(radius float32, rings int, sectors int) *Renderable {
	// nothing to create
	if rings < 2 || sectors < 2 {
		return nil
	}

	const piDiv2 = math.Pi / 2.0

	R := float64(1.0 / float32(rings-1))
	S := float64(1.0 / float32(sectors-1))

	// create the buffer to hold all of the interleaved data
	numOfVerts := (rings + 1) * (sectors + 1)
	indexes := make([]uint32, 0, rings*sectors*6)
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

		}
	}

	for ri := 0; ri < int(rings-1); ri++ {
		for si := 0; si < int(sectors-1); si++ {
			currentRow := ri * sectors
			nextRow := (ri + 1) * sectors

			indexes = append(indexes, uint32(currentRow+si))
			indexes = append(indexes, uint32(nextRow+si))
			indexes = append(indexes, uint32(nextRow+si+1))

			indexes = append(indexes, uint32(nextRow+si+1))
			indexes = append(indexes, uint32(currentRow+si+1))
			indexes = append(indexes, uint32(currentRow+si))

			//indexes = append(indexes, uint32(currentRow+si))
			//indexes = append(indexes, uint32(nextRow+si+1))
			//indexes = append(indexes, uint32(currentRow+si+1))
		}
	}

	// calculate the tangents based on the vertices and UVs.
	// FIXME: disabled for now, there's an error in this code somewhere
	// where indexes end up eclipsing the number of vertices
	//tangents := createTangents(verts[:], indexes[:], uvs[:])

	r := NewRenderable()

	r.FaceCount = uint32(len(indexes) / 3)
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

// CreateCubeMappedSphere creates a sphere that can be used for cubemaps based on the dimensions specified.
// If the cubemapUvs parameter is true, it will map face UVs to a single cubemap texture; if
// this parameter is false, then each face is mapped [0..1] for UVs.
func CreateCubeMappedSphere(gridSize int, radius float32, cubemapUvs bool) *Renderable {
	// based on an implementation in C# for unity here:
	// http://catlikecoding.com/unity/tutorials/cube-sphere/
	//
	// the state of this implementation is pretty bad, but it seems to work.
	// 	¯\_(ツ)_/¯

	// sanity check the gridSize
	if gridSize < 2 {
		return nil
	}

	const floatSize = 4
	const uintSize = 4

	const xmin = float32(-1.0)
	const ymin = float32(-1.0)
	const zmin = float32(-1.0)
	const xmax = float32(1.0)
	const ymax = float32(1.0)
	const zmax = float32(1.0)

	const xWidth = xmax - xmin
	const yWidth = ymax - ymin
	const zWidth = zmax - zmin

	xStep := float32(xWidth) / float32(gridSize)
	yStep := float32(yWidth) / float32(gridSize)
	zStep := float32(zWidth) / float32(gridSize)

	// create the buffer to hold all of the interleaved data
	var vnutBuffer []float32
	var xUv, yUv, s, t float32

	scaleFn := func(gridSize int, x, y, z float32) mgl.Vec3 {
		v := mgl.Vec3{x, y, z}
		x2 := v[0] * v[0]
		y2 := v[1] * v[1]
		z2 := v[2] * v[2]

		return mgl.Vec3{
			v[0] * float32(math.Sqrt(float64(1.0-y2/2.0-z2/2.0+y2*z2/3.0))),
			v[1] * float32(math.Sqrt(float64(1.0-x2/2.0-z2/2.0+x2*z2/3.0))),
			v[2] * float32(math.Sqrt(float64(1.0-x2/2.0-y2/2.0+x2*y2/3.0))),
		}
	}

	addVertData := func(buff []float32, x, y, z, nx, ny, nz, uvs, uvt float32) []float32 {
		// vertex
		buff = append(buff, x)
		buff = append(buff, y)
		buff = append(buff, z)

		// normal
		buff = append(buff, nx)
		buff = append(buff, ny)
		buff = append(buff, nz)

		// uv
		buff = append(buff, uvs)
		buff = append(buff, uvt)

		return buff
	}

	createFaceIndexes := func(gridSize, faceCount int) ([]uint32, int) {
		indexesPerFace := (gridSize + 1) * (gridSize + 1)
		faceOffset := 0
		indexes := make([]uint32, 0, indexesPerFace*faceCount)

		// now add the indexes
		for face := 0; face < faceCount; face++ {
			for y := 0; y < gridSize; y++ {
				for x := 0; x < gridSize; x++ {
					i0 := faceOffset + x + (y * (gridSize + 1))
					i1 := faceOffset + x + 1 + (y * (gridSize + 1))
					i2 := faceOffset + x + ((y + 1) * (gridSize + 1))
					i3 := faceOffset + x + 1 + ((y + 1) * (gridSize + 1))

					indexes = append(indexes, uint32(i0))
					indexes = append(indexes, uint32(i1))
					indexes = append(indexes, uint32(i2))

					indexes = append(indexes, uint32(i1))
					indexes = append(indexes, uint32(i3))
					indexes = append(indexes, uint32(i2))
				} // x
			} // y

			faceOffset += indexesPerFace
		} // face

		return indexes, faceOffset
	}

	// =======================================================================
	// front face

	yUv = float32(0.0)
	for y := ymin; y <= ymax; y += yStep {
		xUv = float32(0.0)
		for x := xmin; x <= xmax; x += xStep {
			defPos := scaleFn(gridSize, x, y, zmax)
			if cubemapUvs {
				s, t = MapUvToCubemap(FaceFront, xUv, yUv)
			} else {
				s, t = xUv, yUv
			}
			vnutBuffer = addVertData(vnutBuffer, defPos[0]*radius, defPos[1]*radius, defPos[2]*radius, defPos[0], defPos[1], defPos[2], s, t)
			xUv += xStep / xWidth
		} // x
		yUv += yStep / yWidth
	} // y

	// =======================================================================
	// back face
	yUv = float32(0.0)
	for y := ymin; y <= ymax; y += yStep {
		xUv = float32(0.0)
		for x := xmax; x >= xmin; x -= xStep {
			defPos := scaleFn(gridSize, x, y, zmin)
			if cubemapUvs {
				s, t = MapUvToCubemap(FaceBack, xUv, yUv)
			} else {
				s, t = xUv, yUv
			}
			vnutBuffer = addVertData(vnutBuffer, defPos[0]*radius, defPos[1]*radius, defPos[2]*radius, defPos[0], defPos[1], defPos[2], s, t)
			xUv += xStep / xWidth
		} // x
		yUv += yStep / yWidth
	} // y

	// =======================================================================
	// right face
	yUv = float32(0.0)
	for y := ymin; y <= ymax; y += yStep {
		xUv = float32(0.0)
		for z := zmax; z >= zmin; z -= zStep {
			defPos := scaleFn(gridSize, xmax, y, z)
			if cubemapUvs {
				s, t = MapUvToCubemap(FaceRight, xUv, yUv)
			} else {
				s, t = xUv, yUv
			}
			vnutBuffer = addVertData(vnutBuffer, defPos[0]*radius, defPos[1]*radius, defPos[2]*radius, defPos[0], defPos[1], defPos[2], s, t)
			xUv += xStep / xWidth
		} // x
		yUv += yStep / yWidth
	} // y

	// =======================================================================
	// left face
	yUv = float32(0.0)
	for y := ymin; y <= ymax; y += yStep {
		xUv = float32(0.0)
		for z := zmin; z <= zmax; z += zStep {
			defPos := scaleFn(gridSize, xmin, y, z)
			if cubemapUvs {
				s, t = MapUvToCubemap(FaceLeft, xUv, yUv)
			} else {
				s, t = xUv, yUv
			}
			vnutBuffer = addVertData(vnutBuffer, defPos[0]*radius, defPos[1]*radius, defPos[2]*radius, defPos[0], defPos[1], defPos[2], s, t)
			xUv += xStep / xWidth
		} // x
		yUv += yStep / yWidth
	} // y

	// =======================================================================
	// bottom face
	yUv = float32(0.0)
	for z := zmin; z <= zmax; z += zStep {
		xUv = float32(0.0)
		for x := xmin; x <= xmax; x += xStep {
			defPos := scaleFn(gridSize, x, ymin, z)
			if cubemapUvs {
				s, t = MapUvToCubemap(FaceBottom, xUv, yUv)
			} else {
				s, t = xUv, yUv
			}
			vnutBuffer = addVertData(vnutBuffer, defPos[0]*radius, defPos[1]*radius, defPos[2]*radius, defPos[0], defPos[1], defPos[2], s, t)
			xUv += xStep / xWidth
		} // x
		yUv += yStep / yWidth
	} // y

	// =======================================================================
	// top face
	yUv = float32(0.0)
	for z := zmax; z >= zmin; z -= zStep {
		xUv = float32(0.0)
		for x := xmin; x <= xmax; x += xStep {
			defPos := scaleFn(gridSize, x, ymax, z)
			if cubemapUvs {
				s, t = MapUvToCubemap(FaceTop, xUv, yUv)
			} else {
				s, t = xUv, yUv
			}
			vnutBuffer = addVertData(vnutBuffer, defPos[0]*radius, defPos[1]*radius, defPos[2]*radius, defPos[0], defPos[1], defPos[2], s, t)
			xUv += xStep / xWidth
		} // x
		yUv += yStep / yWidth
	} // y

	// now add the indexes
	indexes, _ := createFaceIndexes(gridSize, 6)

	// =======================================================================
	r := NewRenderable()
	r.Core = NewRenderableCore()

	r.BoundingRect.Bottom = mgl.Vec3{xmin, ymin, zmin}
	r.BoundingRect.Top = mgl.Vec3{xmax, ymax, zmax}
	r.FaceCount = uint32(len(indexes) / 3)
	byteCount := floatSize*len(vnutBuffer) + uintSize*len(indexes)
	fmt.Printf("Face count = %d ; bytes = %dB (%.2fKB)\n", r.FaceCount, byteCount, float32(byteCount)/1024.0)

	// create a VBO to hold the vertex data
	r.Core.VertVBO = gfx.GenBuffer()
	r.Core.UvVBO = r.Core.VertVBO
	r.Core.NormsVBO = r.Core.VertVBO
	r.Core.TangentsVBO = r.Core.VertVBO
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

// constants used to define faces for use in functions that need to act differently
// based on the face.
const (
	FaceTop    = 0
	FaceFront  = 1
	FaceLeft   = 2
	FaceBottom = 3
	FaceRight  = 4
	FaceBack   = 5
)

// MapUvToCubemap takes a UV coordinate that is in range ([0..1],[0..1]) with
// respect to one side and returns a UV coordinate s and t value that is mapped
// to a single cubemap texture looking something like this:
//      .____.
//      |    |
//      | T  |
// .____.____.____.____.
// |    |    |    |    |
// |  L |  F | R  | Bk |
// .----.----.----.----.
//      |    |
//      | Bt |
//      .----.
//
// The resulting coordintes are for a texture wrapped around the outside
// of the cube.
func MapUvToCubemap(side int, s, t float32) (float32, float32) {
	uvs := [...]mgl.Vec2{
		{0.25, 0.875}, //   -1,  1, -1
		{0.5, 0.875},  //    1,  1, -1
		{0.0, 0.625},  //   -1,  1, -1
		{0.25, 0.625}, //	 -1,  1,  1

		{0.5, 0.625},  //    1,  1,  1
		{0.75, 0.625}, //    1,  1, -1
		{1.0, 0.625},  //   -1,  1, -1
		{0.0, 0.375},  //   -1, -1, -1

		{0.25, 0.375}, //   -1, -1,  1
		{0.5, 0.375},  //    1, -1,  1

		{0.75, 0.375}, //    1, -1, -1
		{1.0, 0.375},  //   -1, -1, -1
		{0.25, 0.125}, //   -1, -1, -1
		{0.5, 0.125},  //    1, -1, -1
	}

	/* Cube vertices are layed out like this:

	  +--------+           6          5
	/ |       /|
	+--------+ |        1          0        +Y
	| |      | |                            |___ +X
	| +------|-+           7          4    /
	|/       |/                           +Z
	+--------+          2          3

	Which makes for the following upwrapped uv locations:

	Top: {6, 5, 1, 0} -> [0, 1, 3, 4]
	Front: {1, 0, 2, 3} -> [3, 4, 8, 9]
	Left: {6, 1, 7, 2} -> [2, 3, 7, 8]
	Bottom: {2, 3, 7, 4} -> [8, 9, 12, 13]
	Right: {0, 5, 3, 4} -> [4, 5, 9, 10]
	Back: {5, 6, 4, 7} -> [5, 6, 10, 11]
	*/

	indexes := [24]int{
		0, 1, 3, 4,
		3, 4, 8, 9,
		2, 3, 7, 8,
		8, 9, 12, 13,
		4, 5, 9, 10,
		5, 6, 10, 11,
	}

	// get the uv's to map to
	topLeft := uvs[indexes[side*4]]
	topRight := uvs[indexes[side*4+1]]
	botLeft := uvs[indexes[side*4+2]]
	//botRight := uvs[indexes[side*4+3]]

	sScale := topRight[0] - topLeft[0]
	tScale := topLeft[1] - botLeft[1]

	return botLeft[0] + (s * sScale), botLeft[1] + (t * tScale)
}
