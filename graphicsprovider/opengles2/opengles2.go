// Copyright 2016, Timothy` Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package opengles2

import "C"

import (
	"fmt"
	"unsafe"

	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	gl "golang.org/x/mobile/gl"
)

// GraphicsImpl is the graphics provider for the mobile
// implementation of OpenGL.
type GraphicsImpl struct {
	// the context to use for the rendering
	Ctx gl.Context
}

// InitOpenGLES2 initializes the OpenGL ES 2 graphics provider and
// sets it to be the current provider for the module.
func InitOpenGLES2() (*GraphicsImpl, error) {
	gp := new(GraphicsImpl)
	gp.Ctx, _ = gl.NewContext()
	return gp, nil
}

// ActiveTexture selects the active texture unit
func (impl *GraphicsImpl) ActiveTexture(t graphics.Texture) {
	impl.Ctx.ActiveTexture(gl.Enum(t))
}

// AttachShader attaches a shader object to a program object
func (impl *GraphicsImpl) AttachShader(p graphics.Program, s graphics.Shader) {
	impl.Ctx.AttachShader(gl.Program{Init: true, Value: uint32(p)}, gl.Shader{Value: uint32(s)})
}

// BindBuffer binds a buffer to the OpenGL target specified by enum
func (impl *GraphicsImpl) BindBuffer(target graphics.Enum, b graphics.Buffer) {
	impl.Ctx.BindBuffer(gl.Enum(target), gl.Buffer{Value: uint32(b)})
}

// BindFragDataLocation binds a user-defined varying out variable
// to a fragment shader color number.
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) BindFragDataLocation(p graphics.Program, color uint32, name string) {
	// NO-OP
}

// BindFramebuffer binds a framebuffer to a framebuffer target
func (impl *GraphicsImpl) BindFramebuffer(target graphics.Enum, fb graphics.Buffer) {
	impl.Ctx.BindFramebuffer(gl.Enum(target), gl.Framebuffer{Value: uint32(fb)})
}

// BindRenderbuffer binds a renderbuffer to a renderbuffer target
func (impl *GraphicsImpl) BindRenderbuffer(target graphics.Enum, renderbuffer graphics.Buffer) {
	impl.Ctx.BindRenderbuffer(gl.Enum(target), gl.Renderbuffer{Value: uint32(renderbuffer)})
}

// BindTexture binds a texture to the OpenGL target specified by enum
func (impl *GraphicsImpl) BindTexture(target graphics.Enum, t graphics.Texture) {
	impl.Ctx.BindTexture(gl.Enum(target), gl.Texture{Value: uint32(t)})
}

// BindVertexArray binds a vertex array object
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) BindVertexArray(a uint32) {
	// NO-OP
}

// BlendEquation specifies the equation used for both the RGB and
// alpha blend equations
func (impl *GraphicsImpl) BlendEquation(mode graphics.Enum) {
	impl.Ctx.BlendEquation(gl.Enum(mode))
}

// BlendFunc specifies the pixel arithmetic for the blend fucntion
func (impl *GraphicsImpl) BlendFunc(sFactor, dFactor graphics.Enum) {
	impl.Ctx.BlendFunc(gl.Enum(sFactor), gl.Enum(dFactor))
}

// BufferData creates a new data store for the bound buffer object.
func (impl *GraphicsImpl) BufferData(target graphics.Enum, size int, data unsafe.Pointer, usage graphics.Enum) {
	byteSlice := C.GoBytes(data, size)
	impl.Ctx.BufferData(gl.Enum(target), byteSlice, gl.Enum(usage))
}

// CheckFramebufferStatus checks the completeness status of a framebuffer
func (impl *GraphicsImpl) CheckFramebufferStatus(target graphics.Enum) graphics.Enum {
	return graphics.Enum(impl.Ctx.CheckFramebufferStatus(gl.Enum(target)))
}

// Clear clears the window buffer specified in mask
func (impl *GraphicsImpl) Clear(mask graphics.Enum) {
	impl.Ctx.Clear(gl.Enum(mask))
}

// ClearColor specifies the RGBA value used to clear the color buffers
func (impl *GraphicsImpl) ClearColor(red, green, blue, alpha float32) {
	impl.Ctx.ClearColor(red, green, blue, alpha)
}

// CompileShader compiles the shader object
func (impl *GraphicsImpl) CompileShader(s graphics.Shader) {
	impl.Ctx.CompileShader(gl.Shader{Value: uint32(s)})
}

// CreateProgram creates a new shader program object
func (impl *GraphicsImpl) CreateProgram() graphics.Program {
	return graphics.Program(impl.Ctx.CreateProgram().Value)
}

