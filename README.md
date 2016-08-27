FIZZLE v0.1.0
=============

Fizzle is an OpenGL rendering engine written in the [Go][golang] programming language
that currently has a forward rendering pipeline and a prototype deferred renderer
as well.

In some regards, it is the spiritual successor to my first 3d engine, [PortableGLUE][pg].


UNDER CONSTRUCTION
==================

The engine is currently in an alpha state, but you are welcome to see how
it's progressing.  Any API break should increment the minor version number and
any patch release tags should remain compatible even in development 0.x versions.


Requirements
------------

* [GLFW][glfw-go] (v3.1) - native library and go binding for window creation
* [Mathgl][mgl] - for 3d math
* [Freetype][ftgo] - for dynamic font texture generation
* [Groggy][groggy] - for flexible logging
* [Gombz][gombz] - provides a serializable data structure for 3d models and animations
* [EweyGewey][ewey] - (optional) some examples use OpenGL GUI library in the lines of imgui

Additionally, a backend graphics provider needs to be used. At present, fizzle
supports the following:

* [Go GL][go-gl] - pre-generated OpenGL bindings using their glow project
* [opengles2][go-gles] - Go bindings to the OpenGL ES 2.0 library

These are included when the `graphicsprovider` subpackage is used and direct
importing is not required.

Installation
------------

The dependency Go libraries can be installed with the following commands.

```bash
go get github.com/go-gl/glfw/v3.1/glfw
go get github.com/go-gl/mathgl/mgl32
go get github.com/golang/freetype
go get github.com/tbogdala/groggy
go get github.com/tbogdala/gombz
```

An OpenGL library will also be required for desktop applications; install
the OpenGL 3.3 library with the following command:

```bash
go get github.com/go-gl/gl/v3.3-core/gl
```

If you're compiling for Android/iOS, then you will need an OpenGL ES library,
and that can be installed with the following command instead:

```bash
go get github.com/remogatto/opengles2
```

This does assume that you have the native GLFW 3.1 library installed already
accessible to Go tools.

Current Features
----------------

* forward rendering engine with limited dynamic lighting
* limited dynamic shadow support
* able to define components using JSON files
* skeletal animations
* basic camera support
* basic particle editor (cmd/particles)
* basic shader explorer (examples/shaders)


TODO
----

The following need to be addressed in order to start releases:

* documentation
* api comments
* samples
* code cleanups
* possibly remove use of [Groggy][groggy]


LICENSE
=======

Fizzle is released under the BSD license. See the [LICENSE][license-link] file for more details.


[golang]: https://golang.org/
[groggy]: https://github.com/tbogdala/groggy
[gombz]: https://github.com/tbogdala/gombz
[pg]: https://bitbucket.org/tbogdala/portableglue
[glfw-go]: https://github.com/go-gl/glfw
[go-gl]: https://github.com/go-gl/glow
[opengles2]: https://github.com/remogatto/opengles2
[mgl]: https://github.com/go-gl/mathgl
[ftgo]: https://github.com/golang/freetype
[ewey]: https://github.com/tbogdala/eweygewey
[license-link]: https://raw.githubusercontent.com/tbogdala/fizzle/master/LICENSE
