#version 330
precision highp float;

uniform mat4 MVP_MATRIX;
uniform mat4 M_MATRIX;
uniform mat4 V_MATRIX;
uniform mat4 MV_MATRIX;
uniform mat4 SHADOW_MATRIX[4];

in vec3 VERTEX_POSITION;
in vec3 VERTEX_NORMAL;
in vec3 VERTEX_TANGENT;
in vec2 VERTEX_UV_0;

out vec3 vs_position;
out vec3 vs_normal;
out vec3 vs_tangent;
out vec2 vs_tex0_uv;
out vec3 camera_eye;
out vec4 vs_shadow_coord[4];

void main()
{
  vs_position = VERTEX_POSITION;
	vs_tangent = normalize(VERTEX_TANGENT);
	vs_normal = normalize(VERTEX_NORMAL);
  vs_tex0_uv = VERTEX_UV_0;

  mat3 camRot = mat3(V_MATRIX);
  vec3 d = vec3(V_MATRIX[3]);
  camera_eye = -d * camRot;

  /* handle the shadow coordinates unrolled since for loop indexing can be problematic */
  vs_shadow_coord[0] = (SHADOW_MATRIX[0] * M_MATRIX) * vec4(VERTEX_POSITION, 1.0);
  vs_shadow_coord[1] = (SHADOW_MATRIX[1] * M_MATRIX) * vec4(VERTEX_POSITION, 1.0);
  vs_shadow_coord[2] = (SHADOW_MATRIX[2] * M_MATRIX) * vec4(VERTEX_POSITION, 1.0);
  vs_shadow_coord[3] = (SHADOW_MATRIX[3] * M_MATRIX) * vec4(VERTEX_POSITION, 1.0);

  gl_Position = MVP_MATRIX * vec4(VERTEX_POSITION, 1.0);
}
