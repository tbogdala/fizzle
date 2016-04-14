// Copyright 2016, Timothy` Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package opengles31

// NOTE: just started implementing some GLES3 features. This isn't complete yet.

/*
#cgo LDFLAGS: -lGLESv3  -lEGL
#include <stdlib.h>
#include <GLES3/gl3.h>
#include <GLES3/gl3ext.h>
#include <GLES3/gl3platform.h>
*/
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	mgl "github.com/go-gl/mathgl/mgl32"
	gles "github.com/remogatto/opengles2"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

// GraphicsImpl is the graphics provider for the mobile
// implementation of OpenGL.
type GraphicsImpl struct {
	// nothing at present
}

// InitOpenGLES2 initializes the OpenGL ES 2 graphics provider and
// sets it to be the current provider for the module.
func InitOpenGLES2() (*GraphicsImpl, error) {
	gp := new(GraphicsImpl)
	return gp, nil
}

// ActiveTexture selects the active texture unit
func (impl *GraphicsImpl) ActiveTexture(t graphics.Texture) {
	gles.ActiveTexture(gles.Enum(t))
}

// AttachShader attaches a shader object to a program object
func (impl *GraphicsImpl) AttachShader(p graphics.Program, s graphics.Shader) {
	gles.AttachShader(uint32(p), uint32(s))
}

// BindBuffer binds a buffer to the OpenGL target specified by enum
func (impl *GraphicsImpl) BindBuffer(target graphics.Enum, b graphics.Buffer) {
	gles.BindBuffer(gles.Enum(target), uint32(b))
}

// BindFragDataLocation binds a user-defined varying out variable
// to a fragment shader color number.
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) BindFragDataLocation(p graphics.Program, color uint32, name string) {
	// NO-OP
}

// BindFramebuffer binds a framebuffer to a framebuffer target
func (impl *GraphicsImpl) BindFramebuffer(target graphics.Enum, fb graphics.Buffer) {
	gles.BindFramebuffer(gles.Enum(target), uint32(fb))
}

// BindRenderbuffer binds a renderbuffer to a renderbuffer target
func (impl *GraphicsImpl) BindRenderbuffer(target graphics.Enum, renderbuffer graphics.Buffer) {
	gles.BindRenderbuffer(gles.Enum(target), uint32(renderbuffer))
}

// BindTexture binds a texture to the OpenGL target specified by enum
func (impl *GraphicsImpl) BindTexture(target graphics.Enum, t graphics.Texture) {
	gles.BindTexture(gles.Enum(target), uint32(t))
}

// BindVertexArray binds a vertex array object
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) BindVertexArray(a uint32) {
	// NO-OP
}

// BlendEquation specifies the equation used for both the RGB and
// alpha blend equations
func (impl *GraphicsImpl) BlendEquation(mode graphics.Enum) {
	gles.BlendEquation(gles.Enum(mode))
}

// BlendFunc specifies the pixel arithmetic for the blend fucntion
func (impl *GraphicsImpl) BlendFunc(sFactor, dFactor graphics.Enum) {
	gles.BlendFunc(gles.Enum(sFactor), gles.Enum(dFactor))
}

// BufferData creates a new data store for the bound buffer object.
func (impl *GraphicsImpl) BufferData(target graphics.Enum, size int, data unsafe.Pointer, usage graphics.Enum) {
	gles.BufferData(gles.Enum(target), gles.SizeiPtr(size), gles.Void(data), gles.Enum(usage))
}

// CheckFramebufferStatus checks the completeness status of a framebuffer
func (impl *GraphicsImpl) CheckFramebufferStatus(target graphics.Enum) graphics.Enum {
	return graphics.Enum(gles.CheckFramebufferStatus(gles.Enum(target)))
}

// Clear clears the window buffer specified in mask
func (impl *GraphicsImpl) Clear(mask graphics.Enum) {
	gles.Clear(gles.Bitfield(mask))
}

