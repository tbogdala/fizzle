// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	"bytes"
	"errors"
	"fmt"
	gl "github.com/go-gl/gl/v3.3-core/gl"
	"github.com/tbogdala/groggy"
	"io/ioutil"
	"strings"
)

type RenderShader struct {
	Prog      uint32
	uniCache  map[string]int32
	attrCache map[string]int32
}

func NewRenderShader(p uint32) *RenderShader {
	rs := new(RenderShader)
	rs.Prog = p
	rs.uniCache = make(map[string]int32)
	rs.attrCache = make(map[string]int32)
	return rs
}

func (rs *RenderShader) GetUniformLocation(name string) int32 {
	// attempt to get it from the cache first
	ul, found := rs.uniCache[name]
	if found {
		return ul
	}

	// pull the location from the shader and cache it
	uniGLName := name + "\x00"
	ul = gl.GetUniformLocation(rs.Prog, gl.Str(uniGLName))

	// cache even if it returns -1 so that it doesn't repeatedly check
	rs.uniCache[name] = ul
	return ul
}

func (rs *RenderShader) AssertUniformsExist(names ...string) error {
	const badUniformLocation int32 = -1

	for _, name := range names {
		ul := rs.GetUniformLocation(name)
		if ul == badUniformLocation {
			return fmt.Errorf("ASSERT FAILED: Shader uniform %s doesn't exist.", name)
		}
	}

	return nil
}

func (rs *RenderShader) GetAttribLocation(name string) int32 {
	// attempt to get it from the cache first
	al, found := rs.attrCache[name]
	if found {
		return al
	}

	// pull the location from the shader and cache it
	attrGLName := name + "\x00"
	al = gl.GetAttribLocation(rs.Prog, gl.Str(attrGLName))

	// cache even if it returns -1 so that it doesn't repeatedly check
	rs.attrCache[name] = al
	return al
}

func (rs *RenderShader) AssertAttribsExist(names ...string) error {
	const badAttributeLocation int32 = -1

	for _, name := range names {
		al := rs.GetAttribLocation(name)
		if al == badAttributeLocation {
			return fmt.Errorf("ASSERT FAILED: Shader uniform %s doesn't exist.", name)
		}
	}
	
	return nil
}

func (rs *RenderShader) Destroy() {
	gl.DeleteProgram(rs.Prog)
}

// PreLinkBinder is a prototype for a function to be called before a shader program is linked
type PreLinkBinder func(p uint32)

// LoadShaderProgramFromFiles loads the glsl shaders from the files specified.
func LoadShaderProgramFromFiles(baseFilename string, prelink PreLinkBinder) (*RenderShader, error) {
	vsBytes, err := ioutil.ReadFile(baseFilename + ".vs")
	if err != nil {
		fmt.Errorf("Failed to read the vertex shader \"%s\".\n%v\n", baseFilename+".vs", err)
	}
	vsBuffer := bytes.NewBuffer(vsBytes)

	fsBytes, err := ioutil.ReadFile(baseFilename + ".fs")
	if err != nil {
		fmt.Errorf("Failed to read the fragment shader \"%s\".\n%v\n", baseFilename+".fs", err)
	}
	fsBuffer := bytes.NewBuffer(fsBytes)

	groggy.Logsf("DEBUG", "Compiling shader: %s.", baseFilename)
	return LoadShaderProgram(vsBuffer.String(), fsBuffer.String(), prelink)
}

// LoadShaderProgram loads shader objects, compiles and then attaches them to a new program
func LoadShaderProgram(vertShader, fragShader string, prelink PreLinkBinder) (*RenderShader, error) {
	// create the program
	prog := gl.CreateProgram()

	// create the vertex shader
	vs := gl.CreateShader(gl.VERTEX_SHADER)
	cVertShader := gl.Str(vertShader + "\x00")
	gl.ShaderSource(vs, 1, &cVertShader, nil)
	gl.CompileShader(vs)

	var status int32
	gl.GetShaderiv(vs, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(vs, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(vs, logLength, nil, gl.Str(log))

		err := fmt.Sprintf("Failed to compile the vertex shader!\n%s", log)
		fmt.Println(err)
		return nil, errors.New(err)
	}
	defer gl.DeleteShader(vs)

	// create the fragment shader
	fs := gl.CreateShader(gl.FRAGMENT_SHADER)
	cFragShader := gl.Str(fragShader + "\x00")
	gl.ShaderSource(fs, 1, &cFragShader, nil)
	gl.CompileShader(fs)

	gl.GetShaderiv(fs, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fs, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fs, logLength, nil, gl.Str(log))

		err := fmt.Sprintf("Failed to compile the fragment shader!\n%s", log)
		fmt.Println(err)
		return nil, errors.New(err)
	}
	defer gl.DeleteShader(fs)

	// call the prelinker if supplied
	if prelink != nil {
		prelink(prog)
	}

	// attach the shaders to the program and link
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)
	gl.LinkProgram(prog)

	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(prog, logLength, nil, gl.Str(log))

		error := fmt.Sprintf("Failed to link the program!\n%s", log)
		fmt.Println(error)
		return nil, errors.New(error)
	}

	rs := NewRenderShader(prog)

	return rs, nil
}
