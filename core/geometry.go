package core

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type MeshData struct {
	Vertices  []rl.Vector3
	Normals   []rl.Vector3
	TexCoords []rl.Vector2
	Indices   []int32
}

type Geometry struct {
	Name          string
	Model         rl.Model
	Position      rl.Vector3
	Rotation      rl.Vector3 // Rotation angles in degrees (pitch, yaw, roll)
	Quaternion    rl.Vector4 // Quaternion rotation (x, y, z, w)
	Scale         rl.Vector3
	Axis          rl.Vector3 // Rotation axis (usually 0, 1, 0 for Y-up)
	UseQuaternion bool       // Flag to determine which rotation to use
	Visibility    bool
}

func NewGeometry(model *rl.Model, name string) *Geometry {
	geom := Geometry{
		Name:          name,
		Model:         *model,
		Position:      rl.NewVector3(0, 0, 0),
		Rotation:      rl.NewVector3(0, 0, 0),
		Quaternion:    rl.NewVector4(0, 0, 0, 1), // Identity quaternion
		Scale:         rl.NewVector3(1, 1, 1),
		Axis:          rl.NewVector3(0, 1, 0),
		UseQuaternion: false,
		Visibility:    true,
	}
	return &geom
}

func NewSphereGeometry() *Geometry {
	sphere := rl.GenMeshSphere(1.0, 20, 20)
	geom := Geometry{
		Name:          "pSphere1",
		Model:         rl.LoadModelFromMesh(sphere),
		Position:      rl.NewVector3(0, 0.8, 0),
		Rotation:      rl.NewVector3(0, 0, 0),
		Quaternion:    rl.NewVector4(0, 0, 0, 1),
		Scale:         rl.NewVector3(1, 1, 1),
		Axis:          rl.NewVector3(0, 1, 0),
		UseQuaternion: false,
		Visibility:    true,
	}
	return &geom
}

func NewPlaneGeometry() *Geometry {
	plane := rl.GenMeshPlane(10, 10, 10, 10)
	geom := Geometry{
		Name:          "Plane",
		Model:         rl.LoadModelFromMesh(plane),
		Position:      rl.NewVector3(0, 0, 0),
		Rotation:      rl.NewVector3(0, 0, 0),
		Quaternion:    rl.NewVector4(0, 0, 0, 1),
		Scale:         rl.NewVector3(1, 1, 1),
		Axis:          rl.NewVector3(0, 1, 0),
		UseQuaternion: false,
		Visibility:    true,
	}
	return &geom
}

// Add methods to manipulate the geometry
func (g *Geometry) SetPosition(x, y, z float32) {
	g.Position = rl.NewVector3(x, y, z)
}

func (g *Geometry) SetRotation(pitch, yaw, roll float32) {
	g.Rotation = rl.NewVector3(pitch, yaw, roll)
	g.UseQuaternion = false
}

func (g *Geometry) SetScale(x, y, z float32) {
	g.Scale = rl.NewVector3(x, y, z)
}

func (g *Geometry) SetAxis(x, y, z float32) {
	g.Axis = rl.NewVector3(x, y, z)
}

// Add quaternion methods
func (g *Geometry) SetQuaternion(x, y, z, w float32) {
	g.Quaternion = rl.NewVector4(x, y, z, w)
	g.UseQuaternion = true
}

func (g *Geometry) SetRotationFromQuaternion(q rl.Vector4) {
	g.Quaternion = q
	g.UseQuaternion = true
}

// Convert quaternion to Euler angles (if needed)
func (g *Geometry) QuaternionToEuler() rl.Vector3 {
	if !g.UseQuaternion {
		return g.Rotation
	}

	// Quaternion to Euler conversion implementation
	x, y, z, w := float64(g.Quaternion.X), float64(g.Quaternion.Y), float64(g.Quaternion.Z), float64(g.Quaternion.W)

	// Roll (x-axis rotation)
	sinr_cosp := 2 * (w*x + y*z)
	cosr_cosp := 1 - 2*(x*x+y*y)
	roll := float32(math.Atan2(sinr_cosp, cosr_cosp)) * rl.Rad2deg

	// Pitch (y-axis rotation)
	sinp := 2 * (w*y - z*x)
	var pitch float32
	if math.Abs(sinp) >= 1 {
		pitch = float32(math.Copysign(math.Pi/2, sinp)) * rl.Rad2deg // Use 90 degrees if out of range
	} else {
		pitch = float32(math.Asin(sinp)) * rl.Rad2deg
	}

	// Yaw (z-axis rotation)
	siny_cosp := 2 * (w*z + x*y)
	cosy_cosp := 1 - 2*(y*y+z*z)
	yaw := float32(math.Atan2(siny_cosp, cosy_cosp)) * rl.Rad2deg

	return rl.NewVector3(pitch, yaw, roll)
}

// Rotate by adding to current rotation
func (g *Geometry) Rotate(pitch, yaw, roll float32) {
	g.Rotation.X += pitch
	g.Rotation.Y += yaw
	g.Rotation.Z += roll
	g.UseQuaternion = false
}

func (g *Geometry) Cleanup() {
	// Only unload if we have a valid model with meshes
	if g.Model.MeshCount > 0 {
		rl.UnloadModel(g.Model)
		// Reset the model to avoid double-free
		g.Model = rl.Model{}
	}
}

func (geom *Geometry) Draw() {
	if geom.Visibility {
		if geom.UseQuaternion {
			// Convert quaternion to axis-angle for drawing
			axis, angle := QuaternionToAxisAngle(geom.Quaternion)
			rl.DrawModelEx(
				geom.Model,
				geom.Position,
				axis,
				angle,
				geom.Scale,
				rl.White,
			)
			// Debug: Draw position marker
			rl.DrawSphere(geom.Position, 0.1, rl.Red)
		} else {
			// Use Euler angles as before
			rl.DrawModelEx(
				geom.Model,
				geom.Position,
				geom.Axis,
				geom.Rotation.Y,
				geom.Scale,
				rl.White,
			)
		}
	}
}

func QuaternionToAxisAngle(q rl.Vector4) (rl.Vector3, float32) {
	// Normalize the quaternion
	length := float32(math.Sqrt(float64(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)))
	if length == 0 {
		return rl.NewVector3(0, 1, 0), 0
	}

	x, y, z, w := q.X/length, q.Y/length, q.Z/length, q.W/length

	angle := 2 * float32(math.Acos(float64(w)))
	if angle < 0.001 {
		return rl.NewVector3(1, 0, 0), 0
	}

	s := float32(math.Sin(float64(angle / 2)))
	if s < 0.001 {
		return rl.NewVector3(1, 0, 0), angle * rl.Rad2deg
	}

	return rl.NewVector3(
		x/s,
		y/s,
		z/s,
	), angle * rl.Rad2deg
}