// CreateShader creates a new shader object
func (impl *GraphicsImpl) CreateShader(ty graphics.Enum) graphics.Shader {
	return graphics.Shader(impl.Ctx.CreateShader(gl.Enum(ty)).Value)
}

// CullFace specifies whether to use front or back face culling
func (impl *GraphicsImpl) CullFace(mode graphics.Enum) {
	impl.Ctx.CullFace(gl.Enum(mode))
}

// DeleteBuffer deletes the OpenGL buffer object
func (impl *GraphicsImpl) DeleteBuffer(b graphics.Buffer) {
	impl.Ctx.DeleteBuffer(gl.Buffer{Value: uint32(b)})
}

// DeleteFramebuffer deletes the framebuffer object
func (impl *GraphicsImpl) DeleteFramebuffer(fb graphics.Buffer) {
	impl.Ctx.DeleteFramebuffer(gl.Framebuffer{Value: uint32(fb)})
}

// DeleteProgram deletes the shader program object
func (impl *GraphicsImpl) DeleteProgram(p graphics.Program) {
	impl.Ctx.DeleteProgram(gl.Program{Value: uint32(p)})
}

// DeleteRenderbuffer deletes the renderbuffer object
func (impl *GraphicsImpl) DeleteRenderbuffer(rb graphics.Buffer) {
	impl.Ctx.DeleteRenderbuffer(gl.Renderbuffer{Value: uint32(rb)})
}

// DeleteShader deletes the shader object
func (impl *GraphicsImpl) DeleteShader(s graphics.Shader) {
	impl.Ctx.DeleteShader(gl.Shader{Value: uint32(s)})
}

// DeleteTexture deletes the specified texture
func (impl *GraphicsImpl) DeleteTexture(v graphics.Texture) {
	impl.Ctx.DeleteTexture(gl.Texture{Value: uint32(v)})
}

// DeleteVertexArray deletes an OpenGL VAO
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) DeleteVertexArray(a uint32) {
	// NO-OP
}

// DepthMask enables or disables writing into the depth buffer
func (impl *GraphicsImpl) DepthMask(flag bool) {
	impl.Ctx.DepthMask(flag)
}

// Disable disables various GL capabilities.
func (impl *GraphicsImpl) Disable(e graphics.Enum) {
	impl.Ctx.Disable(gl.Enum(e))
}

// DrawBuffers specifies a list of color buffers to be drawn into
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) DrawBuffers(buffers []uint32) {
	// NO-OP
}

// DrawElements renders primitives from array data
func (impl *GraphicsImpl) DrawElements(mode graphics.Enum, count int32, ty graphics.Enum, indices unsafe.Pointer) {
	impl.Ctx.DrawElements(gl.Enum(mode), int(count), gl.Enum(ty), int(uintptr(indices)))
}

// Enable enables various GL capabilities
func (impl *GraphicsImpl) Enable(e graphics.Enum) {
	impl.Ctx.Enable(gl.Enum(e))
}

// EnableVertexAttribArray enables a vertex attribute array
func (impl *GraphicsImpl) EnableVertexAttribArray(a uint32) {
	impl.Ctx.EnableVertexAttribArray(gl.Attrib{Value: uint(a)})
}

// Finish blocks until the effects of all previously called GL commands are complete
func (impl *GraphicsImpl) Finish() {
	impl.Ctx.Finish()
}

// FramebufferRenderbuffer attaches a renderbuffer as a logical buffer
// of a framebuffer object
func (impl *GraphicsImpl) FramebufferRenderbuffer(target, attachment, renderbuffertarget graphics.Enum, renderbuffer graphics.Buffer) {
	impl.Ctx.FramebufferRenderbuffer(gl.Enum(target),
		gl.Enum(attachment),
		gl.Enum(renderbuffertarget),
		gl.Renderbuffer{Value: uint32(renderbuffer)})
}

// FramebufferTexture2D attaches a texture object to a framebuffer
func (impl *GraphicsImpl) FramebufferTexture2D(target, attachment, textarget graphics.Enum, texture graphics.Texture, level int32) {
	impl.Ctx.FramebufferTexture2D(gl.Enum(target),
		gl.Enum(attachment),
		gl.Enum(textarget),
		gl.Texture{Value: uint32(texture)},
		int(level))
}

// GenBuffer creates an OpenGL buffer object
func (impl *GraphicsImpl) GenBuffer() graphics.Buffer {
	return graphics.Buffer(impl.Ctx.CreateBuffer().Value)
}

// GenerateMipmap generates mipmaps for a specified texture target
func (impl *GraphicsImpl) GenerateMipmap(t graphics.Enum) {
	impl.Ctx.GenerateMipmap(gl.Enum(t))
}

// GenFramebuffer generates a OpenGL framebuffer object
func (impl *GraphicsImpl) GenFramebuffer() graphics.Buffer {
	return graphics.Buffer(impl.Ctx.CreateFramebuffer().Value)
}

