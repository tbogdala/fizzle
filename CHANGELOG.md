Changes since v0.1.0
====================

* APIBREAK: Many `fizzle/component` changes, including API breaks
* APIBREAK: Renderable.Core.Tex0 and Tex1 have been replaced with
  Renderable.Core.Tex which is a slice of texture OpenGL objects.
  The maximum number of textures is set with `MaxRenderableTextures`.

* NEW: `cmd/compeditor` for a component editor.
* NEW: basicSkinned shader for skeletal animation on GPU.
* NEW: fizzle.CreateLineV() to create a line using two Vec3 instead
  of six floats.

* BUG: GLSL VERTEX_BONE_IDS and VERTEX_BONE_WEIGHTS uniforms will now
  only be bound if the Renderable has a Skeleton.
* BUG: Changed base Renderable.Core.Shininess to 1.0 instead of 0.01 since
  values less than 1.0 produce artifacts with stander ADS lighting in the
  basic shader.