// ClearColor specifies the RGBA value used to clear the color buffers
func (impl *GraphicsImpl) ClearColor(red, green, blue, alpha float32) {
	gles.ClearColor(gles.Clampf(red), gles.Clampf(green), gles.Clampf(blue), gles.Clampf(alpha))
}

// CompileShader compiles the shader object
func (impl *GraphicsImpl) CompileShader(s graphics.Shader) {
	gles.CompileShader(uint32(s))
}

// CreateProgram creates a new shader program object
func (impl *GraphicsImpl) CreateProgram() graphics.Program {
	return graphics.Program(gles.CreateProgram())
}

// CreateShader creates a new shader object
func (impl *GraphicsImpl) CreateShader(ty graphics.Enum) graphics.Shader {
	return graphics.Shader(gles.CreateShader(gles.Enum(ty)))
}

// CullFace specifies whether to use front or back face culling
func (impl *GraphicsImpl) CullFace(mode graphics.Enum) {
	gles.CullFace(gles.Enum(mode))
}

// DeleteBuffer deletes the OpenGL buffer object
func (impl *GraphicsImpl) DeleteBuffer(b graphics.Buffer) {
	ui := uint32(b)
	gles.DeleteBuffers(1, &ui)
}

// DeleteFramebuffer deletes the framebuffer object
func (impl *GraphicsImpl) DeleteFramebuffer(fb graphics.Buffer) {
	ui := uint32(fb)
	gles.DeleteFramebuffers(1, &ui)
}

// DeleteProgram deletes the shader program object
func (impl *GraphicsImpl) DeleteProgram(p graphics.Program) {
	gles.DeleteProgram(uint32(p))
}

// DeleteRenderbuffer deletes the renderbuffer object
func (impl *GraphicsImpl) DeleteRenderbuffer(rb graphics.Buffer) {
	ui := uint32(rb)
	gles.DeleteRenderbuffers(1, &ui)
}

// DeleteShader deletes the shader object
func (impl *GraphicsImpl) DeleteShader(s graphics.Shader) {
	gles.DeleteShader(uint32(s))
}

// DeleteTexture deletes the specified texture
func (impl *GraphicsImpl) DeleteTexture(v graphics.Texture) {
	ui := uint32(v)
	gles.DeleteTextures(1, &ui)
}

// DeleteVertexArray deletes an OpenGL VAO
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) DeleteVertexArray(a uint32) {
	// NO-OP
}

// DepthMask enables or disables writing into the depth buffer
func (impl *GraphicsImpl) DepthMask(flag bool) {
	gles.DepthMask(flag)
}

// Disable disables various GL capabilities.
func (impl *GraphicsImpl) Disable(e graphics.Enum) {
	gles.Disable(gles.Enum(e))
}

// DrawBuffers specifies a list of color buffers to be drawn into
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) DrawBuffers(buffers []uint32) {
	// NO-OP
}

// DrawElements renders primitives from array data
func (impl *GraphicsImpl) DrawElements(mode graphics.Enum, count int32, ty graphics.Enum, indices unsafe.Pointer) {
	gles.DrawElements(gles.Enum(mode), gles.Sizei(count), gles.Enum(ty), gles.Void(indices))
}

// DrawArrays renders primitives from array data
func (impl *GraphicsImpl) DrawArrays(mode graphics.Enum, first int32, count int32) {
	gles.DrawArrays(gles.Enum(mode), first, gles.Sizei(count))
}

// Enable enables various GL capabilities
func (impl *GraphicsImpl) Enable(e graphics.Enum) {
	gles.Enable(gles.Enum(e))
}

// EnableVertexAttribArray enables a vertex attribute array
func (impl *GraphicsImpl) EnableVertexAttribArray(a uint32) {
	gles.EnableVertexAttribArray(a)
}

// Finish blocks until the effects of all previously called GL commands are complete
func (impl *GraphicsImpl) Finish() {
	gles.Finish()
}

// Flush forces the execution of GL commands that have been buffered
func (impl *GraphicsImpl) Flush() {
	gles.Flush()
}

