// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package forward

import (
	"github.com/tbogdala/fizzle"
)

const (
	calcSkinnedData = `struct skinnedData {
    	mat4 matrix;
    	vec4 position;
    	vec3 normal;
    	vec3 tangent;
    };

    skinnedData calculateSkinnedData() {
    	skinnedData data;
    	data.matrix =  BONES[int(VERTEX_BONE_IDS.x)] * VERTEX_BONE_WEIGHTS.x;
    	data.matrix += BONES[int(VERTEX_BONE_IDS.y)] * VERTEX_BONE_WEIGHTS.y;
    	data.matrix += BONES[int(VERTEX_BONE_IDS.z)] * VERTEX_BONE_WEIGHTS.z;
    	data.matrix += BONES[int(VERTEX_BONE_IDS.w)] * VERTEX_BONE_WEIGHTS.w;

    	data.position =  data.matrix * vec4(VERTEX_POSITION, 1.0);
    	data.position.w = 1.0;

    	vec4 temp_skinned_norm = data.matrix * vec4(VERTEX_NORMAL, 0.0);
    	data.normal = temp_skinned_norm.xyz;

    	vec4 temp_skinned_tangent = data.matrix * vec4(VERTEX_TANGENT, 0.0);
    	data.tangent = temp_skinned_tangent.xyz;

    	return data;
    }
    `

	calcShadowFactor = `vec4 CalcShadowFactor() {
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
    }`

	calcADSLights = `vec3 CalcADSLights(vec3 v_model, vec3 n_model, vec3 color)
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
    			vec3 reflection = reflect(-incidence, n_model);
    			vec3 s_to_camera = normalize(vs_camera_world - v_model);
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
    `
	/*

	    ____                  _
	   |  _ \                (_)
	   | |_) |   __ _   ___   _    ___
	   |  _ <   / _` | / __| | |  / __|
	   | |_) | | (_| | \__ \ | | | (__
	   |____/   \__,_| |___/ |_|  \___|

	*/

	basicShaderV = `#version 330
    precision highp float;

    const int MAX_LIGHTS=4;
    const int MAX_BONES=32;

    uniform mat4 MVP_MATRIX;
    uniform mat4 M_MATRIX;
    uniform mat4 V_MATRIX;
    uniform mat4 MV_MATRIX;
    uniform vec3 CAMERA_WORLD_POSITION;
    uniform mat4 SHADOW_MATRIX[MAX_LIGHTS];
    in vec3 VERTEX_POSITION;
    in vec3 VERTEX_NORMAL;
    in vec3 VERTEX_TANGENT;
    in vec2 VERTEX_UV_0;

    out vec3 vs_normal_model;
    out vec3 vs_position_model;
    out vec3 vs_position_view;
    out vec3 vs_tangent;
    out vec2 vs_tex0_uv;
    out vec3 vs_camera_world;
    out vec4 vs_shadow_coord[4];

    void main()
    {
    	vec4 vertex4 = vec4(VERTEX_POSITION, 1.0);
    	mat3 vs_normal_mat = transpose(inverse(mat3(M_MATRIX)));

    	vs_normal_model = vs_normal_mat * VERTEX_NORMAL;
			vs_position_model = vec3(M_MATRIX * vertex4);
    	vs_position_view = vec3(MV_MATRIX * vertex4);
    	vs_camera_world = CAMERA_WORLD_POSITION;
    	vs_tangent = mat3(M_MATRIX) * VERTEX_TANGENT;
    	vs_tex0_uv = VERTEX_UV_0;

    	/* handle the shadow coordinates unrolled since for loop indexing can be problematic */
    	vs_shadow_coord[0] = (SHADOW_MATRIX[0] * M_MATRIX) * vertex4;
    	vs_shadow_coord[1] = (SHADOW_MATRIX[1] * M_MATRIX) * vertex4;
    	vs_shadow_coord[2] = (SHADOW_MATRIX[2] * M_MATRIX) * vertex4;
    	vs_shadow_coord[3] = (SHADOW_MATRIX[3] * M_MATRIX) * vertex4;

    	gl_Position = MVP_MATRIX * vertex4;
    }
    `

	basicShaderF = `#version 330
    precision highp float;

    const int MAX_LIGHTS=4;

    uniform mat4 V_MATRIX;
    uniform vec4 MATERIAL_DIFFUSE;
    uniform vec4 MATERIAL_SPECULAR;
    uniform float MATERIAL_SHININESS;
    uniform sampler2D MATERIAL_TEX_DIFFUSE; // dif
    uniform sampler2D MATERIAL_TEX_NORMALS; // norm
    uniform float MATERIAL_TEX_DIFFUSE_VALID;
    uniform float MATERIAL_TEX_NORMALS_VALID;
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
    in vec3 vs_camera_world;
    in vec4 vs_shadow_coord[4];

    out vec4 frag_color;

    ` + calcShadowFactor + `

    ` + calcADSLights + `

    void main()
    {
    	vec4 color = MATERIAL_DIFFUSE;
    	if (MATERIAL_TEX_DIFFUSE_VALID > 0.0) {
    		color *= texture(MATERIAL_TEX_DIFFUSE, vs_tex0_uv);
    	}

    	vec4 shadowFactor = CalcShadowFactor();

    	vec3 normal = vs_normal_model;
    	if (MATERIAL_TEX_NORMALS_VALID > 0.0) {
    		vec3 T = normalize(vs_tangent - dot(vs_tangent, vs_normal_model) * vs_normal_model);
    		vec3 BT = cross(T, vs_normal_model);
    		vec3 bump_normal = texture(MATERIAL_TEX_NORMALS, vs_tex0_uv).rgb;
    		bump_normal = 2.0 * bump_normal - vec3(1.0, 1.0, 1.0);
    		mat3 TBN = mat3(T, BT, vs_normal_model);
    		normal = TBN * bump_normal;
    	}

			frag_color = vec4(shadowFactor.rgb * CalcADSLights(vs_position_model, normalize(normal), color.rgb), 1.0);
    }
    `

	/*

	    ____                  _             _____   _      _                              _
	   |  _ \                (_)           / ____| | |    (_)                            | |
	   | |_) |   __ _   ___   _    ___    | (___   | | __  _   _ __    _ __     ___    __| |
	   |  _ <   / _` | / __| | |  / __|    \___ \  | |/ / | | | '_ \  | '_ \   / _ \  / _` |
	   | |_) | | (_| | \__ \ | | | (__     ____) | |   <  | | | | | | | | | | |  __/ | (_| |
	   |____/   \__,_| |___/ |_|  \___|   |_____/  |_|\_\ |_| |_| |_| |_| |_|  \___|  \__,_|


	*/

	basicSkinnedShaderV = `#version 330
    precision highp float;

    const int MAX_LIGHTS=4;
    const int MAX_BONES=32;

    uniform mat4 MVP_MATRIX;
    uniform mat4 M_MATRIX;
    uniform mat4 V_MATRIX;
    uniform mat4 MV_MATRIX;
    uniform vec3 CAMERA_WORLD_POSITION;
    uniform mat4 SHADOW_MATRIX[MAX_LIGHTS];
    uniform mat4 BONES[MAX_BONES];
    uniform float HAS_BONES;
    in vec3 VERTEX_POSITION;
    in vec3 VERTEX_NORMAL;
    in vec3 VERTEX_TANGENT;
    in vec2 VERTEX_UV_0;
    in vec4 VERTEX_BONE_IDS;
    in vec4 VERTEX_BONE_WEIGHTS;

    out vec3 vs_normal_model;
    out vec3 vs_position_model;
    out vec3 vs_position_view;
    out vec3 vs_tangent;
    out vec2 vs_tex0_uv;
    out vec3 vs_camera_world;
    out vec4 vs_shadow_coord[4];

    ` + calcSkinnedData + `

    void main()
    {
    	skinnedData skinned;
    	if (HAS_BONES > 0.0) {
    		skinned = calculateSkinnedData();
    	} else {
    		skinned.position = vec4(VERTEX_POSITION, 1.0);
    		skinned.normal = VERTEX_NORMAL;
    		skinned.tangent = VERTEX_TANGENT;
    	}

    	mat3 vs_normal_mat = transpose(inverse(mat3(M_MATRIX)));

    	vs_normal_model = vs_normal_mat * skinned.normal;
    	vs_position_model = vec3(M_MATRIX * skinned.position);
    	vs_position_view = vec3(MV_MATRIX * skinned.position);
    	vs_camera_world = CAMERA_WORLD_POSITION;
    	vs_tangent = mat3(M_MATRIX) * skinned.tangent;
    	vs_tex0_uv = VERTEX_UV_0;

    	/* handle the shadow coordinates unrolled since for loop indexing can be problematic */
    	vs_shadow_coord[0] = (SHADOW_MATRIX[0] * M_MATRIX) * skinned.position;
    	vs_shadow_coord[1] = (SHADOW_MATRIX[1] * M_MATRIX) * skinned.position;
    	vs_shadow_coord[2] = (SHADOW_MATRIX[2] * M_MATRIX) * skinned.position;
    	vs_shadow_coord[3] = (SHADOW_MATRIX[3] * M_MATRIX) * skinned.position;

    	gl_Position = MVP_MATRIX * skinned.position;
    }
    `

	basicSkinnedShaderF = `#version 330
    precision highp float;

    const int MAX_LIGHTS=4;

    uniform mat4 V_MATRIX;
    uniform vec4 MATERIAL_DIFFUSE;
    uniform vec4 MATERIAL_SPECULAR;
    uniform float MATERIAL_SHININESS;
    uniform sampler2D MATERIAL_TEX_DIFFUSE;
    uniform sampler2D MATERIAL_TEX_NORMALS;
    uniform float MATERIAL_TEX_DIFFUSE_VALID;
    uniform float MATERIAL_TEX_NORMALS_VALID;
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
    in vec3 vs_camera_world;
    in vec4 vs_shadow_coord[4];

    out vec4 frag_color;

    ` + calcShadowFactor + `

    ` + calcADSLights + `

    void main()
    {
    	vec4 color = MATERIAL_DIFFUSE;
    	if (MATERIAL_TEX_DIFFUSE_VALID > 0.0) {
    		color *= texture(MATERIAL_TEX_DIFFUSE, vs_tex0_uv);
    	}

      	vec4 shadowFactor = CalcShadowFactor();

    	vec3 normal = vs_normal_model;
    	if (MATERIAL_TEX_NORMALS_VALID > 0.0) {
    		vec3 T = normalize(vs_tangent - dot(vs_tangent, vs_normal_model) * vs_normal_model);
    		vec3 BT = cross(T, vs_normal_model);
    		vec3 bump_normal = texture(MATERIAL_TEX_NORMALS, vs_tex0_uv).rgb;
    		bump_normal = 2.0 * bump_normal - vec3(1.0, 1.0, 1.0);
    		mat3 TBN = mat3(T, BT, vs_normal_model);
    		normal = TBN * bump_normal;
    	}

    	frag_color = vec4(shadowFactor.rgb * CalcADSLights(vs_position_model, normalize(normal), color.rgb), 1.0);
    }
    `

	/*

	    _____           _
	   / ____|         | |
	  | |        ___   | |   ___    _ __
	  | |       / _ \  | |  / _ \  | '__|
	  | |____  | (_) | | | | (_) | | |
	   \_____|  \___/  |_|  \___/  |_|

	*/

	colorShaderV = `#version 330
    precision highp float;

    uniform mat4 MVP_MATRIX;

    in vec3 VERTEX_POSITION;

    void main(void) {
    	gl_Position = MVP_MATRIX * vec4(VERTEX_POSITION, 1.0);
    }
    `

	colorShaderF = `#version 330
    precision highp float;

    uniform vec4 MATERIAL_DIFFUSE;

    out vec4 frag_color;

    void main (void) {
    	frag_color = MATERIAL_DIFFUSE;
    }
    `

	/*

	    _____           _                    _______                 _
	   / ____|         | |                  |__   __|               | |
	  | |        ___   | |   ___    _ __       | |      ___  __  __ | |_
	  | |       / _ \  | |  / _ \  | '__|      | |     / _ \ \ \/ / | __|
	  | |____  | (_) | | | | (_) | | |         | |    |  __/  >  <  | |_
	   \_____|  \___/  |_|  \___/  |_|         |_|     \___| /_/\_\  \__|

	*/

	colorTextShaderV = `#version 330
    precision highp float;

    uniform mat4 MVP_MATRIX;

    in vec3 VERTEX_POSITION;
    in vec2 VERTEX_UV_0;

    out vec2 vs_tex0_uv;

    void main(void) {
    	gl_Position = MVP_MATRIX * vec4(VERTEX_POSITION, 1.0);
    	vs_tex0_uv = VERTEX_UV_0;
    }
    `

	colorTextShaderF = `#version 330
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
    `

	/*
		     ______  _   __   __                     _   _         _  _  _
				|  _  \(_) / _| / _|                   | | | |       | |(_)| |
				| | | | _ | |_ | |_  _   _  ___   ___  | | | | _ __  | | _ | |_
				| | | || ||  _||  _|| | | |/ __| / _ \ | | | || '_ \ | || || __|
				| |/ / | || |  | |  | |_| |\__ \|  __/ | |_| || | | || || || |_
				|___/  |_||_|  |_|   \__,_||___/ \___|  \___/ |_| |_||_||_| \__|
	*/

	diffuseUnlitShaderV = `#version 330
			precision highp float;

			uniform mat4 MVP_MATRIX;

			in vec3 VERTEX_POSITION;
			in vec2 VERTEX_UV_0;

			out vec2 vs_tex0_uv;

			void main(void) {
				gl_Position = MVP_MATRIX * vec4(VERTEX_POSITION, 1.0);
				vs_tex0_uv = VERTEX_UV_0;
			}
			`

	diffuseUnlitShaderF = `#version 330
			precision highp float;

			uniform sampler2D MATERIAL_TEX_DIFFUSE;
			uniform vec4 MATERIAL_DIFFUSE;

			in vec2 vs_tex0_uv;
			out vec4 frag_color;

			void main (void) {
				vec4 texColor = texture(MATERIAL_TEX_DIFFUSE, vs_tex0_uv);
				frag_color = texColor * MATERIAL_DIFFUSE;
			}
			`

	/*
	   _____   _                   _                                                     _____
	   / ____| | |                 | |                                                   / ____|
	   | (___   | |__     __ _    __| |   ___   __      __  _ __ ___     __ _   _ __     | |  __    ___   _ __
	   \___ \  | '_ \   / _` |  / _` |  / _ \  \ \ /\ / / | '_ ` _ \   / _` | | '_ \    | | |_ |  / _ \ | '_ \
	   ____) | | | | | | (_| | | (_| | | (_) |  \ V  V /  | | | | | | | (_| | | |_) |   | |__| | |  __/ | | | |
	   |_____/  |_| |_|  \__,_|  \__,_|  \___/    \_/\_/   |_| |_| |_|  \__,_| | .__/     \_____|  \___| |_| |_|
	   																	  | |
	   																	  |_|
	*/

	shadowmapGeneratorV = `#version 330
	precision highp float;

	uniform mat4 M_MATRIX;
	uniform mat4 SHADOW_VP_MATRIX;
	in vec4 VERTEX_POSITION;

	/* shadow pass */
	void main() {
	  gl_Position = SHADOW_VP_MATRIX * M_MATRIX * VERTEX_POSITION;
	}
	`

	shadowmapGeneratorF = `#version 330
	precision highp float;

	out vec4 frag_color;

	void main (void) {
	  frag_color = vec4(gl_FragCoord.z);
	}
	`
)

