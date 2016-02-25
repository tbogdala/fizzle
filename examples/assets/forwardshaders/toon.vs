#version 330
precision highp float;

uniform mat4 MVP_MATRIX;
uniform mat4 M_MATRIX;
uniform mat4 V_MATRIX;
uniform mat4 MV_MATRIX;
in vec3 VERTEX_POSITION;
in vec3 VERTEX_NORMAL;
in vec3 VERTEX_COLOR;

out vec3 vs_vert_color;
out vec4 w_position;
out vec3 w_normal;

void main()
{
  vs_vert_color = VERTEX_COLOR;

  vec4 vert4 = vec4(VERTEX_POSITION, 1.0);
  mat3 normal_mat = transpose(inverse(mat3(M_MATRIX)));

  w_position = M_MATRIX * vert4;
  w_normal = normal_mat * VERTEX_NORMAL;
  gl_Position = MVP_MATRIX * vert4;
}
