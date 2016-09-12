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
uniform mat4 BONES[MAX_BONES];
uniform float HAS_BONES;
in vec3 VERTEX_POSITION;
in vec3 VERTEX_NORMAL;
in vec3 VERTEX_TANGENT;
in vec2 VERTEX_UV_0;
in vec4 VERTEX_BONE_IDS;
in vec4 VERTEX_BONE_WEIGHTS;

out vec3 vs_normal_model;
out vec3 vs_position_model;
out vec3 vs_position_view;
out vec3 vs_tangent;
out vec2 vs_tex0_uv;
out vec3 vs_camera_eye;
out vec4 vs_shadow_coord[4];


struct skinnedData {
	mat4 matrix;
	vec4 position;
	vec3 normal;
	vec3 tangent;
};

skinnedData calculateSkinnedData() {
	skinnedData data;
	data.matrix =  BONES[int(VERTEX_BONE_IDS.x)] * VERTEX_BONE_WEIGHTS.x;
	data.matrix += BONES[int(VERTEX_BONE_IDS.y)] * VERTEX_BONE_WEIGHTS.y;
	data.matrix += BONES[int(VERTEX_BONE_IDS.z)] * VERTEX_BONE_WEIGHTS.z;
	data.matrix += BONES[int(VERTEX_BONE_IDS.w)] * VERTEX_BONE_WEIGHTS.w;

	data.position =  data.matrix * vec4(VERTEX_POSITION, 1.0);
	data.position.w = 1.0;

	vec4 temp_skinned_norm = data.matrix * vec4(VERTEX_NORMAL, 0.0);
	data.normal = temp_skinned_norm.xyz;

	vec4 temp_skinned_tangent = data.matrix * vec4(VERTEX_TANGENT, 0.0);
	data.tangent = temp_skinned_tangent.xyz;

	return data;
}


void main()
{
	skinnedData skinned;
	if (HAS_BONES > 0.0) {
		skinned = calculateSkinnedData();
	} else {
		skinned.position = vec4(VERTEX_POSITION, 1.0);
		skinned.normal = VERTEX_NORMAL;
		skinned.tangent = VERTEX_TANGENT;
	}

	mat3 vs_normal_mat = transpose(inverse(mat3(M_MATRIX)));

	vs_normal_model = vs_normal_mat * skinned.normal;
	vs_position_model = vec3(M_MATRIX * skinned.position);
	vs_position_view = vec3(MV_MATRIX * skinned.position);
	vs_camera_eye = vec3(V_MATRIX * vec4(CAMERA_WORLD_POSITION,1.0));
	vs_tangent = mat3(M_MATRIX) * skinned.tangent;
	vs_tex0_uv = VERTEX_UV_0;

	/* handle the shadow coordinates unrolled since for loop indexing can be problematic */
	vs_shadow_coord[0] = (SHADOW_MATRIX[0] * M_MATRIX) * skinned.position;
	vs_shadow_coord[1] = (SHADOW_MATRIX[1] * M_MATRIX) * skinned.position;
	vs_shadow_coord[2] = (SHADOW_MATRIX[2] * M_MATRIX) * skinned.position;
	vs_shadow_coord[3] = (SHADOW_MATRIX[3] * M_MATRIX) * skinned.position;

	gl_Position = MVP_MATRIX * skinned.position;
}
