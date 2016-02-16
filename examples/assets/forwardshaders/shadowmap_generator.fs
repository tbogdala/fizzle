#version 330
precision highp float;

out vec4 frag_color;

void main (void) {
  frag_color = vec4(gl_FragCoord.z);
}
