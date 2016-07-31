#version 330
precision highp float;

const int MAX_LIGHTS=4;

uniform mat4 V_MATRIX;
uniform vec4 MATERIAL_DIFFUSE;
uniform vec4 MATERIAL_SPECULAR;
uniform float MATERIAL_SHININESS;
uniform sampler2D MATERIAL_TEX_0;
uniform float MATERIAL_TEX_0_VALID;

uniform vec3 LIGHT_POSITION[MAX_LIGHTS];
uniform vec4 LIGHT_DIFFUSE[MAX_LIGHTS];
uniform float LIGHT_DIFFUSE_INTENSITY[MAX_LIGHTS];
uniform float LIGHT_AMBIENT_INTENSITY[MAX_LIGHTS];
uniform vec3 LIGHT_DIRECTION[MAX_LIGHTS];
uniform float LIGHT_CONST_ATTENUATION[MAX_LIGHTS];
uniform float LIGHT_LINEAR_ATTENUATION[MAX_LIGHTS];
uniform float LIGHT_QUADRATIC_ATTENUATION[MAX_LIGHTS];
uniform float LIGHT_STRENGTH[MAX_LIGHTS];
uniform int LIGHT_COUNT;

smooth in vec3 vs_normal_model;
in vec3 vs_position_model;
in vec3 vs_position_view;
in vec2 vs_tex0_uv;
in vec3 vs_camera_eye;

out vec4 frag_color;

vec3 CalcADSLights(vec3 v_model, vec3 n_model, vec3 color)
{
	vec3 scattered_light = vec3(0.0);
	vec3 reflected_light = vec3(0.0);

	for (int i=0; i<MAX_LIGHTS; i++) {
  		if (i >= LIGHT_COUNT) {
			break;
		}

		vec3 s;
		float attenuation = 1.0;
		vec3 light_direction = LIGHT_DIRECTION[i]; // in world space

		if (light_direction.x == 0.0 && light_direction.y == 0.0 && light_direction.z == 0.0) {
			// point light
			light_direction = LIGHT_POSITION[i] - v_model;
			float distance = length(light_direction);
		
			attenuation = LIGHT_STRENGTH[i] / (1.0 + 
				(LIGHT_CONST_ATTENUATION[i] +
				 LIGHT_LINEAR_ATTENUATION[i] * distance +
				 LIGHT_QUADRATIC_ATTENUATION[i] * distance * distance));
			
			light_direction = normalize(light_direction / distance);	
			s = normalize(light_direction);
	        } else {
			// directional light
			light_direction = normalize(light_direction);
			s = -light_direction;
		}

		float specularF = 0.0;
		float diffuseF = max(0.0, dot(n_model, s));
		if (diffuseF != 0.0) {
			vec3 incidence = s;
			vec3 reflection = reflect(incidence, n_model);
			vec3 s_to_camera = normalize(vs_camera_eye - v_model);
			specularF = pow(max(0.0, dot(s_to_camera, reflection)), MATERIAL_SHININESS);
		}

		vec3 ambient = LIGHT_DIFFUSE[i].rgb * LIGHT_AMBIENT_INTENSITY[i] * attenuation;
		vec3 diffuse = LIGHT_DIFFUSE[i].rgb * LIGHT_DIFFUSE_INTENSITY[i] * diffuseF * attenuation;
		vec3 specular = LIGHT_DIFFUSE[i].rgb * specularF * attenuation;

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
	frag_color = vec4(CalcADSLights(vs_position_model, normalize(vs_normal_model), color.rgb), 1.0);
}
