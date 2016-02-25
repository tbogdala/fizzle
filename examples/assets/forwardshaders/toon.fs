#version 330
precision highp float;

uniform vec4 MATERIAL_DIFFUSE;
uniform vec4 MATERIAL_SPECULAR;
uniform float MATERIAL_SHININESS;

uniform vec3 LIGHT_POSITION[4];
uniform vec4 LIGHT_DIFFUSE[4];
uniform float LIGHT_DIFFUSE_INTENSITY[4];
uniform float LIGHT_SPECULAR_INTENSITY[4];
uniform float LIGHT_AMBIENT_INTENSITY[4];
uniform vec3 LIGHT_DIRECTION[4];
uniform int LIGHT_COUNT;
uniform vec3 CAMERA_WORLD_POSITION;

in vec3 vs_vert_color;
in vec4 w_position;
in vec3 w_normal;

out vec4 frag_color;

const float Epsilon = 0.0001;

// Gamma correction routines
const float gamma = 2.2;
vec3 toLinear(vec3 c) {
    return pow(c, vec3(gamma));
}
vec4 toLinear(vec4 c) {
    return pow(c, vec4(gamma));
}
vec3 toGamma(vec3 c) {
    return pow(c, vec3(1.0 / gamma));
}
vec4 toGamma(vec4 c) {
    return pow(c, vec4(1.0 / gamma));
}

// return a fixed float value based on incoming value x
float step4(float x) {
    const float A = 0.1;
    const float B = 0.3;
    const float C = 0.6;
    const float D = 1.0;
    if (x < A) return 0.0;
    else if (x < B) return B;
    else if (x < C) return C;
    else return D;
}

// return a fixed float value based on incoming value x
float step3(float x) {
    const float A = 0.2;
    const float B = 0.6;
    const float C = 1.0;
    if (x < A) return 0.0;
    else if (x < B) return B;
    else return C;
}


float step(float edge, float x) {
    return x < edge ? 0.0 : 1.0;
}

vec4 Toon(int light_i)
{
    // apply gamma correction
    vec3 l_light_color = toLinear(LIGHT_DIFFUSE[light_i].rgb);
    vec3 l_base_color = toLinear(MATERIAL_DIFFUSE.xyz);
    vec3 l_specular_color = toLinear(MATERIAL_SPECULAR.xyz);

    // world space normal
    vec3 N = normalize(w_normal.xyz);

    // calculate the direction towards the light in world space
    vec3 L;
    // if light direction is not set, calculate it from the position
    if (abs(LIGHT_DIRECTION[light_i].x) < Epsilon && abs(LIGHT_DIRECTION[light_i].y) < Epsilon && abs(LIGHT_DIRECTION[light_i].z) < Epsilon) {
      L = normalize(LIGHT_POSITION[light_i] - w_position.xyz);
    } else {
      // otherwise we just use the direction here
      L = normalize(-LIGHT_DIRECTION[light_i]);
    }

    // get the diffuse intensity
    float Id = max(0.0, dot(N, L));

    // TOON the diffuse value
    float diffFactor = step4(Id);

    // calculate the specular coefficient appropriate based on diffuse intensity
    float specFactor = 0.0;
    if (Id > 0.0 && MATERIAL_SHININESS > 0.0) {
        // calculate the specular
        vec3 R = normalize(reflect(-L, N));

        // calculate surface to camera in world space
        vec3 E = normalize(CAMERA_WORLD_POSITION - w_position.xyz);

        specFactor = pow(max(0.0, dot(E, R)), MATERIAL_SHININESS);
        clamp(specFactor, 0.0, 1.0);

        // TOON the specular
        specFactor = step(0.5, specFactor);
    }

    // calculate the color based on the diffuse intensity and copy
    // over the alpha setting of the base color.
    vec3 final_ambient = LIGHT_AMBIENT_INTENSITY[light_i] * l_base_color * l_light_color;
    vec3 final_diffuse = LIGHT_DIFFUSE_INTENSITY[light_i] * l_base_color * l_light_color * diffFactor;
    vec3 final_specular = LIGHT_SPECULAR_INTENSITY[light_i] * l_specular_color * l_light_color * specFactor;

    return toGamma(vec4(final_ambient + final_diffuse + final_specular, MATERIAL_DIFFUSE.a));
}

void main()
{
  vec4 final_color;
  for (int i=0; i<LIGHT_COUNT; i++) {
    final_color += Toon(i);
  }
  frag_color = final_color;
}
