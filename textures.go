// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

func loadFile(filePath string) (rgbaFFlipped *image.NRGBA, e error) {
	imgFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open the texture file: %v\n", err)
	}

	img, err := png.Decode(imgFile)
	imgFile.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed to decode the texture: %v\n", err)
	}

	// if the source image doesn't have alpha, set it manually
	b := img.Bounds()
	rgba := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Bounds(), img, b.Min, draw.Src)

	return rgba, nil

	// flip the image vertically
	// NOTE: I guess we don't need to do this anymore ...
	/*
		rows := b.Max.Y
		rgba_flipped = image.NewNRGBA(image.Rect(0, 0, b.Max.X, b.Max.Y))
		for dy := 0; dy < rows; dy++ {
			sy := b.Max.Y - dy - 1
			for dx := 0; dx < b.Max.X; dx++ {
				soffset := sy*rgba.Stride + dx*4
				doffset := dy*rgba_flipped.Stride + dx*4
				copy(rgba_flipped.Pix[doffset:doffset+4], rgba.Pix[soffset:soffset+4])
			}
		}
		return rgba_flipped, nil
	*/
}

// LoadRGBAToTexture takes a byte slice and throws it into an OpenGL texture.
func LoadRGBAToTexture(rgba []byte, imageSize int32) graphics.Texture {
	tex := gfx.GenTexture()
	gfx.ActiveTexture(graphics.TEXTURE0)
	gfx.BindTexture(graphics.TEXTURE_2D, tex)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.REPEAT)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.REPEAT)
	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.RGBA, imageSize, imageSize, 0, graphics.RGBA, graphics.UNSIGNED_BYTE, gfx.Ptr(rgba))
	return tex
}

// LoadImageToTexture loads an image from a file into an OpenGL texture.
func LoadImageToTexture(filePath string) (graphics.Texture, error) {
	tex := gfx.GenTexture()
	gfx.ActiveTexture(graphics.TEXTURE0)
	gfx.BindTexture(graphics.TEXTURE_2D, tex)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.REPEAT)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.REPEAT)

	rgbaFlipped, err := loadFile(filePath)
	if err != nil {
		return tex, err
	}

	imageSize := int32(rgbaFlipped.Bounds().Max.X)

	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.RGBA, imageSize, imageSize, 0, graphics.RGBA, graphics.UNSIGNED_BYTE, gfx.Ptr(rgbaFlipped.Pix))
	return tex, nil
}

// LoadImagesToTextureArray loads image files and buffers them into a new GL texture.
func LoadImagesToTextureArray(filepaths map[string]string, size int32) (texArray graphics.Texture, texLoc map[string]int32, e error) {
	// I thought this could be used for mipmap generation, but it causes crashes on some
	// Intel drivers.
	const levels int32 = 1

	// make the map that will hold the locations for the textures in the array
	texLoc = make(map[string]int32)

	// generate the texture array
	texArray = gfx.GenTexture()
	gfx.BindTexture(graphics.TEXTURE_2D_ARRAY, texArray)

	// create the texture array with the specified number of levels that's big enough
	// to fit all of the textures specified in the filepaths parameter.
	gfx.TexStorage3D(graphics.TEXTURE_2D_ARRAY, levels, graphics.RGBA8, size, size, int32(len(filepaths)))

	// for each texture listed in filepaths
	var arrayIndex int32
	for texName, filePath := range filepaths {
		rgbaFlipped, err := loadFile(filePath)
		if err != nil {
			return texArray, texLoc, fmt.Errorf("Failed to load the PNG file into an image.\n%v\n", err)
		}

		const byteDepth int32 = 1
		gfx.TexSubImage3D(graphics.TEXTURE_2D_ARRAY, 0, 0, 0, arrayIndex, size, size, byteDepth, graphics.RGBA, graphics.UNSIGNED_BYTE, gfx.Ptr(rgbaFlipped.Pix))

		// store the array index in a map so that we can access it correctly later
		texLoc[texName] = arrayIndex
		arrayIndex++
	}

	if levels != 1 {
		gfx.GenerateMipmap(graphics.TEXTURE_2D_ARRAY)
	}

	return texArray, texLoc, nil
}