// FramebufferRenderbuffer attaches a renderbuffer as a logical buffer
// of a framebuffer object
func (impl *GraphicsImpl) FramebufferRenderbuffer(target, attachment, renderbuffertarget graphics.Enum, renderbuffer graphics.Buffer) {
	gles.FramebufferRenderbuffer(gles.Enum(target), gles.Enum(attachment), gles.Enum(renderbuffertarget), uint32(renderbuffer))
}

// FramebufferTexture2D attaches a texture object to a framebuffer
func (impl *GraphicsImpl) FramebufferTexture2D(target, attachment, textarget graphics.Enum, texture graphics.Texture, level int32) {
	gles.FramebufferTexture2D(gles.Enum(target), gles.Enum(attachment), gles.Enum(textarget), uint32(texture), level)
}

// GenBuffer creates an OpenGL buffer object
func (impl *GraphicsImpl) GenBuffer() graphics.Buffer {
	var b uint32
	gles.GenBuffers(1, &b)
	return graphics.Buffer(b)
}

// GenerateMipmap generates mipmaps for a specified texture target
func (impl *GraphicsImpl) GenerateMipmap(t graphics.Enum) {
	gles.GenerateMipmap(gles.Enum(t))
}

// GenFramebuffer generates a OpenGL framebuffer object
func (impl *GraphicsImpl) GenFramebuffer() graphics.Buffer {
	var b uint32
	gles.GenFramebuffers(1, &b)
	return graphics.Buffer(b)
}

// GenRenderbuffer generates a OpenGL renderbuffer object
func (impl *GraphicsImpl) GenRenderbuffer() graphics.Buffer {
	var b uint32
	gles.GenRenderbuffers(1, &b)
	return graphics.Buffer(b)
}

// GenTexture creates an OpenGL texture object
func (impl *GraphicsImpl) GenTexture() graphics.Texture {
	var t uint32
	gles.GenTextures(1, &t)
	return graphics.Texture(t)
}

// GenVertexArray creates an OpoenGL VAO
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) GenVertexArray() uint32 {
	// NO-OP
	return 0
}

// GetAttribLocation returns the location of a attribute variable
func (impl *GraphicsImpl) GetAttribLocation(p graphics.Program, name string) int32 {
	return int32(gles.GetAttribLocation(uint32(p), name))
}

// GetError returns the next error
func (impl *GraphicsImpl) GetError() uint32 {
	return uint32(gles.GetError())
}

// GetProgramInfoLog returns the information log for a program object
func (impl *GraphicsImpl) GetProgramInfoLog(p graphics.Program) string {
	var logLength int32
	gles.GetProgramiv(uint32(p), gles.Enum(graphics.INFO_LOG_LENGTH), &logLength)
	return gles.GetProgramInfoLog(uint32(p), gles.Sizei(logLength+1), nil)
}

// GetProgramiv returns a parameter from the program object
func (impl *GraphicsImpl) GetProgramiv(p graphics.Program, pname graphics.Enum, params *int32) {
	gles.GetProgramiv(uint32(p), gles.Enum(pname), params)
}

// GetShaderInfoLog returns the information log for a shader object
func (impl *GraphicsImpl) GetShaderInfoLog(s graphics.Shader) string {
	var logLength int32
	gles.GetShaderiv(uint32(s), gles.Enum(graphics.INFO_LOG_LENGTH), &logLength)
	return gles.GetShaderInfoLog(uint32(s), gles.Sizei(logLength+1), nil)
}

// GetShaderiv returns a parameter from the shader object
func (impl *GraphicsImpl) GetShaderiv(s graphics.Shader, pname graphics.Enum, params *int32) {
	gles.GetShaderiv(uint32(s), gles.Enum(pname), params)
}

// GetUniformLocation returns the location of a uniform variable
func (impl *GraphicsImpl) GetUniformLocation(p graphics.Program, name string) int32 {
	return int32(gles.GetUniformLocation(uint32(p), name))
}

