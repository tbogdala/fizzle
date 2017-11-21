NEW
===

* APIBREAK: removed DrawLines from Renderer interface and replaced it with
  a more generic DrawRenderableWithMode().

* NEW: Supporting the TRIANGLE_STRIP mode for glDrawElements in DrawRenderable*().

* NEW: Added a skybox shader; use CreateSkyboxShader() to instance it.

* NEW: Added a landscape primitive that can be created using 16-bit greyscale PNG heightmaps.
  Create the Renderable with CreateLandscape or CreateLandscapeFromFile

* NEW: The `example\landscape` example project was added to show landscape creation
  via heightmap 16-bit PNG files.


Version v0.3.1
==============

* BUG: Fixed RenderSystem.OnRemoveEntity() so that it correctly creates a new
  slice for surviving entities that is empty.

* MISC: scene/BasicSceneManager got a new function: MapEntities() to iterate
  over Entity objects in the scene.

* MISC: scene/BasicEntity got a new function: CreateCollidersFromComponent()
  to create collision objects.

* BUG: `cmd/compeditor` now embeds the Oswald-Heavy font from eweygewey so that
  it doesn't have to locate it at runtime and is now more pleasant to use
  with `go install`.


Version v0.3.0
==============

* APIBREAK: Many `fizzle/component` changes, including API breaks.

* APIBREAK: Added a Material struct and a pointer to one Renderable. All
  material settings were pulled from RenderableCore and placed in Material.

* APIBREAK: Specific shader uniforms were added for diffuse, normals and specular
  textures and the basic and basicSkinned shaders were updated to use the
  respective texture from the new Material structure for each of these. The old
  []Tex array has been renamed to []CustomTex for custom textures not covered
  by the standard types above.

* NEW: 'HAS_BONES' uniform float in shaders now identifies whether or not
  a skeleton is present in the renderable.

* NEW: Basic, BasicSkinned, Color and ColorText shaders are now built into the
  `renderer/forward` package. Look for the create functions there. The shaders
  have been removed from the `examples/assets/forwardshaders` directory.

* NEW: DiffuseUnlit shader was added to the built in list of shaders.

* NEW: A `scene` package that contains bare-bone implementations of an entity
  system and provides common interfaces to use.

* NEW: A new example called `testscene` which shows off the new `scene` package
  and displays a scene the client can move around.

* BUG: Fixed skeletal animation in basicSkinned shader for bone id 0 not
  being transformed.

* BUG: Many fixes to `cmd/compeditor` and broader support for features
  found in `fizzle/component`.

* BUG: Improved the specular component for the basic and basicSkinned shaders.


Version v0.2.0
==============

* APIBREAK: Many `fizzle/component` changes, including API breaks.
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
