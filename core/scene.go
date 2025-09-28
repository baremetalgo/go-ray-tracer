package core

import (
	"go-ray-tracing/materials"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GeoData struct {
	Vertices  []rl.Vector3
	Normals   []rl.Vector3
	TexCoords []rl.Vector2
	Indices   []int32
}

type Scene3D struct {
	Camera        *PerspectiveCamera
	Geometries    []*Geometry
	Material      *materials.Material
	DefaultShader *rl.Shader
	LightCamera   rl.Camera
	Renderer      Renderer3D
}

func NewScene3D() *Scene3D {
	scene := Scene3D{}
	scene.Camera = NewPerspectiveCamera()
	scene.Material = materials.NewMaterial()
	scene.DefaultShader = &scene.Material.Shader
	scene.Geometries = make([]*Geometry, 0)
	scene.LightCamera = rl.Camera3D{}
	scene.Renderer = *NewRenderer()
	return &scene
}

func (s *Scene3D) InitScene() {
	// creating a plane
	plane := NewPlaneGeometry()
	s.Geometries = append(s.Geometries, plane)
	(plane.Model.Materials).Shader = *s.DefaultShader

	// creating a sphere
	sp := NewSphereGeometry()
	s.Geometries = append(s.Geometries, sp)
	(sp.Model.Materials).Shader = *s.DefaultShader

	// Create light camera for shadow mapping

	s.LightCamera.Position = rl.NewVector3(-2.0, 4.0, -1.0) // Sun position
	s.LightCamera.Target = rl.NewVector3(0.0, 0.0, 0.0)     // Looking at origin
	s.LightCamera.Up = rl.NewVector3(0.0, 1.0, 0.0)         // Up vector
	s.LightCamera.Projection = rl.CameraOrthographic
}

func (s *Scene3D) UpdateScene() {
	rl.UpdateCamera(&s.Camera.Camera, rl.CameraMode(rl.CameraFirstPerson))

}

func (s *Scene3D) AddGeometry(model *rl.Model, name string) {
	new_geom := NewGeometry(model, name)
	(new_geom.Model.Materials).Shader = *s.DefaultShader
	s.Geometries = append(s.Geometries, new_geom)

}
