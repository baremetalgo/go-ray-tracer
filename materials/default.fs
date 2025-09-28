#version 330

in vec3 fragNormal;
in vec3 fragPos;

out vec4 finalColor;

uniform vec3 lightDir;
uniform vec3 objectColor;

void main()
{
    vec3 norm = normalize(fragNormal);
    float diff = max(dot(norm, normalize(-lightDir)), 0.0);

    vec3 diffuse = diff * objectColor;
    vec3 ambient = 0.2 * objectColor;

    finalColor = vec4(diffuse + ambient, 1.0);
}
