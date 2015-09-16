#version 330
precision highp float;

uniform mat4 MV_MATRIX;
uniform mat4 MVP_MATRIX;
uniform vec4 MATERIAL_DIFFUSE;
uniform vec4 MATERIAL_SPECULAR;
uniform float MATERIAL_SHININESS;

uniform vec3 CAMERA_WORLD_POSITION;
uniform vec4 LIGHT_POSITION[4];
uniform vec4 LIGHT_DIFFUSE[4];
uniform float LIGHT_DIFFUSE_INTENSITY[4];
uniform float LIGHT_AMBIENT_INTENSITY[4];
uniform vec3 LIGHT_DIRECTION[4];
uniform int LIGHT_COUNT;

in vec3 vs_position;
in vec3 vs_eye_normal;

out vec4 frag_color;

vec4 CalcADSLights(vec3 world_pos, vec3 eye_normal)
{
  const float Epsilon = 0.0001;
  vec4 ambient_color = vec4(0, 0, 0, 0);
  vec4 diffuse_color  = vec4(0, 0, 0, 0);
  vec4 specular_color = vec4(0, 0, 0, 0);

  vec4 eye_vertex = MV_MATRIX * vec4(world_pos, 1.0);

  for (int i=0; i<LIGHT_COUNT; i++) {
    vec3 s;

    // if light direction is not set, calculate it from the position
    if (LIGHT_DIRECTION[i].x < Epsilon && LIGHT_DIRECTION[i].y < Epsilon && LIGHT_DIRECTION[i].z < Epsilon) {
      s = normalize(vec3(LIGHT_POSITION[i]-eye_vertex));
    } else {
      // otherwise we just use the direction here
      s = normalize(-LIGHT_DIRECTION[i]);
    }
    vec3 v = normalize(-eye_vertex.xyz);
    vec3 r = reflect(-s, eye_normal);
    ambient_color += LIGHT_DIFFUSE[i] * LIGHT_AMBIENT_INTENSITY[i];
    float sDotN = max(dot(s,eye_normal), 0.0);
    diffuse_color += LIGHT_DIFFUSE[i] * LIGHT_DIFFUSE_INTENSITY[i] * sDotN;
    if( sDotN > 0.0 ) {
      specular_color += MATERIAL_SPECULAR[i] * pow(max(dot(r,v), 0.0), MATERIAL_SHININESS);
    }
  }
  return (ambient_color + diffuse_color + specular_color);
}


void main()
{
  frag_color = MATERIAL_DIFFUSE * CalcADSLights(vs_position, vs_eye_normal);
}