// GenRenderbuffer generates a OpenGL renderbuffer object
func (impl *GraphicsImpl) GenRenderbuffer() graphics.Buffer {
	return graphics.Buffer(impl.Ctx.CreateRenderbuffer().Value)
}

// GenTexture creates an OpenGL texture object
func (impl *GraphicsImpl) GenTexture() graphics.Texture {
	return graphics.Texture(impl.Ctx.CreateTexture().Value)
}

// GenVertexArray creates an OpoenGL VAO
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) GenVertexArray() uint32 {
	// NO-OP
	return 0
}

// GetAttribLocation returns the location of a attribute variable
func (impl *GraphicsImpl) GetAttribLocation(p graphics.Program, name string) int32 {
	return int32(impl.Ctx.GetAttribLocation(gl.Program{Value: uint32(p)}, name).Value)
}

// GetError returns the next error
func (impl *GraphicsImpl) GetError() uint32 {
	return uint32(impl.Ctx.GetError())
}

// GetProgramInfoLog returns the information log for a program object
func (impl *GraphicsImpl) GetProgramInfoLog(p graphics.Program) string {
	return impl.Ctx.GetProgramInfoLog(gl.Program{Value: uint32(p)})
}

// GetProgramiv returns a parameter from the program object
func (impl *GraphicsImpl) GetProgramiv(p graphics.Program, pname graphics.Enum, params *int32) {
	*params = int32(impl.Ctx.GetProgrami(gl.Program{Value: uint32(p)}, gl.Enum(pname)))
}

// GetShaderInfoLog returns the information log for a shader object
func (impl *GraphicsImpl) GetShaderInfoLog(s graphics.Shader) string {
	return impl.Ctx.GetShaderInfoLog(gl.Shader{Value: uint32(s)})
}

// GetShaderiv returns a parameter from the shader object
func (impl *GraphicsImpl) GetShaderiv(s graphics.Shader, pname graphics.Enum, params *int32) {
	*params = int32(impl.Ctx.GetShaderi(gl.Shader{Value: uint32(s)}, gl.Enum(pname)))
}

// GetUniformLocation returns the location of a uniform variable
func (impl *GraphicsImpl) GetUniformLocation(p graphics.Program, name string) int32 {
	return int32(impl.Ctx.GetUniformLocation(gl.Program{Value: uint32(p)}, name).Value)
}

// LinkProgram links a program object
func (impl *GraphicsImpl) LinkProgram(p graphics.Program) {
	impl.Ctx.LinkProgram(gl.Program{Value: uint32(p)})
}

// PolygonOffset sets the scale and units used to calculate depth values
func (impl *GraphicsImpl) PolygonOffset(factor float32, units float32) {
	impl.Ctx.PolygonOffset(factor, units)
}

// Ptr takes a slice or a pointer and returns an OpenGL compatbile address
func (impl *GraphicsImpl) Ptr(data interface{}) unsafe.Pointer {
	// This may not be quite right ... ... ...
	return unsafe.Pointer(&data)
}

// PtrOffset takes a pointer offset and returns a GL-compatible pointer.
// Useful for functions such as glVertexAttribPointer that take pointer
// parameters indicating an offset rather than an absolute memory address.
func (impl *GraphicsImpl) PtrOffset(offset int) unsafe.Pointer {
	// This may not be quite right ... ... ...
	var ptr = (uintptr)(offset)
	return unsafe.Pointer(ptr)
}

// ReadBuffer specifies the color buffer source for pixels
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) ReadBuffer(src graphics.Enum) {
	// NO-OP
}

// RenderbufferStorage establishes the format and dimensions of a renderbuffer
func (impl *GraphicsImpl) RenderbufferStorage(target graphics.Enum, internalformat graphics.Enum, width int32, height int32) {
	impl.Ctx.RenderbufferStorage(gl.Enum(target), gl.Enum(internalformat), int(width), int(height))
}

// ShaderSource replaces the source code for a shader object.
func (impl *GraphicsImpl) ShaderSource(s graphics.Shader, source string) {
	impl.Ctx.ShaderSource(gl.Shader{Value: uint32(s)}, source)
}

// TexImage2D writes a 2D texture image.
func (impl *GraphicsImpl) TexImage2D(target graphics.Enum, level, intfmt, width, height, border int32, format graphics.Enum, ty graphics.Enum, ptr unsafe.Pointer, dataLength int) {
	impl.Ctx.TexImage2D(gl.Enum(target),
		int(level),
		int(width),
		int(height),
		gl.Enum(format),
		gl.Enum(ty),
		C.GoBytes(ptr, dataLength))
}

