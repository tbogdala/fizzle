#version 330
precision highp float;

uniform mat4 MV_MATRIX;
uniform mat4 V_MATRIX;

uniform vec4 MATERIAL_DIFFUSE;
uniform vec4 MATERIAL_SPECULAR;
uniform float MATERIAL_SHININESS;
uniform sampler2D MATERIAL_TEX_0;
uniform sampler2D MATERIAL_TEX_1;

uniform vec3 LIGHT_POSITION[4];
uniform vec4 LIGHT_DIFFUSE[4];
uniform float LIGHT_DIFFUSE_INTENSITY[4];
uniform float LIGHT_AMBIENT_INTENSITY[4];
uniform vec3 LIGHT_DIRECTION[4];
uniform float LIGHT_ATTENUATION[4];
uniform int LIGHT_COUNT;

in vec3 vs_position;
in vec3 vs_normal;
in vec3 vs_tangent;
in vec2 vs_tex0_uv;
in vec3 camera_eye;

out vec4 frag_color;

vec4 CalcADSLights(vec3 p, vec3 n)
{
  // eye-space
  vec4 P_view = MV_MATRIX * vec4(p, 1.0);
  vec3 V_view = normalize(-P_view.xyz);

  // eye-space normal
	mat3 normal_matrix = mat3(MV_MATRIX);
	vec3 N_view = normalize(normal_matrix * n);

  vec4 ambient_color = vec4(0, 0, 0, 0);
  vec4 diffuse_color  = vec4(0, 0, 0, 0);
  vec4 specular_color = vec4(0, 0, 0, 0);

  for (int i=0; i<LIGHT_COUNT; i++) {
    const float Epsilon = 0.0001;

    // eye-space light vector
    vec3 L_view;
    float attenuation;

    // if the direction is not set, then assume we have a positional point light.
    if (abs(LIGHT_DIRECTION[i].x) < Epsilon && abs(LIGHT_DIRECTION[i].y) < Epsilon && abs(LIGHT_DIRECTION[i].z) < Epsilon) {
      vec3 L_pos_view = (V_MATRIX * vec4(LIGHT_POSITION[i], 1.0)).xyz;
      vec3 L_distance = L_pos_view - P_view.xyz;
      attenuation = 1.0 / (1.0 +  LIGHT_ATTENUATION[i] * pow(length(L_distance),2));
      L_view = normalize(L_distance);
    }

    // this is the directional light branch where attenuation is a little simpler
    else {
      attenuation = 1.0 / (1.0 +  LIGHT_ATTENUATION[i]);
      L_view = normalize(-LIGHT_DIRECTION[i]);
    }

    // calculate R by reflecting -L around the plane dfeined by N
    vec3 R = reflect(-L_view, N_view);

    // calculate diffuse and specular
		float light_intensity = max(dot(N_view, L_view), 0.0);

		float specular_intensity = 0.0;
		if (light_intensity > 0.0 && MATERIAL_SHININESS > Epsilon) {
      specular_intensity = attenuation * pow(max(dot(R, V_view), 0.0), MATERIAL_SHININESS);
    }

    ambient_color += LIGHT_DIFFUSE[i] * LIGHT_AMBIENT_INTENSITY[i];
    diffuse_color += LIGHT_DIFFUSE[i] * LIGHT_DIFFUSE_INTENSITY[i] * light_intensity * attenuation;
    specular_color += MATERIAL_SPECULAR[i] * LIGHT_DIFFUSE_INTENSITY[i] * specular_intensity;
  }

  vec4 result = ambient_color + diffuse_color + specular_color;
  return clamp(result, 0.0, 1.0);
}

void main()
{
  vec3 T = normalize(vs_tangent - dot(vs_tangent, vs_normal) * vs_normal);
	vec3 BT = cross(T, vs_normal);
	vec3 bump_normal = texture(MATERIAL_TEX_1, vs_tex0_uv).rgb;
  bump_normal = 2.0 * bump_normal - vec3(1.0, 1.0, 1.0);
  mat3 TBN = mat3(T, BT, vs_normal);
	vec3 final_bumped_normal = normalize(TBN * bump_normal);

  vec4 texture_color = texture(MATERIAL_TEX_0, vs_tex0_uv).rgba;
  frag_color =  MATERIAL_DIFFUSE * texture_color * CalcADSLights(vs_position, final_bumped_normal);
}
