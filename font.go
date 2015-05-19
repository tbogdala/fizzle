// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

/*
Based primarily on gltext found at https://github.com/go-gl/gltext
But also based on examples from the freetype-go project:
  code.google.com/p/freetype-go/freetype

This implementation differs in the way the images are rendered and then
copied into an OpenGL texture. In addition to that, this module can
create a renderable 'string' node which is a bunch of polygons with uv's
mapped to the appropriate glyphs.
*/

import (
	ft "code.google.com/p/freetype-go/freetype"
	"errors"
	"fmt"
	gl "github.com/go-gl/gl/v3.3-core/gl"
	"image"
	"image/draw"
	"io/ioutil"
	"os"
)

// runeData stores information pulled from the freetype parsing of glyphs.
type runeData struct {
	imgX, imgY                    int // offset into the image texture for the top left position of rune
	advanceWidth, leftSideBearing int // HMetric data from glyph
	advanceHeight, topSideBearing int // VMetric data from glyph
	uvMinX, uvMinY                float32
	uvMaxX, uvMaxY                float32
}

// GLFont contains data regarding a font and the texture that was created
// with the specified set of glyphs. It can then be used to create
// renderable string objects.
type GLFont struct {
	Texture     uint32
	TextureSize int
	Glyphs      string
	GlyphHeight int
	GlyphWidth  int
	locations   map[rune]runeData
	Shader      *RenderShader
}

// NewGLFont takes a fontFilepath and uses the Go freetype library to parse it
// and render the specified glyphs to a texture that is then buffered into OpenGL.
func NewGLFont(fontFilepath string, scale int32, glyphs string) (f *GLFont, e error) {
	f = new(GLFont)

	// allocate the location map
	f.locations = make(map[rune]runeData)

	// Load the font used for UI interaction
	fontFile, err := os.Open(fontFilepath)
	if err != nil {
		return f, fmt.Errorf("Failed to open the font file.\n%v", err)
	}
	defer fontFile.Close()

	// load in the font
	fontBytes, err := ioutil.ReadAll(fontFile)
	if err != nil {
		return f, fmt.Errorf("Failed to load font data from stream.\n%v", err)
	}

	// parse the truetype font data
	ttfData, err := ft.ParseFont(fontBytes)
	if err != nil {
		return f, fmt.Errorf("Failed to prase the truetype font data.\n%v", err)
	}

	// this may have negative components, but get the bounds for the font
	glyphBounds := ttfData.Bounds(scale)

	// width and height are getting +2 here since the glyph will be buffered by a
	// pixel in the texture
	glyphWidth := int(glyphBounds.XMax-glyphBounds.XMin) + 4
	glyphHeight := int(glyphBounds.YMax-glyphBounds.YMin) + 4

	// create the buffer image used to draw the glyphs
	glyphRect := image.Rect(0, 0, glyphWidth, glyphHeight)
	glyphImg := image.NewRGBA(glyphRect)

	// calculate the area needed for the font texture
	var fontTexSize int = 2
	minAreaNeeded := (glyphWidth) * (glyphHeight) * len(glyphs)
	for (fontTexSize * fontTexSize) < minAreaNeeded {
		fontTexSize *= 2
		if fontTexSize > 2048 {
			return f, errors.New("Font texture was going to exceed 2048x2048 and that's currently not supported.")
		}
	}

	// create the font image
	fontImgRect := image.Rect(0, 0, fontTexSize, fontTexSize)
	fontImg := image.NewRGBA(fontImgRect)

	// the number of glyphs
	fontRowSize := fontTexSize/(glyphWidth) - 1

	// create the freetype context
	c := ft.NewContext()
	c.SetDPI(72)
	c.SetFont(ttfData)
	c.SetFontSize(float64(scale))
	c.SetClip(glyphImg.Bounds())
	c.SetDst(glyphImg)
	c.SetSrc(image.White)

	var fx, fy int
	fixedPoint := int(c.PointToFix32(float64(scale)) >> 8)
	for _, ch := range glyphs {
		index := ttfData.Index(ch)
		metricH := ttfData.HMetric(scale, index)
		metricV := ttfData.VMetric(scale, index)

		fxGW := fx * glyphWidth
		fyGH := fy * glyphHeight

		f.locations[ch] = runeData{
			fxGW, fyGH,
			int(metricH.AdvanceWidth), int(metricH.LeftSideBearing),
			int(metricV.AdvanceHeight), int(metricV.TopSideBearing),
			float32(fxGW) / float32(fontTexSize), float32(fyGH+glyphHeight) / float32(fontTexSize),
			float32(fxGW+glyphWidth) / float32(fontTexSize), float32(fyGH) / float32(fontTexSize),
		}

		pt := ft.Pt(1, 1+fixedPoint)
		c.DrawString(string(ch), pt)

		// copy the glyph image into the font image
		for subY := 0; subY < glyphHeight; subY++ {
			for subX := 0; subX < glyphWidth; subX++ {
				glyphRGBA := glyphImg.RGBAAt(subX, subY)
				fontImg.SetRGBA((fxGW)+subX, (fyGH)+subY, glyphRGBA)
			}
		}

		// erase the glyph image buffer
		draw.Draw(glyphImg, glyphImg.Bounds(), image.Transparent, image.ZP, draw.Src)

		// adjust the pointers into the font image
		fx++
		if fx > fontRowSize {
			fx = 0
			fy++
		}

	}

	// buffer the font image into an OpenGL texture
	f.Texture = LoadRGBAToTexture(fontImg.Pix, int32(fontImg.Rect.Max.X))
	f.Glyphs = glyphs
	f.TextureSize = fontTexSize
	f.GlyphWidth = glyphWidth
	f.GlyphHeight = glyphHeight
	return
}