// TexParameterf sets a float texture parameter
func (impl *GraphicsImpl) TexParameterf(target, pname graphics.Enum, param float32) {
	impl.Ctx.TexParameterf(gl.Enum(target), gl.Enum(pname), param)
}

// TexParameterfv sets a float texture parameter
func (impl *GraphicsImpl) TexParameterfv(target, pname graphics.Enum, params *float32) {
	impl.Ctx.TexParameterfv(gl.Enum(target), gl.Enum(pname), []float32{*params})
}

// TexParameteri sets a float texture parameter
func (impl *GraphicsImpl) TexParameteri(target, pname graphics.Enum, param int32) {
	impl.Ctx.TexParameteri(gl.Enum(target), gl.Enum(pname), int(param))
}

// TexStorage3D simultaneously specifies storage for all levels of a three-dimensional,
// two-dimensional array or cube-map array texture
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) TexStorage3D(target graphics.Enum, level int32, intfmt uint32, width, height, depth int32) {
	// NO-OP
}

// TexSubImage3D specifies a three-dimensonal texture subimage
// NOTE: not implemented in OpenGL ES 2
func (impl *GraphicsImpl) TexSubImage3D(target graphics.Enum, level, xoff, yoff, zoff, width, height, depth int32, fmt, ty graphics.Enum, ptr unsafe.Pointer) {
	// NO-OP
}

// Uniform1i specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1i(location int32, v int32) {
	impl.Ctx.Uniform1i(gl.Uniform{Value: location}, int(v))
}

// Uniform1iv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1iv(location int32, values []int32) {
	impl.Ctx.Uniform1iv(gl.Uniform{Value: location}, values)
}

// Uniform1f specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1f(location int32, v float32) {
	impl.Ctx.Uniform1f(gl.Uniform{Value: location}, v)
}

// Uniform1fv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform1fv(location int32, values []float32) {
	impl.Ctx.Uniform1fv(gl.Uniform{Value: location}, values)
}

// Uniform3f specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform3f(location int32, v0, v1, v2 float32) {
	impl.Ctx.Uniform3f(gl.Uniform{Value: location}, v0, v1, v2)
}

// Uniform3fv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform3fv(location int32, values []float32) {
	impl.Ctx.Uniform3fv(gl.Uniform{Value: location}, values)
}

// Uniform4f specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform4f(location int32, v0, v1, v2, v3 float32) {
	impl.Ctx.Uniform4f(gl.Uniform{Value: location}, v0, v1, v2, v3)
}

// Uniform4fv specifies the value of a uniform variable for the current program object
func (impl *GraphicsImpl) Uniform4fv(location int32, values []float32) {
	impl.Ctx.Uniform4fv(gl.Uniform{Value: location}, values)
}

// UniformMatrix4fv specifies the value of a uniform variable for the current program object.
// NOTE: value should be a mgl.Mat4 or []mgl.Mat4, else it will panic.
func (impl *GraphicsImpl) UniformMatrix4fv(location, count int32, transpose bool, value interface{}) {
	switch t := value.(type) {
	case mgl.Mat4:
		fa := [16]float32(t)
		impl.Ctx.UniformMatrix4fv(gl.Uniform{Value: location}, fa[:])
	case []mgl.Mat4:
		// sadly, we're going to have to build a master slice from the tiny slices
		master := make([]float32, 0, 16*count)
		for _, mat4 := range t {
			fa := [16]float32(mat4)
			for i := 0; i < 16; i++ {
				master = append(master, fa[i])
			}
		}
		impl.Ctx.UniformMatrix4fv(gl.Uniform{Value: location}, master)
	default:
		panic(fmt.Sprintf("Unhandled case of type for %T in opengles2.UniformMatrix4fv()\n", value))
	}
}

// UseProgram installs a program object as part of the current rendering state
func (impl *GraphicsImpl) UseProgram(p graphics.Program) {
	impl.Ctx.UseProgram(gl.Program{Value: uint32(p)})
}

// VertexAttribPointer uses a bound buffer to define vertex attribute data.
//
// The size argument specifies the number of components per attribute,
// between 1-4. The stride argument specifies the byte offset between
// consecutive vertex attributes.
func (impl *GraphicsImpl) VertexAttribPointer(dst uint32, size int32, ty graphics.Enum, normalized bool, stride int32, ptr unsafe.Pointer) {
	impl.Ctx.VertexAttribPointer(gl.Attrib{Value: uint(dst)}, int(size), gl.Enum(ty), normalized, int(stride), int(uintptr(ptr)))
}

// Viewport sets the viewport, an affine transformation that
// normalizes device coordinates to window coordinates.
func (impl *GraphicsImpl) Viewport(x, y, width, height int32) {
	impl.Ctx.Viewport(int(x), int(y), int(width), int(height))
}
