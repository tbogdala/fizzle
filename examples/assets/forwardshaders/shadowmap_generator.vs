#version 330
precision highp float;

uniform mat4 M_MATRIX;
uniform mat4 SHADOW_VP_MATRIX;
in vec4 VERTEX_POSITION;

/* shadow pass */
void main() {
  gl_Position = SHADOW_VP_MATRIX * M_MATRIX * VERTEX_POSITION;
}
