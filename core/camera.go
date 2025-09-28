package core

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type PerspectiveCamera struct {
	Camera rl.Camera3D
}

func NewPerspectiveCamera() *PerspectiveCamera {
	camera_3d := PerspectiveCamera{}
	camera := rl.Camera3D{}
	camera.Position = rl.NewVector3(4.0, 4.0, 4.0)
	camera.Target = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	camera_3d.Camera = camera

	return &camera_3d
}

func UpdateCameraManually(cam *rl.Camera3D, speed float32, rotSpeed float32) {
	dt := rl.GetFrameTime()

	// Movement
	if rl.IsKeyDown(rl.KeyUp) {
		cam.Position.Z -= speed * dt
		cam.Target.Z -= speed * dt
	}
	if rl.IsKeyDown(rl.KeyDown) {
		cam.Position.Z += speed * dt
		cam.Target.Z += speed * dt
	}
	if rl.IsKeyDown(rl.KeyLeft) {
		cam.Position.X -= speed * dt
		cam.Target.X -= speed * dt
	}
	if rl.IsKeyDown(rl.KeyRight) {
		cam.Position.X += speed * dt
		cam.Target.X += speed * dt
	}

	// Optional: rotate with A/D or Q/E
	if rl.IsKeyDown(rl.KeyA) {
		angle := rotSpeed * dt
		dx := cam.Target.X - cam.Position.X
		dz := cam.Target.Z - cam.Position.Z
		cam.Target.X = cam.Position.X + float32(math.Cos(float64(angle)))*dx - float32(math.Sin(float64(angle)))*dz
		cam.Target.Z = cam.Position.Z + float32(math.Sin(float64(angle)))*dx + float32(math.Cos(float64(angle)))*dz
	}
	if rl.IsKeyDown(rl.KeyD) {
		angle := -rotSpeed * dt
		dx := cam.Target.X - cam.Position.X
		dz := cam.Target.Z - cam.Position.Z
		cam.Target.X = cam.Position.X + float32(math.Cos(float64(angle)))*dx - float32(math.Sin(float64(angle)))*dz
		cam.Target.Z = cam.Position.Z + float32(math.Sin(float64(angle)))*dx + float32(math.Cos(float64(angle)))*dz
	}
}
