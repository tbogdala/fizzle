// Copyright 2015, Timothy` Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package opengl

import (
	"fmt"
	"strings"
	"unsafe"

	gl "github.com/go-gl/gl/v3.3-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

// GraphicsImpl is the graphics provider for the desktop
// implementation of OpenGL.
type GraphicsImpl struct {
	// currently nothing in use
}

// InitOpenGL initializes the OpenGL graphics provider and
// sets it to be the current provider for the module.
func InitOpenGL() (*GraphicsImpl, error) {
	gp := new(GraphicsImpl)

	// make sure that all of the GL functions are initialized
	err := gl.Init()
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize GL! %v", err)
	}

	return gp, nil
}

// ActiveTexture selects the active texture unit
func (impl *GraphicsImpl) ActiveTexture(t graphics.Texture) {
	gl.ActiveTexture(uint32(t))
}

// AttachShader attaches a shader object to a program object
func (impl *GraphicsImpl) AttachShader(p graphics.Program, s graphics.Shader) {
	gl.AttachShader(uint32(p), uint32(s))
}

// BindBuffer binds a buffer to the OpenGL target specified by enum
func (impl *GraphicsImpl) BindBuffer(target graphics.Enum, b graphics.Buffer) {
	gl.BindBuffer(uint32(target), uint32(b))
}

// BindFragDataLocation binds a user-defined varying out variable
// to a fragment shader color number
func (impl *GraphicsImpl) BindFragDataLocation(p graphics.Program, color uint32, name string) {
	// name has to be zero terminated for gl.Str()
	glName := name + "\x00"
	gl.BindFragDataLocation(uint32(p), color, gl.Str(glName))
}

// BindFramebuffer binds a framebuffer to a framebuffer target
func (impl *GraphicsImpl) BindFramebuffer(target graphics.Enum, fb graphics.Buffer) {
	gl.BindFramebuffer(uint32(target), uint32(fb))
}

// BindRenderbuffer binds a renderbuffer to a renderbuffer target
func (impl *GraphicsImpl) BindRenderbuffer(target graphics.Enum, renderbuffer graphics.Buffer) {
	gl.BindRenderbuffer(uint32(target), uint32(renderbuffer))
}

// BindTexture binds a texture to the OpenGL target specified by enum
func (impl *GraphicsImpl) BindTexture(target graphics.Enum, t graphics.Texture) {
	gl.BindTexture(uint32(target), uint32(t))
}

// BindVertexArray binds a vertex array object
func (impl *GraphicsImpl) BindVertexArray(a uint32) {
	gl.BindVertexArray(a)
}

// BlendEquation specifies the equation used for both the RGB and
// alpha blend equations
func (impl *GraphicsImpl) BlendEquation(mode graphics.Enum) {
	gl.BlendEquation(uint32(mode))
}

// BlendFunc specifies the pixel arithmetic for the blend fucntion
func (impl *GraphicsImpl) BlendFunc(sFactor, dFactor graphics.Enum) {
	gl.BlendFunc(uint32(sFactor), uint32(dFactor))
}

// BufferData creates a new data store for the bound buffer object.
func (impl *GraphicsImpl) BufferData(target graphics.Enum, size int, data unsafe.Pointer, usage graphics.Enum) {
	gl.BufferData(uint32(target), size, data, uint32(usage))
}

// CheckFramebufferStatus checks the completeness status of a framebuffer
func (impl *GraphicsImpl) CheckFramebufferStatus(target graphics.Enum) graphics.Enum {
	return graphics.Enum(gl.CheckFramebufferStatus(uint32(target)))
}

// Clear clears the window buffer specified in mask
func (impl *GraphicsImpl) Clear(mask graphics.Enum) {
	gl.Clear(uint32(mask))
}

// ClearColor specifies the RGBA value used to clear the color buffers
func (impl *GraphicsImpl) ClearColor(red, green, blue, alpha float32) {
	gl.ClearColor(red, green, blue, alpha)
}

// CompileShader compiles the shader object
func (impl *GraphicsImpl) CompileShader(s graphics.Shader) {
	gl.CompileShader(uint32(s))
}

// CreateProgram creates a new shader program object
func (impl *GraphicsImpl) CreateProgram() graphics.Program {
	return graphics.Program(gl.CreateProgram())
}

// CreateShader creates a new shader object
func (impl *GraphicsImpl) CreateShader(ty graphics.Enum) graphics.Shader {
	return graphics.Shader(gl.CreateShader(uint32(ty)))
}

