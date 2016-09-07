#version 330
precision highp float;

const int MAX_LIGHTS=4;
const int MAX_BONES=32;

uniform mat4 MVP_MATRIX;
uniform mat4 M_MATRIX;
uniform mat4 V_MATRIX;
uniform mat4 MV_MATRIX;
uniform vec3 CAMERA_WORLD_POSITION;
uniform mat4 SHADOW_MATRIX[MAX_LIGHTS];
in vec3 VERTEX_POSITION;
in vec3 VERTEX_NORMAL;
in vec3 VERTEX_TANGENT;
in vec2 VERTEX_UV_0;

out vec3 vs_normal_model;
out vec3 vs_position_model;
out vec3 vs_position_view;
out vec3 vs_tangent;
out vec2 vs_tex0_uv;
out vec3 vs_camera_eye;
out vec4 vs_shadow_coord[4];

void main()
{
	vec4 vertex4 = vec4(VERTEX_POSITION, 1.0);
	mat3 vs_normal_mat = transpose(inverse(mat3(M_MATRIX)));

	vs_normal_model = vs_normal_mat * VERTEX_NORMAL;
	vs_position_model = vec3(M_MATRIX * vertex4);
	vs_position_view = vec3(MV_MATRIX * vertex4);
	vs_camera_eye = vec3(V_MATRIX * vec4(CAMERA_WORLD_POSITION,1.0));
	vs_tangent = mat3(M_MATRIX) * VERTEX_TANGENT;
	vs_tex0_uv = VERTEX_UV_0;

	/* handle the shadow coordinates unrolled since for loop indexing can be problematic */
	vs_shadow_coord[0] = (SHADOW_MATRIX[0] * M_MATRIX) * vertex4;
	vs_shadow_coord[1] = (SHADOW_MATRIX[1] * M_MATRIX) * vertex4;
	vs_shadow_coord[2] = (SHADOW_MATRIX[2] * M_MATRIX) * vertex4;
	vs_shadow_coord[3] = (SHADOW_MATRIX[3] * M_MATRIX) * vertex4;

	gl_Position = MVP_MATRIX * vertex4;
}
