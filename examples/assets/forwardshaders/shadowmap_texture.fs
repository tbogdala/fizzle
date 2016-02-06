#version 330
precision highp float;

uniform sampler2D MATERIAL_TEX_0;
uniform vec4 MATERIAL_DIFFUSE;

in vec2 vs_tex0_uv;

out vec4 frag_color;

void main (void) {
	float f=50.0;
	float n = 0.5;
	float z = (2 * n) / (f + n - texture(MATERIAL_TEX_0, vs_tex0_uv).x * (f - n));
	frag_color = MATERIAL_DIFFUSE * vec4(z, z, z, 1.0);
}