// CullFace specifies whether to use front or back face culling
func (impl *GraphicsImpl) CullFace(mode graphics.Enum) {
	gl.CullFace(uint32(mode))
}

// DeleteBuffer deletes the OpenGL buffer object
func (impl *GraphicsImpl) DeleteBuffer(b graphics.Buffer) {
	uintV := uint32(b)
	gl.DeleteBuffers(1, &uintV)
}

// DeleteFramebuffer deletes the framebuffer object
func (impl *GraphicsImpl) DeleteFramebuffer(fb graphics.Buffer) {
	uintV := uint32(fb)
	gl.DeleteFramebuffers(1, &uintV)
}

// DeleteProgram deletes the shader program object
func (impl *GraphicsImpl) DeleteProgram(p graphics.Program) {
	gl.DeleteProgram(uint32(p))
}

// DeleteRenderbuffer deletes the renderbuffer object
func (impl *GraphicsImpl) DeleteRenderbuffer(rb graphics.Buffer) {
	uintV := uint32(rb)
	gl.DeleteRenderbuffers(1, &uintV)
}

// DeleteShader deletes the shader object
func (impl *GraphicsImpl) DeleteShader(s graphics.Shader) {
	gl.DeleteShader(uint32(s))
}

// DeleteTexture deletes the specified texture
func (impl *GraphicsImpl) DeleteTexture(v graphics.Texture) {
	uintV := uint32(v)
	gl.DeleteTextures(1, &uintV)
}

// DeleteVertexArray deletes an OpenGL VAO
func (impl *GraphicsImpl) DeleteVertexArray(a uint32) {
	uintV := uint32(a)
	gl.DeleteVertexArrays(1, &uintV)
}

// DepthMask enables or disables writing into the depth buffer
func (impl *GraphicsImpl) DepthMask(flag bool) {
	gl.DepthMask(flag)
}

// Disable disables various GL capabilities.
func (impl *GraphicsImpl) Disable(e graphics.Enum) {
	gl.Disable(uint32(e))
}

// DrawBuffers specifies a list of color buffers to be drawn into
func (impl *GraphicsImpl) DrawBuffers(buffers []uint32) {
	c := int32(len(buffers))
	gl.DrawBuffers(c, &buffers[0])
}

// DrawElements renders primitives from array data
func (impl *GraphicsImpl) DrawElements(mode graphics.Enum, count int32, ty graphics.Enum, indices unsafe.Pointer) {
	gl.DrawElements(uint32(mode), count, uint32(ty), indices)
}

// Enable enables various GL capabilities.
func (impl *GraphicsImpl) Enable(e graphics.Enum) {
	gl.Enable(uint32(e))
}

// EnableVertexAttribArray enables a vertex attribute array
func (impl *GraphicsImpl) EnableVertexAttribArray(a uint32) {
	gl.EnableVertexAttribArray(a)
}

// FramebufferRenderbuffer attaches a renderbuffer as a logical buffer
// of a framebuffer object
func (impl *GraphicsImpl) FramebufferRenderbuffer(target, attachment, renderbuffertarget graphics.Enum, renderbuffer graphics.Buffer) {
	gl.FramebufferRenderbuffer(uint32(target), uint32(attachment), uint32(renderbuffertarget), uint32(renderbuffer))
}

// FramebufferTexture2D attaches a texture object to a framebuffer
func (impl *GraphicsImpl) FramebufferTexture2D(target, attachment, textarget graphics.Enum, texture graphics.Texture, level int32) {
	gl.FramebufferTexture2D(uint32(target), uint32(attachment), uint32(textarget), uint32(texture), level)
}

// GenBuffer creates an OpenGL buffer object
func (impl *GraphicsImpl) GenBuffer() graphics.Buffer {
	var b uint32
	gl.GenBuffers(1, &b)
	return graphics.Buffer(b)
}

// GenerateMipmap generates mipmaps for a specified texture target
func (impl *GraphicsImpl) GenerateMipmap(t graphics.Enum) {
	gl.GenerateMipmap(uint32(t))
}

// GenFramebuffer generates a OpenGL framebuffer object
func (impl *GraphicsImpl) GenFramebuffer() graphics.Buffer {
	var b uint32
	gl.GenFramebuffers(1, &b)
	return graphics.Buffer(b)
}

// GenRenderbuffer generates a OpenGL renderbuffer object
func (impl *GraphicsImpl) GenRenderbuffer() graphics.Buffer {
	var b uint32
	gl.GenRenderbuffers(1, &b)
	return graphics.Buffer(b)
}

