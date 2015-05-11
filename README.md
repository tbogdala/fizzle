FIZZLE
======

Fizzle is an OpenGL rendering engine written in the [Go][golang] programming language
that currently has a deferred rendering pipeline.

In some regards, it is the spiritual successor to my first 3d engine, [PortableGLUE][pg].


UNDER CONSTRUCTION
==================

The engine is currently in an alpha state, but you are welcome to see how
it's progressing.

Requirements
------------

* [GLFW][glfw-go] (v3.1) - native library and go binding for window creation
* [Go GL][go-gl] - pre-generated OpenGL bindings using their glow project
* [Mathgl][mgl] - for 3d math
* [Freetype-go][ft-go] - for dynamic font texture generation
* [Groggy][groggy] - for flexible logging

Current Features
----------------

* deferred rendering engine
* able to define components using JSON files
* support for freetype compatible fonts

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

Fizzle is released under the BSD license. See the `LICENSE` file for more details.


[golang]: https://golang.org/
[groggy]: https://github.com/tbogdala/groggy
[pg]: https://bitbucket.org/tbogdala/portableglue
[glfw-go]: https://github.com/go-gl/glfw
[go-gl]: https://github.com/go-gl/glow
[mgl]: https://github.com/go-gl/mathgl
[ftgo]: code.google.com/p/freetype-go/freetype