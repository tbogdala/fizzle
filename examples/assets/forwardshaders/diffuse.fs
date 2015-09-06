#version 330
precision highp float;

uniform mat4 MV_MATRIX;
uniform mat4 MVP_MATRIX;
uniform vec4 MATERIAL_DIFFUSE;

uniform vec3 CAMERA_WORLD_POSITION;
uniform vec4 LIGHT_DIFFUSE[4];
uniform float LIGHT_INTENSITY[4];
uniform vec3 LIGHT_DIRECTION[4];
uniform float LIGHT_SPECULAR_POWER[4];

in vec3 vs_position;
in vec3 vs_normal;

//uniform float LIGHT_ATTENUATION[4];
//uniform int LIGHT_COUNT;
// need light position
//    mat ambient/spec/shini
//    global ambient

out vec4 frag_color;

// lighting code inspired from: http://ogldev.atspace.co.uk/www/tutorial36/tutorial36.html

vec4 CalcDirectionalLight(vec3 world_pos, vec3 normal)
{
  // FIXME: hardcoded for now since there's no proper material support yet.
  const float gMatSpecularIntensity = 0.10f;

  vec4 ambient_color = LIGHT_DIFFUSE[0] * LIGHT_INTENSITY[0];
  float diffuse_factor = dot(normal, -LIGHT_DIRECTION[0]);
  
  vec4 diffuse_color  = vec4(0, 0, 0, 0);
  vec4 specular_color = vec4(0, 0, 0, 0);

  if (diffuse_factor > 0) {
    diffuse_color = LIGHT_DIFFUSE[0] * LIGHT_INTENSITY[0] * diffuse_factor;

      vec3 VertexToEye = normalize(CAMERA_WORLD_POSITION - world_pos);
      vec3 LightReflect = normalize(reflect(LIGHT_DIRECTION[0], normal));

      float SpecularFactor = dot(VertexToEye, LightReflect);
      SpecularFactor = pow(SpecularFactor, LIGHT_SPECULAR_POWER[0]);
      if (SpecularFactor > 0) {
        specular_color = LIGHT_DIFFUSE[0] * gMatSpecularIntensity * SpecularFactor;
      }

      vec4 SpecularColor = vec4(0.0);
  }

  return (ambient_color + diffuse_color + specular_color);
}


void main()
{
    frag_color = MATERIAL_DIFFUSE * CalcDirectionalLight(vs_position, vs_normal);
    //frag_color = vec4(1.0,1.0,1.0,1.0);
}