// GenTexture creates an OpenGL texture object
func (impl *GraphicsImpl) GenTexture() graphics.Texture {
	var t uint32
	gl.GenTextures(1, &t)
	return graphics.Texture(t)
}

// GenVertexArray creates an OpoenGL VAO
func (impl *GraphicsImpl) GenVertexArray() uint32 {
	var a uint32
	gl.GenVertexArrays(1, &a)
	return a
}

// GetAttribLocation returns the location of a attribute variable
func (impl *GraphicsImpl) GetAttribLocation(p graphics.Program, name string) int32 {
	glName := name + "\x00"
	return gl.GetAttribLocation(uint32(p), gl.Str(glName))
}

// GetError returns the next error
func (impl *GraphicsImpl) GetError() uint32 {
	return gl.GetError()
}

// GetProgramInfoLog returns the information log for a program object
func (impl *GraphicsImpl) GetProgramInfoLog(p graphics.Program) string {
	var logLength int32
	impl.GetProgramiv(p, graphics.INFO_LOG_LENGTH, &logLength)

	// make sure the string is zero'd out to start with
	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetProgramInfoLog(uint32(p), logLength, nil, gl.Str(log))

	return log
}

// GetProgramiv returns a parameter from the program object
func (impl *GraphicsImpl) GetProgramiv(p graphics.Program, pname graphics.Enum, params *int32) {
	gl.GetProgramiv(uint32(p), uint32(pname), params)
}

// GetShaderInfoLog returns the information log for a shader object
func (impl *GraphicsImpl) GetShaderInfoLog(s graphics.Shader) string {
	var logLength int32
	impl.GetShaderiv(s, graphics.INFO_LOG_LENGTH, &logLength)

	// make sure the string is zero'd out to start with
	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetShaderInfoLog(uint32(s), logLength, nil, gl.Str(log))

	return log
}

// GetShaderiv returns a parameter from the shader object
func (impl *GraphicsImpl) GetShaderiv(s graphics.Shader, pname graphics.Enum, params *int32) {
	gl.GetShaderiv(uint32(s), uint32(pname), params)
}

// GetUniformLocation returns the location of a uniform variable
func (impl *GraphicsImpl) GetUniformLocation(p graphics.Program, name string) int32 {
	glName := name + "\x00"
	return gl.GetUniformLocation(uint32(p), gl.Str(glName))
}

// LinkProgram links a program object
func (impl *GraphicsImpl) LinkProgram(p graphics.Program) {
	gl.LinkProgram(uint32(p))
}

// PolygonOffset sets the scale and units used to calculate depth values
func (impl *GraphicsImpl) PolygonOffset(factor float32, units float32) {
	gl.PolygonOffset(factor, units)
}

// Ptr takes a slice or a pointer and returns an OpenGL compatbile address
func (impl *GraphicsImpl) Ptr(data interface{}) unsafe.Pointer {
	return gl.Ptr(data)
}

// PtrOffset takes a pointer offset and returns a GL-compatible pointer.
// Useful for functions such as glVertexAttribPointer that take pointer
// parameters indicating an offset rather than an absolute memory address.
func (impl *GraphicsImpl) PtrOffset(offset int) unsafe.Pointer {
	return gl.PtrOffset(offset)
}

// ReadBuffer specifies the color buffer source for pixels
func (impl *GraphicsImpl) ReadBuffer(src graphics.Enum) {
	gl.ReadBuffer(uint32(src))
}

// RenderbufferStorage establishes the format and dimensions of a renderbuffer
func (impl *GraphicsImpl) RenderbufferStorage(target graphics.Enum, internalformat graphics.Enum, width int32, height int32) {
	gl.RenderbufferStorage(uint32(target), uint32(internalformat), width, height)
}

// ShaderSource replaces the source code for a shader object.
func (impl *GraphicsImpl) ShaderSource(s graphics.Shader, source string) {
	glSource := gl.Str(source + "\x00")
	gl.ShaderSource(uint32(s), 1, &glSource, nil)
}

// TexImage2D writes a 2D texture image.
func (impl *GraphicsImpl) TexImage2D(target graphics.Enum, level, intfmt, width, height, border int32, format graphics.Enum, ty graphics.Enum, ptr unsafe.Pointer, dataLength int) {
	gl.TexImage2D(uint32(target), level, intfmt, width, height, border, uint32(format), uint32(ty), ptr)
}

