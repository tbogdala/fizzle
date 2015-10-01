#version 330
precision highp float;

uniform mat4 MVP_MATRIX;
uniform mat4 M_MATRIX;
uniform mat4 V_MATRIX;
uniform mat4 MV_MATRIX;
in vec3 VERTEX_POSITION;
in vec3 VERTEX_NORMAL;
in vec3 VERTEX_TANGENT;
in vec2 VERTEX_UV_0;

out vec3 vs_position;
out vec3 vs_normal;
out vec3 vs_tangent;
out vec2 vs_tex0_uv;
out vec3 camera_eye;

void main()
{
  vs_position = VERTEX_POSITION;
	vs_tangent = normalize(VERTEX_TANGENT);
	vs_normal = normalize(VERTEX_NORMAL);
  vs_tex0_uv = VERTEX_UV_0;
  
  mat3 camRot = mat3(V_MATRIX);
  vec3 d = vec3(V_MATRIX[3]);
  camera_eye = -d * camRot;

  gl_Position = MVP_MATRIX * vec4(VERTEX_POSITION, 1.0);
}