func setVertFloats(fa []float32, vertIdx int, x0, y0, x1, y1 float32) {
	fa[vertIdx+0] = x1
	fa[vertIdx+1] = y0
	fa[vertIdx+2] = 0.0

	fa[vertIdx+3] = x1
	fa[vertIdx+4] = y1
	fa[vertIdx+5] = 0.0

	fa[vertIdx+6] = x0
	fa[vertIdx+7] = y1
	fa[vertIdx+8] = 0.0

	fa[vertIdx+9] = x0
	fa[vertIdx+10] = y0
	fa[vertIdx+11] = 0.0
}

func setUvFloats(fa []float32, uvIdx int, s0, t0, s1, t1 float32) {
	fa[uvIdx+0] = s1
	fa[uvIdx+1] = t0

	fa[uvIdx+2] = s1
	fa[uvIdx+3] = t1

	fa[uvIdx+4] = s0
	fa[uvIdx+5] = t1

	fa[uvIdx+6] = s0
	fa[uvIdx+7] = t0
}

func setFaceInts(fa []uint32, faceIdx int, startIndex uint32) {

	fa[faceIdx+0] = startIndex
	fa[faceIdx+1] = startIndex + 1
	fa[faceIdx+2] = startIndex + 2

	fa[faceIdx+3] = startIndex + 2
	fa[faceIdx+4] = startIndex + 3
	fa[faceIdx+5] = startIndex
}

// Destroy releases the OpenGL texture for the font but does
// not release the associated shader.
func (f *GLFont) Destroy() {
	gl.DeleteTextures(1, &f.Texture)
}

// CreateLabel makes a new renderable object from the supplied string
// using the data in the font.
func (f *GLFont) CreateLabel(msg string) *Renderable {
	msgLength := len(msg)

	const vertsPerChar int = 4
	const floatsPerUv int = 2
	const floatsPerVert int = 3
	const intsPerFace int = 3
	const facesPerChar int = 2
	const vertFloatsPerChar int = vertsPerChar * floatsPerVert
	const uvFloatsPerChar int = vertsPerChar * floatsPerUv
	const facesIntsPerChar int = facesPerChar * intsPerFace
	const indexesPerChar uint32 = 4

	// create the arrays to hold the data to buffer to OpenGL
	stringVerts := make([]float32, msgLength*vertFloatsPerChar)
	stringUVs := make([]float32, msgLength*uvFloatsPerChar)
	stringIndexes := make([]uint32, msgLength*facesIntsPerChar)

	// loop through the message
	var pen_x float32
	for chi, ch := range msg {
		// get the rune data
		chData := f.locations[ch]

		// setup the coordinates for ther vetexes
		x0 := pen_x
		y0 := 2.0 - (float32(f.GlyphHeight) - float32(chData.topSideBearing))
		x1 := x0 + float32(f.GlyphWidth)
		y1 := y0 + float32(f.GlyphHeight)
		s0 := chData.uvMinX
		t0 := chData.uvMinY
		s1 := chData.uvMaxX
		t1 := chData.uvMaxY

		// set the vertex data
		setVertFloats(stringVerts, chi*vertFloatsPerChar, x0, y0, x1, y1)
		setUvFloats(stringUVs, chi*uvFloatsPerChar, s0, t0, s1, t1)
		setFaceInts(stringIndexes, chi*facesIntsPerChar, uint32(chi)*indexesPerChar)

		// advance the pen
		pen_x += float32(chData.advanceWidth)
	}

	// create the renderable object
	r := NewRenderable()
	r.FaceCount = uint32(msgLength * 2)
	r.Core.Tex0 = f.Texture
	r.Core.Shader = f.Shader
	r.BoundingRect = GetBoundingRect(stringVerts)

	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4

	// create a VBO to hold the vertex data
	gl.GenBuffers(1, &r.Core.VertVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.VertVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(stringVerts), gl.Ptr(stringVerts), gl.STATIC_DRAW)

	// create a VBO to hold the uv data
	gl.GenBuffers(1, &r.Core.UvVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.Core.UvVBO)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(stringUVs), gl.Ptr(stringUVs), gl.STATIC_DRAW)

	// create a VBO to hold the face indexes
	gl.GenBuffers(1, &r.Core.ElementsVBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, uintSize*len(stringIndexes), gl.Ptr(stringIndexes), gl.STATIC_DRAW)

	return r
}
