// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"bytes"
	"fmt"
	"io/ioutil"

	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/groggy"
)

// RenderShader is an OpenGL shader that is used for easier access
// to uniforms and attributes at runtime.
type RenderShader struct {
	// Prog is the OpenGL program associated with the RenderShader.
	Prog graphics.Program

	// uniCache is the cache of uniform locations.
	uniCache map[string]int32

	// attrCache is the cache of attribute locations.
	attrCache map[string]int32
}

// NewRenderShader creates a new RenderShader object with the OpenGL shader specified.
func NewRenderShader(p graphics.Program) *RenderShader {
	rs := new(RenderShader)
	rs.Prog = p
	rs.uniCache = make(map[string]int32)
	rs.attrCache = make(map[string]int32)
	return rs
}

// GetUniformLocation gets the location of a uniform variable, preferably from
// an internal cached value stored in a map.
func (rs *RenderShader) GetUniformLocation(name string) int32 {
	// attempt to get it from the cache first
	ul, found := rs.uniCache[name]
	if found {
		return ul
	}

	// pull the location from the shader and cache it
	ul = gfx.GetUniformLocation(rs.Prog, name)

	// cache even if it returns -1 so that it doesn't repeatedly check
	rs.uniCache[name] = ul
	return ul
}

// AssertUniformsExist attempts to get uniforms for the names passed in and returns
// an error value if a name doesn't exist.
func (rs *RenderShader) AssertUniformsExist(names ...string) error {
	const badUniformLocation int32 = -1

	for _, name := range names {
		ul := rs.GetUniformLocation(name)
		if ul == badUniformLocation {
			return fmt.Errorf("ASSERT FAILED: Shader uniform %s doesn't exist", name)
		}
	}

	return nil
}

// GetAttribLocation gets the location of a attribute variable, preferably from
// an internal cached value stored in a map.
func (rs *RenderShader) GetAttribLocation(name string) int32 {
	// attempt to get it from the cache first
	al, found := rs.attrCache[name]
	if found {
		return al
	}

	// pull the location from the shader and cache it
	al = gfx.GetAttribLocation(rs.Prog, name)

	// cache even if it returns -1 so that it doesn't repeatedly check
	rs.attrCache[name] = al
	return al
}

// AssertAttribsExist attempts to get attributes for the names passed in and returns
// an error value if a name doesn't exist.
func (rs *RenderShader) AssertAttribsExist(names ...string) error {
	const badAttributeLocation int32 = -1

	for _, name := range names {
		al := rs.GetAttribLocation(name)
		if al == badAttributeLocation {
			return fmt.Errorf("ASSERT FAILED: Shader uniform %s doesn't exist", name)
		}
	}

	return nil
}

// Destroy deallocates the shader from OpenGL.
func (rs *RenderShader) Destroy() {
	gfx.DeleteProgram(rs.Prog)
}

// PreLinkBinder is a prototype for a function to be called before a shader program is linked
type PreLinkBinder func(p graphics.Program)

// LoadShaderProgramFromFiles loads the GLSL shaders from the files specified. This function
// expects that the vertex and fragment shader files can be opened by appending the '.vs' and '.fs'
// extensions respectively to the baseFilename. preLink is an optional function that will be
// called just prior to linking the shaders into a program.
func LoadShaderProgramFromFiles(baseFilename string, prelink PreLinkBinder) (*RenderShader, error) {
	vsBytes, err := ioutil.ReadFile(baseFilename + ".vs")
	if err != nil {
		fmt.Errorf("Failed to read the vertex shader \"%s.vs\".\n%v", baseFilename, err)
	}
	vsBuffer := bytes.NewBuffer(vsBytes)

	fsBytes, err := ioutil.ReadFile(baseFilename + ".fs")
	if err != nil {
		fmt.Errorf("Failed to read the fragment shader \"%s.fs\".\n%v", baseFilename, err)
	}
	fsBuffer := bytes.NewBuffer(fsBytes)

	groggy.Logsf("DEBUG", "Compiling shader: %s.", baseFilename)
	return LoadShaderProgram(vsBuffer.String(), fsBuffer.String(), prelink)
}

// LoadShaderProgram loads shaders from code passed in as strings, compiles and then attaches them to a new program.
// preLink is an optional function that will be called just prior to linking the shaders into a program.
func LoadShaderProgram(vertShader, fragShader string, prelink PreLinkBinder) (*RenderShader, error) {
	// create the program
	prog := gfx.CreateProgram()

	// create the vertex shader
	var status int32
	vs := gfx.CreateShader(graphics.VERTEX_SHADER)
	gfx.ShaderSource(vs, vertShader)
	gfx.CompileShader(vs)
	gfx.GetShaderiv(vs, graphics.COMPILE_STATUS, &status)
	if status == graphics.FALSE {
		log := gfx.GetShaderInfoLog(vs)
		return nil, fmt.Errorf("Failed to compile the vertex shader:\n%s", log)
	}
	defer gfx.DeleteShader(vs)

	// create the fragment shader
	fs := gfx.CreateShader(graphics.FRAGMENT_SHADER)
	gfx.ShaderSource(fs, fragShader)
	gfx.CompileShader(fs)
	gfx.GetShaderiv(fs, graphics.COMPILE_STATUS, &status)
	if status == graphics.FALSE {
		log := gfx.GetShaderInfoLog(fs)
		return nil, fmt.Errorf("Failed to compile the fragment shader:\n%s", log)
	}
	defer gfx.DeleteShader(fs)

	// call the prelinker if supplied
	if prelink != nil {
		prelink(prog)
	}

	// attach the shaders to the program and link
	gfx.AttachShader(prog, vs)
	gfx.AttachShader(prog, fs)
	gfx.LinkProgram(prog)
	gfx.GetProgramiv(prog, graphics.LINK_STATUS, &status)
	if status == graphics.FALSE {
		log := gfx.GetProgramInfoLog(prog)
		return nil, fmt.Errorf("Failed to link the program!\n%s", log)
	}

	rs := NewRenderShader(prog)
	return rs, nil
}
