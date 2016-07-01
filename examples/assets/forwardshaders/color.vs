#version 330
precision highp float;

uniform mat4 MVP_MATRIX;

in vec3 VERTEX_POSITION;

void main(void) {
	gl_Position = MVP_MATRIX * vec4(VERTEX_POSITION, 1.0);
}