// LinkProgram links a program object
func (impl *GraphicsImpl) LinkProgram(p graphics.Program) {
	gles.LinkProgram(uint32(p))
}

// PolygonMode sets a polygon rasterization mode.
func (impl *GraphicsImpl) PolygonMode(face, mode graphics.Enum) {
	// NO-OP: no support in OpenGL ES
}

// PolygonOffset sets the scale and units used to calculate depth values
func (impl *GraphicsImpl) PolygonOffset(factor float32, units float32) {
	gles.PolygonOffset(factor, units)
}

// Ptr takes a slice or pointer (to a singular scalar value or the first
// element of an array or slice) and returns its GL-compatible address.
// NOTE: Shamelessly ripped from: github.com/go-gl/gl/blob/master/v3.3-core/gl/conversions.go
// Thanks, guys!
func (impl *GraphicsImpl) Ptr(data interface{}) unsafe.Pointer {
	if data == nil {
		return unsafe.Pointer(nil)
	}
	var addr unsafe.Pointer
	v := reflect.ValueOf(data)
	switch v.Type().Kind() {
	case reflect.Ptr:
		e := v.Elem()
		switch e.Kind() {
		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			addr = unsafe.Pointer(e.UnsafeAddr())
		}
	case reflect.Uintptr:
		addr = unsafe.Pointer(v.Pointer())
	case reflect.Slice:
		addr = unsafe.Pointer(v.Index(0).UnsafeAddr())
	default:
		panic(fmt.Sprintf("Unsupported type %s; must be a pointer, slice, or array", v.Type()))
	}
	return addr
}

// PtrOffset takes a pointer offset and returns a GL-compatible pointer.
// Useful for functions such as glVertexAttribPointer that take pointer
// parameters indicating an offset rather than an absolute memory address.
func (impl *GraphicsImpl) PtrOffset(offset int) unsafe.Pointer {
	// This may not be quite right ... ... ...
	return unsafe.Pointer(uintptr(offset))
}

// ReadBuffer specifies the color buffer source for pixels
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) ReadBuffer(src graphics.Enum) {
	// NO-OP
}

// RenderbufferStorage establishes the format and dimensions of a renderbuffer
func (impl *GraphicsImpl) RenderbufferStorage(target graphics.Enum, internalformat graphics.Enum, width int32, height int32) {
	gles.RenderbufferStorage(gles.Enum(target), gles.Enum(internalformat), gles.Sizei(width), gles.Sizei(height))
}

// ShaderSource replaces the source code for a shader object.
func (impl *GraphicsImpl) ShaderSource(s graphics.Shader, source string) {
	gles.ShaderSource(uint32(s), 1, &source, nil)
}

// TexImage2D writes a 2D texture image.
func (impl *GraphicsImpl) TexImage2D(target graphics.Enum, level, intfmt, width, height, border int32, format graphics.Enum, ty graphics.Enum, ptr unsafe.Pointer, dataLength int) {
	gles.TexImage2D(gles.Enum(target), level, intfmt, gles.Sizei(width), gles.Sizei(height), border, gles.Enum(format), gles.Enum(ty), gles.Void(ptr))
}

// TexParameterf sets a float texture parameter
func (impl *GraphicsImpl) TexParameterf(target, pname graphics.Enum, param float32) {
	gles.TexParameterf(gles.Enum(target), gles.Enum(pname), param)
}

// TexParameterfv sets a float texture parameter
func (impl *GraphicsImpl) TexParameterfv(target, pname graphics.Enum, params *float32) {
	gles.TexParameterfv(gles.Enum(target), gles.Enum(pname), params)
}

// TexParameteri sets a float texture parameter
func (impl *GraphicsImpl) TexParameteri(target, pname graphics.Enum, param int32) {
	gles.TexParameteri(gles.Enum(target), gles.Enum(pname), param)
}

