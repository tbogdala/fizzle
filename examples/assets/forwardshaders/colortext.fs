#version 330
precision highp float;

uniform sampler2D MATERIAL_TEX_0;
uniform vec4 MATERIAL_DIFFUSE;

in vec2 vs_tex0_uv;
out vec4 frag_color;

void main (void) {
	// Modify the color's alpha based on the alpha map used in
	// in sample2D channel 0
	float alph = texture(MATERIAL_TEX_0, vs_tex0_uv).r;
	if (alph < 0.10) {
	    discard;
	}
	frag_color = vec4(MATERIAL_DIFFUSE.rgb, MATERIAL_DIFFUSE.a);
}
