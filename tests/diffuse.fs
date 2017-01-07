#version 330
precision highp float;

uniform vec4 MATERIAL_DIFFUSE;
uniform vec4 MATERIAL_SPECULAR;
uniform float MATERIAL_SHININESS;

uniform vec3 LIGHT_POSITION[4];
uniform vec4 LIGHT_DIFFUSE[4];
uniform float LIGHT_DIFFUSE_INTENSITY[4];
uniform float LIGHT_AMBIENT_INTENSITY[4];
uniform vec3 LIGHT_DIRECTION[4];
uniform int LIGHT_COUNT;

in vec3 vs_normal_model;
in vec3 vs_position_model;
in vec3 camera_eye;

out vec4 frag_color;

vec4 CalcADSLights(vec3 v_model, vec3 n_model)
{
  const float Epsilon = 0.0001;
  vec4 ambient_color = vec4(0, 0, 0, 0);
  vec4 diffuse_color  = vec4(0, 0, 0, 0);
  vec4 specular_color = vec4(0, 0, 0, 0);

  vec3 s;
  for (int i=0; i<LIGHT_COUNT; i++) {
    // if light direction is not set, calculate it from the position
    if (abs(LIGHT_DIRECTION[i].x) < Epsilon && abs(LIGHT_DIRECTION[i].y) < Epsilon && abs(LIGHT_DIRECTION[i].z) < Epsilon) {
      s = vec3(LIGHT_POSITION[i] - v_model);
    } else {
      // otherwise we just use the direction here
      s = -LIGHT_DIRECTION[i];
    }

    vec3 sN = normalize(s);
    float sDotN = dot(n_model, sN);
    float brightness = clamp(sDotN / length(s), 0.0, 1.0);

    ambient_color += LIGHT_DIFFUSE[i] * LIGHT_AMBIENT_INTENSITY[i];
    diffuse_color += LIGHT_DIFFUSE[i] * LIGHT_DIFFUSE_INTENSITY[i] * brightness;

    if( sDotN > 0.0 && MATERIAL_SHININESS > Epsilon) {
      vec3 r = reflect(-sN, n_model);
      vec3 v = normalize(camera_eye - v_model);
      specular_color += MATERIAL_SPECULAR * pow(max(0.0, dot(v,r)), MATERIAL_SHININESS);
    }
  }

  return (ambient_color + diffuse_color + specular_color);
}


void main()
{
  frag_color =  MATERIAL_DIFFUSE * CalcADSLights(vs_position_model, vs_normal_model);
}
