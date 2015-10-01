// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"fmt"
	gl "github.com/go-gl/gl/v3.3-core/gl"
	"image"
	"image/draw"
	"image/png"
	"os"
)

func loadFile(filePath string) (rgba_flipped *image.NRGBA, e error) {
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

func LoadRGBAToTexture(rgba []byte, imageSize int32) uint32 {
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, imageSize, imageSize, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba))
	return tex
}

func LoadImageToTexture(filePath string) (glTex uint32, e error) {
	gl.GenTextures(1, &glTex)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, glTex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	rgba_flipped, err := loadFile(filePath)
	if err != nil {
		return glTex, err
	}

	imageSize := int32(rgba_flipped.Bounds().Max.X)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, imageSize, imageSize, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba_flipped.Pix))
	return glTex, nil
}

// LoadImageToTexture loads an image file and buffers it into a new GL texture.
func LoadImagesToTextureArray(filepaths map[string]string, size int32) (texArray uint32, texLoc map[string]int32, e error) {
	// I thought this could be used for mipmap generation, but it causes crashes on some
	// Intel drivers.
	const levels int32 = 1

	// make the map that will hold the locations for the textures in the array
	texLoc = make(map[string]int32)

	// generate the texture array
	gl.GenTextures(1, &texArray)
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, texArray)

	// create the texture array with the specified number of levels that's big enough
	// to fit all of the textures specified in the filepaths parameter.
	gl.TexStorage3D(gl.TEXTURE_2D_ARRAY, levels, gl.RGBA8, size, size, int32(len(filepaths)))

	// for each texture listed in filepaths
	var arrayIndex int32
	for texName, filePath := range filepaths {
		rgba_flipped, err := loadFile(filePath)
		if err != nil {
			return texArray, texLoc, fmt.Errorf("Failed to load the PNG file into an image.\n%v\n", err)
		}

		const byteDepth int32 = 1
		gl.TexSubImage3D(gl.TEXTURE_2D_ARRAY, 0, 0, 0, arrayIndex, size, size, byteDepth, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba_flipped.Pix))

		// store the array index in a map so that we can access it correctly later
		texLoc[texName] = arrayIndex
		arrayIndex += 1
	}

	if levels != 1 {
		gl.GenerateMipmap(gl.TEXTURE_2D_ARRAY)
	}

	return texArray, texLoc, nil
}
