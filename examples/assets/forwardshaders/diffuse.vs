#version 330
precision highp float;

uniform mat4 MVP_MATRIX;
uniform mat4 M_MATRIX;
uniform mat4 V_MATRIX;
in vec3 VERTEX_POSITION;
in vec3 VERTEX_NORMAL;

out vec3 vs_normal_model;
out vec3 vs_position_model;
out vec3 camera_eye;

void main()
{
  mat3 vs_normal_mat = transpose(inverse(mat3(M_MATRIX)));
  vs_normal_model = normalize(vs_normal_mat * VERTEX_NORMAL);
  vs_position_model = vec3(M_MATRIX * vec4(VERTEX_POSITION,1.0));

  mat3 camRot = mat3(V_MATRIX);
  vec3 d = vec3(V_MATRIX[3]);
  camera_eye = -d * camRot;


  gl_Position = MVP_MATRIX * vec4(VERTEX_POSITION, 1.0);
}
