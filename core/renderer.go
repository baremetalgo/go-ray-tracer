package core

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Renderer3D struct {
}

func NewRenderer() *Renderer3D {
	return &Renderer3D{}
}

func (r *Renderer3D) CalculateLighting(scene *Scene3D) {
	// Update light uniforms with camera position
	scene.Material.UpdateLightUniforms(scene.Camera.Camera.Position)

	// Get shader uniform locations
	objColorLoc := rl.GetShaderLocation(*scene.DefaultShader, "objectColor")

	// Set object color (you might want to make this per-object)
	rl.SetShaderValue(*scene.DefaultShader, objColorLoc, []float32{0.7, 0.7, 0.7}, rl.ShaderUniformVec3)

	// Bind shadow map texture
	shadowMapLoc := rl.GetShaderLocation(*scene.DefaultShader, "shadowMap")
	rl.SetShaderValueTexture(*scene.DefaultShader, shadowMapLoc, scene.Material.ShadowMap.Texture)
}

func (r *Renderer3D) RunPreRenderProcess(scene *Scene3D) {
	// UPDATE LIGHT CAMERA MATRIX
	scene.Material.UpdateLightCamera(scene.LightCamera.Position, scene.LightCamera.Target)

	// PASS 1: Render shadow map from light's perspective
	scene.Renderer.RenderShadowMap(scene)
}

func (r *Renderer3D) RunPostRenderProcess(scene *Scene3D) {
	rl.UnloadShader(*scene.DefaultShader)
	for _, geom := range scene.Geometries {
		geom.Cleanup()
	}
}

func (r *Renderer3D) Render(scene *Scene3D) {
	rl.BeginMode3D(scene.Camera.Camera)

	scene.Renderer.CalculateLighting(scene)
	rl.DrawGrid(20, 10.0)

	for _, geom := range scene.Geometries {
		geom.Draw()
	}
	rl.EndMode3D()

}

func (r *Renderer3D) RenderShadowMap(scene *Scene3D) {
	rl.BeginTextureMode(scene.Material.ShadowMap)
	rl.ClearBackground(rl.White) // Clear with white (far depth)

	// Use the depth shader for shadow mapping
	rl.BeginShaderMode(scene.Material.DepthShader)

	// Set up light space matrix for depth shader
	lightSpaceLoc := rl.GetShaderLocation(scene.Material.DepthShader, "lightSpaceMatrix")
	rl.SetShaderValueMatrix(scene.Material.DepthShader, lightSpaceLoc, scene.Material.LightSpaceMatrix)
	// Simple rendering for shadow map - just draw the models
	for _, geom := range scene.Geometries {
		if geom.Visibility {
			// Set model matrix for depth shader
			modelMatrix := rl.MatrixIdentity()
			modelMatrix = rl.MatrixMultiply(modelMatrix, rl.MatrixTranslate(geom.Position.X, geom.Position.Y, geom.Position.Z))
			modelMatrix = rl.MatrixMultiply(modelMatrix, rl.MatrixScale(geom.Scale.X, geom.Scale.Y, geom.Scale.Z))

			modelLoc := rl.GetShaderLocation(scene.Material.DepthShader, "matModel")
			rl.SetShaderValueMatrix(scene.Material.DepthShader, modelLoc, modelMatrix)

			rl.DrawModel(geom.Model, rl.Vector3Zero(), 1.0, rl.White)
		}
	}
	rl.EndShaderMode()
	rl.EndTextureMode()

}
