#version 330
precision highp float;

uniform vec4 MATERIAL_DIFFUSE;

out vec4 frag_color;

void main (void) {
	frag_color = MATERIAL_DIFFUSE;
}