// TexStorage3D simultaneously specifies storage for all levels of a three-dimensional,
// two-dimensional array or cube-map array texture
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) TexStorage3D(target graphics.Enum, level int32, intfmt uint32, width, height, depth int32) {
	/* void TexStorage3D(enum target, sizei levels, enum internalformat, sizei width, sizei height, sizei depth); */
	C.glTexStorage3D(C.GLenum(target), C.GLsizei(level), C.GLenum(intfmt), C.GLsizei(width), C.GLsizei(height), C.GLsizei(depth))
}

// TexSubImage3D specifies a three-dimensonal texture subimage
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) TexSubImage3D(target graphics.Enum, level, xoff, yoff, zoff, width, height, depth int32, fmt, ty graphics.Enum, ptr unsafe.Pointer) {
	C.glTexSubImage3D(C.GLenum(target), C.GLint(level), C.GLint(xoff), C.GLint(yoff), C.GLint(zoff), C.GLsizei(width),
		C.GLsizei(height), C.GLsizei(depth), C.GLenum(fmt), C.GLenum(ty), unsafe.Pointer(ptr))
}

// Uniform1i specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1i(location int32, v int32) {
	gles.Uniform1i(location, v)
}

// Uniform1iv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1iv(location int32, values []int32) {
	gles.Uniform1iv(location, gles.Sizei(len(values)), &values[0])
}

// Uniform1f specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1f(location int32, v float32) {
	gles.Uniform1f(location, v)
}

// Uniform1fv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1fv(location int32, values []float32) {
	gles.Uniform1fv(location, gles.Sizei(len(values)), &values[0])
}

// Uniform3f specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform3f(location int32, v0, v1, v2 float32) {
	gles.Uniform3f(location, v0, v1, v2)
}

// Uniform3fv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform3fv(location int32, values []float32) {
	gles.Uniform3fv(location, gles.Sizei(len(values)), &values[0])
}

// Uniform4f specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform4f(location int32, v0, v1, v2, v3 float32) {
	gles.Uniform4f(location, v0, v1, v2, v3)
}

// Uniform4fv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform4fv(location int32, values []float32) {
	gles.Uniform4fv(location, gles.Sizei(len(values)), &values[0])
}

// UniformMatrix4fv specifies the value of a uniform variable for the current program object.
// NOTE: value should be a mgl.Mat4 or []mgl.Mat4, else it will panic.
func (impl *GraphicsImpl) UniformMatrix4fv(location, count int32, transpose bool, value interface{}) {
	switch t := value.(type) {
	case mgl.Mat4:
		gles.UniformMatrix4fv(location, gles.Sizei(count), transpose, &t[0])
	case []mgl.Mat4:
		gles.UniformMatrix4fv(location, gles.Sizei(count), transpose, &t[0][0])
	default:
		panic(fmt.Sprintf("Unhandled case of type for %T in opengles2.UniformMatrix4fv()\n", value))
	}
}

// UseProgram installs a program object as part of the current rendering state
func (impl *GraphicsImpl) UseProgram(p graphics.Program) {
	gles.UseProgram(uint32(p))
}

// VertexAttribPointer uses a bound buffer to define vertex attribute data.
//
// The size argument specifies the number of components per attribute,
// between 1-4. The stride argument specifies the byte offset between
// consecutive vertex attributes.
func (impl *GraphicsImpl) VertexAttribPointer(dst uint32, size int32, ty graphics.Enum, normalized bool, stride int32, ptr unsafe.Pointer) {
	gles.VertexAttribPointer(dst, size, gles.Enum(ty), normalized, gles.Sizei(stride), ptr)
}

// VertexAttribPointer uses a bound buffer to define vertex attribute data.
// Only integer types are accepted by this function.
// Note: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) VertexAttribIPointer(dst uint32, size int32, ty graphics.Enum, stride int32, ptr unsafe.Pointer) {
	// NO-OP
}

// Viewport sets the viewport, an affine transformation that
// normalizes device coordinates to window coordinates.
func (impl *GraphicsImpl) Viewport(x, y, width, height int32) {
	gles.Viewport(x, y, gles.Sizei(width), gles.Sizei(height))
}
