#version 330
precision highp float;

const int MAX_LIGHTS=4;

uniform mat4 V_MATRIX;
uniform vec4 MATERIAL_DIFFUSE;
uniform vec4 MATERIAL_SPECULAR;
uniform float MATERIAL_SHININESS;
uniform sampler2D MATERIAL_TEX_0;
uniform sampler2D MATERIAL_TEX_1;
uniform float MATERIAL_TEX_0_VALID;
uniform float MATERIAL_TEX_1_VALID;
uniform sampler2DShadow SHADOW_MAPS[4];

uniform vec3 LIGHT_POSITION[MAX_LIGHTS];
uniform vec4 LIGHT_DIFFUSE[MAX_LIGHTS];
uniform float LIGHT_DIFFUSE_INTENSITY[MAX_LIGHTS];
uniform float LIGHT_AMBIENT_INTENSITY[MAX_LIGHTS];
uniform float LIGHT_SPECULAR_INTENSITY[MAX_LIGHTS];
uniform vec3 LIGHT_DIRECTION[MAX_LIGHTS];
uniform float LIGHT_CONST_ATTENUATION[MAX_LIGHTS];
uniform float LIGHT_LINEAR_ATTENUATION[MAX_LIGHTS];
uniform float LIGHT_QUADRATIC_ATTENUATION[MAX_LIGHTS];
uniform float LIGHT_STRENGTH[MAX_LIGHTS];
uniform int LIGHT_COUNT;
uniform int SHADOW_COUNT;

in vec3 vs_normal_model;
in vec3 vs_position_model;
in vec3 vs_position_view;
in vec3 vs_tangent;
in vec2 vs_tex0_uv;
in vec3 vs_camera_eye;
in vec4 vs_shadow_coord[4];

out vec4 frag_color;

vec4 CalcShadowFactor() {
	float shadow = 1.0;
	if (SHADOW_COUNT > 0) {
		shadow = 0.0;
		shadow += textureProj(SHADOW_MAPS[0], vs_shadow_coord[0]);
		if (SHADOW_COUNT > 1) {
			shadow += textureProj(SHADOW_MAPS[1], vs_shadow_coord[1]);
		}
		if (SHADOW_COUNT > 2) {
			shadow += textureProj(SHADOW_MAPS[2], vs_shadow_coord[2]);
		}
		if (SHADOW_COUNT > 3) {
			shadow += textureProj(SHADOW_MAPS[3], vs_shadow_coord[3]);
		}
		shadow = shadow / SHADOW_COUNT;
	}
	return vec4(shadow,shadow,shadow,1.0);
}


vec3 CalcADSLights(vec3 v_model, vec3 n_model, vec3 color)
{
	vec3 scattered_light = vec3(0.0);
	vec3 reflected_light = vec3(0.0);

	for (int i=0; i<MAX_LIGHTS; i++) {
  		if (i >= LIGHT_COUNT) {
			break;
		}

		vec3 incidence;
		float attenuation = LIGHT_STRENGTH[i];
		vec3 light_direction = LIGHT_DIRECTION[i]; // in world space

		if (light_direction.x == 0.0 && light_direction.y == 0.0 && light_direction.z == 0.0) {
			// point light
			light_direction = LIGHT_POSITION[i] - v_model;
			float distance = length(light_direction);
		
			attenuation = LIGHT_STRENGTH[i] / (1.0 + 
				(LIGHT_CONST_ATTENUATION[i] +
				 LIGHT_LINEAR_ATTENUATION[i] * distance +
				 LIGHT_QUADRATIC_ATTENUATION[i] * distance * distance));
			
			light_direction = light_direction / distance;	
			incidence = light_direction;
	        } else {
			// directional light
			light_direction = normalize(light_direction);
			incidence = -light_direction;
		}

		float specularF = 0.0;
		float diffuseF = max(0.0, dot(n_model, incidence));
		if (MATERIAL_SHININESS != 0.0 && diffuseF != 0.0) {
			vec3 reflection = reflect(incidence, n_model);
			vec3 s_to_camera = normalize(vs_camera_eye - v_model);
			specularF = pow(max(0.0, dot(s_to_camera, reflection)), MATERIAL_SHININESS);
		}

		vec3 ambient = LIGHT_DIFFUSE[i].rgb * LIGHT_AMBIENT_INTENSITY[i] * attenuation;
		vec3 diffuse = LIGHT_DIFFUSE[i].rgb * LIGHT_DIFFUSE_INTENSITY[i] * diffuseF * attenuation;
		vec3 specular = LIGHT_DIFFUSE[i].rgb * LIGHT_SPECULAR_INTENSITY[i] * specularF * attenuation;

		scattered_light += ambient + diffuse;
		reflected_light += specular; 
	}

	return min(color * scattered_light + reflected_light, vec3(1.0));
}


void main()
{
	vec4 color = MATERIAL_DIFFUSE; 
	if (MATERIAL_TEX_0_VALID > 0.0) {
		color *= texture(MATERIAL_TEX_0, vs_tex0_uv);
	}

  	vec4 shadowFactor = CalcShadowFactor();

	vec3 normal = vs_normal_model;
	if (MATERIAL_TEX_1_VALID > 0.0) {
		vec3 T = normalize(vs_tangent - dot(vs_tangent, vs_normal_model) * vs_normal_model);
		vec3 BT = cross(T, vs_normal_model);
		vec3 bump_normal = texture(MATERIAL_TEX_1, vs_tex0_uv).rgb;
		bump_normal = 2.0 * bump_normal - vec3(1.0, 1.0, 1.0);
		mat3 TBN = mat3(T, BT, vs_normal_model);
		normal = TBN * bump_normal;
	}

	frag_color = vec4(shadowFactor.rgb * CalcADSLights(vs_position_model, normalize(normal), color.rgb), 1.0);
}
