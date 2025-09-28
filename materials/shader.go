package materials

import rl "github.com/gen2brain/raylib-go/raylib"

const vertexShaderCode = `
#version 330

in vec3 vertexPosition;
in vec3 vertexNormal;

uniform mat4 mvp;
uniform mat4 matModel;
uniform mat4 lightSpaceMatrix;

out vec3 fragNormal;
out vec3 fragPos;
out vec4 fragPosLightSpace;

void main()
{
    vec4 worldPos = matModel * vec4(vertexPosition, 1.0);
    fragPos = worldPos.xyz;
    fragNormal = normalize(mat3(matModel) * vertexNormal);
    fragPosLightSpace = lightSpaceMatrix * worldPos;
    gl_Position = mvp * vec4(vertexPosition, 1.0);
}
`

const fragmentShaderCode = `
#version 330

in vec3 fragNormal;
in vec3 fragPos;
in vec4 fragPosLightSpace;

out vec4 finalColor;

uniform vec3 lightDir;
uniform vec3 objectColor;
uniform sampler2D shadowMap;
uniform vec3 viewPos;

float ShadowCalculation(vec4 fragPosLightSpace)
{
    // Perform perspective divide
    vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
    
    // Transform to [0,1] range
    projCoords = projCoords * 0.5 + 0.5;
    
    // Get closest depth value from light's perspective
    float closestDepth = texture(shadowMap, projCoords.xy).r;
    
    // Get depth of current fragment from light's perspective
    float currentDepth = projCoords.z;
    
    // Check if current fragment is in shadow
    float shadow = currentDepth > closestDepth + 0.001 ? 1.0 : 0.0;
    
    // Shadow edge smoothing with PCF
    float shadowSmooth = 0.0;
    vec2 texelSize = 1.0 / textureSize(shadowMap, 0);
    for(int x = -1; x <= 1; ++x)
    {
        for(int y = -1; y <= 1; ++y)
        {
            float pcfDepth = texture(shadowMap, projCoords.xy + vec2(x, y) * texelSize).r;
            shadowSmooth += currentDepth > pcfDepth + 0.001 ? 1.0 : 0.0;
        }
    }
    shadowSmooth /= 9.0;
    
    return shadowSmooth;
}

void main()
{
    vec3 norm = normalize(fragNormal);
    vec3 lightDirection = normalize(-lightDir);
    
    // Diffuse lighting
    float diff = max(dot(norm, lightDirection), 0.0);
    vec3 diffuse = diff * objectColor;
    
    // Ambient lighting
    vec3 ambient = 0.2 * objectColor;
    
    // Specular lighting
    vec3 viewDir = normalize(viewPos - fragPos);
    vec3 reflectDir = reflect(-lightDirection, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), 32.0);
    vec3 specular = spec * vec3(0.3);
    
    // Calculate shadow
    float shadow = ShadowCalculation(fragPosLightSpace);
    
    // Final color with shadows
    vec3 lighting = ambient + (1.0 - shadow) * (diffuse + specular);
    finalColor = vec4(lighting, 1.0);
}
`

const depthVertexShaderCode = `
#version 330

in vec3 vertexPosition;
uniform mat4 lightSpaceMatrix;
uniform mat4 matModel;

void main()
{
    vec4 worldPos = matModel * vec4(vertexPosition, 1.0);
    gl_Position = lightSpaceMatrix * worldPos;
}
`

const depthFragmentShaderCode = `
#version 330
out vec4 fragColor;

void main()
{
    // Just output depth, but we need to write something
    fragColor = vec4(1.0);
}
`

type Material struct {
	VertexShader     string
	FragmentShader   string
	Shader           rl.Shader
	DepthShader      rl.Shader // Add this
	ShadowMap        rl.RenderTexture2D
	LightSpaceMatrix rl.Matrix
}

// Update NewMaterial function
func NewMaterial() *Material {
	shader := Material{}
	shader.VertexShader = vertexShaderCode
	shader.FragmentShader = fragmentShaderCode
	shader.Shader = rl.LoadShaderFromMemory(shader.VertexShader, shader.FragmentShader)

	// Load depth shader
	shader.DepthShader = rl.LoadShaderFromMemory(depthVertexShaderCode, depthFragmentShaderCode)

	shader.ShadowMap = rl.LoadRenderTexture(1024, 1024)

	// Initial light space matrix
	lightProjection := rl.MatrixOrtho(-10.0, 10.0, -10.0, 10.0, 1.0, 50.0)
	lightView := rl.MatrixLookAt(
		rl.NewVector3(-2.0, 4.0, -1.0),
		rl.NewVector3(0.0, 0.0, 0.0),
		rl.NewVector3(0.0, 1.0, 0.0),
	)
	shader.LightSpaceMatrix = rl.MatrixMultiply(lightView, lightProjection)

	return &shader
}

func (m *Material) UpdateLightUniforms(cameraPos rl.Vector3) {
	// Set light direction (sun direction - pointing downward)
	lightDir := []float32{-0.5, -1.0, -0.5}
	lightDirLoc := rl.GetShaderLocation(m.Shader, "lightDir")
	rl.SetShaderValue(m.Shader, lightDirLoc, lightDir, rl.ShaderUniformVec3)

	// Set view position for specular calculations
	viewPos := []float32{cameraPos.X, cameraPos.Y, cameraPos.Z}
	viewPosLoc := rl.GetShaderLocation(m.Shader, "viewPos")
	rl.SetShaderValue(m.Shader, viewPosLoc, viewPos, rl.ShaderUniformVec3)

	// Set light space matrix for shadow mapping
	lightSpaceLoc := rl.GetShaderLocation(m.Shader, "lightSpaceMatrix")
	rl.SetShaderValueMatrix(m.Shader, lightSpaceLoc, m.LightSpaceMatrix)
}

func (m *Material) UpdateLightCamera(lightPos, lightTarget rl.Vector3) {
	lightProjection := rl.MatrixOrtho(-10.0, 10.0, -10.0, 10.0, 1.0, 50.0)
	lightView := rl.MatrixLookAt(lightPos, lightTarget, rl.NewVector3(0.0, 1.0, 0.0))
	m.LightSpaceMatrix = rl.MatrixMultiply(lightView, lightProjection)
}
