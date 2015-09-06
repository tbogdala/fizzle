#version 330
precision highp float;

uniform mat4 MVP_MATRIX;
uniform mat4 MV_MATRIX;
in vec3 VERTEX_POSITION;
in vec3 VERTEX_NORMAL;

out vec3 vs_position;
out vec3 vs_eye_normal;

void main()
{
  vs_position = VERTEX_POSITION;
  vs_eye_normal = normalize(mat3(MV_MATRIX) * VERTEX_NORMAL);
  gl_Position = MVP_MATRIX * vec4(VERTEX_POSITION, 1.0);
}