// CreateBasicShader creates a new shader object using the built
// in basic shader code.
func CreateBasicShader() (*fizzle.RenderShader, error) {
	return fizzle.LoadShaderProgram(basicShaderV, basicShaderF, nil)
}

// CreateBasicSkinnedShader creates a new shader object using the built
// in basic shader code with GPU skinning for bones.
func CreateBasicSkinnedShader() (*fizzle.RenderShader, error) {
	return fizzle.LoadShaderProgram(basicSkinnedShaderV, basicSkinnedShaderF, nil)
}

// CreateColorShader creates a new shader object using the built
// in flat color shader code that uses Material.DiffuseColor.
func CreateColorShader() (*fizzle.RenderShader, error) {
	return fizzle.LoadShaderProgram(colorShaderV, colorShaderF, nil)
}

// CreateColorTextShader creates a new shader object using the built
// in flat color shader code that uses Material.DiffuseColor and is
// meant to be used to draw characters in a texture font.
func CreateColorTextShader() (*fizzle.RenderShader, error) {
	return fizzle.LoadShaderProgram(colorTextShaderV, colorTextShaderF, nil)
}

// CreateShadowmapGeneratorShader creates a new shader object using the built
// in shadowmap generator shader. This can be used to render objects for a
// shadow map texture to do dynamic shadows in a scene.
func CreateShadowmapGeneratorShader() (*fizzle.RenderShader, error) {
	return fizzle.LoadShaderProgram(shadowmapGeneratorV, shadowmapGeneratorF, nil)
}

// CreateDiffuseUnlitShader creates a new shader object using the built
// in diffuse texture shader that is unlit (no lighting calculated).
func CreateDiffuseUnlitShader() (*fizzle.RenderShader, error) {
	return fizzle.LoadShaderProgram(diffuseUnlitShaderV, diffuseUnlitShaderF, nil)
}
