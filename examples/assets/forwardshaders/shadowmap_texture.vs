#version 330
precision highp float;

uniform mat4 MVP_MATRIX;

in vec3 VERTEX_POSITION;
in vec2 VERTEX_UV_0;

out vec2 vs_tex0_uv;

void main(void) {
	gl_Position = MVP_MATRIX * vec4(VERTEX_POSITION, 1.0);
	vs_tex0_uv = VERTEX_UV_0;
}