// TexParameterf sets a float texture parameter
func (impl *GraphicsImpl) TexParameterf(target, pname graphics.Enum, param float32) {
	gl.TexParameterf(uint32(target), uint32(pname), param)
}

// TexParameterfv sets a float texture parameter
func (impl *GraphicsImpl) TexParameterfv(target, pname graphics.Enum, params *float32) {
	gl.TexParameterfv(uint32(target), uint32(pname), params)
}

// TexParameteri sets a float texture parameter
func (impl *GraphicsImpl) TexParameteri(target, pname graphics.Enum, param int32) {
	gl.TexParameteri(uint32(target), uint32(pname), param)
}

// TexStorage3D simultaneously specifies storage for all levels of a three-dimensional,
// two-dimensional array or cube-map array texture
func (impl *GraphicsImpl) TexStorage3D(target graphics.Enum, level int32, intfmt uint32, width, height, depth int32) {
	gl.TexStorage3D(uint32(target), level, intfmt, width, height, depth)
}

// TexSubImage3D specifies a three-dimensonal texture subimage
func (impl *GraphicsImpl) TexSubImage3D(target graphics.Enum, level, xoff, yoff, zoff, width, height, depth int32, fmt, ty graphics.Enum, ptr unsafe.Pointer) {
	gl.TexSubImage3D(uint32(target), level, xoff, yoff, zoff, width, height, depth, uint32(fmt), uint32(ty), ptr)
}

// Uniform1i specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1i(location int32, v int32) {
	gl.Uniform1i(location, v)
}

// Uniform1iv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1iv(location int32, values []int32) {
	gl.Uniform1iv(location, int32(len(values)), &values[0])
}

// Uniform1f specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1f(location int32, v float32) {
	gl.Uniform1f(location, v)
}

// Uniform1fv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1fv(location int32, values []float32) {
	gl.Uniform1fv(location, int32(len(values)), &values[0])
}

// Uniform3f specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform3f(location int32, v0, v1, v2 float32) {
	gl.Uniform3f(location, v0, v1, v2)
}

// Uniform3fv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform3fv(location int32, values []float32) {
	gl.Uniform3fv(location, int32(len(values)), &values[0])
}

// Uniform4f specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform4f(location int32, v0, v1, v2, v3 float32) {
	gl.Uniform4f(location, v0, v1, v2, v3)
}

// Uniform4fv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform4fv(location int32, values []float32) {
	gl.Uniform4fv(location, int32(len(values)), &values[0])
}

// UniformMatrix4fv specifies the value of a uniform variable for the current program object
// NOTE: value should be a mgl.Mat4 or []mgl.Mat4, else it will panic.
func (impl *GraphicsImpl) UniformMatrix4fv(location, count int32, transpose bool, value interface{}) {
	switch t := value.(type) {
	case mgl.Mat4:
		gl.UniformMatrix4fv(location, count, transpose, &(t[0]))
	case []mgl.Mat4:
		gl.UniformMatrix4fv(location, count, transpose, &(t[0][0]))
	default:
		panic(fmt.Sprintf("Unhandled case of type for %T in opengl.UniformMatrix4fv()\n", value))
	}
}

// UseProgram installs a program object as part of the current rendering state
func (impl *GraphicsImpl) UseProgram(p graphics.Program) {
	gl.UseProgram(uint32(p))
}

// VertexAttribPointer uses a bound buffer to define vertex attribute data.
//
// The size argument specifies the number of components per attribute,
// between 1-4. The stride argument specifies the byte offset between
// consecutive vertex attributes.
func (impl *GraphicsImpl) VertexAttribPointer(dst uint32, size int32, ty graphics.Enum, normalized bool, stride int32, ptr unsafe.Pointer) {
	gl.VertexAttribPointer(dst, size, uint32(ty), normalized, stride, ptr)
}

// VertexAttribPointer uses a bound buffer to define vertex attribute data.
// Only integer types are accepted by this function.
func (impl *GraphicsImpl) VertexAttribIPointer(dst uint32, size int32, ty graphics.Enum, stride int32, ptr unsafe.Pointer) {
	gl.VertexAttribIPointer(dst, size, uint32(ty), stride, ptr)
}

// Viewport sets the viewport, an affine transformation that
// normalizes device coordinates to window coordinates.
func (impl *GraphicsImpl) Viewport(x, y, width, height int32) {
	gl.Viewport(x, y, width, height)
}
