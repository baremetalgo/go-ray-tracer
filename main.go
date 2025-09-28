package main

import (
	"fmt"
	"go-ray-tracing/core"
	"go-ray-tracing/link_server"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// Init window
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 600, "GoEngine :: GameView")
	defer rl.CloseWindow()

	customFont := rl.LoadFont("E:/GitHub/GoEngine/fonts/CALIBRIB.TTF")
	defer rl.UnloadFont(customFont)
	// font := customFont

	scene := core.NewScene3D()
	scene.InitScene()

	/*
		server := link_server.Start_Server(scene)
		defer server.Stop()
	*/
	rl.SetTargetFPS(60)
	rl.EnableCursor()

	for !rl.WindowShouldClose() {
		scene.UpdateScene()
		/*
			server.ProcessStream()
		*/

		scene.Renderer.RunPreRenderProcess(scene)
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)

		scene.Renderer.Render(scene)

		// drawStatusInfo(scene, server, font)

		rl.EndDrawing()
	}

	scene.Renderer.RunPostRenderProcess(scene)
}

func drawStatusInfo(scene *core.Scene3D, server *link_server.LiveLinkServer, font rl.Font) {
	// Draw FPS
	rl.DrawFPS(10, 10)

	// Status panel background
	panelWidth := int32(200)
	panelHeight := int32(150)
	rl.DrawRectangle(10, 40, panelWidth, panelHeight, rl.Fade(rl.Black, 0.2))
	rl.DrawRectangleLines(10, 40, panelWidth, panelHeight, rl.White)

	// Server status
	statusColor := rl.Green
	if link_server.SERVER_STATUS == "Off" {
		statusColor = rl.Red
	}
	rl.DrawTextEx(font, fmt.Sprintf("Live-Link Server: %s", link_server.SERVER_STATUS), rl.NewVector2(20, 50), 16, 0, statusColor)

	// Client connections
	clientCount := len(link_server.ClientConnections)
	clientStatus := "Connected"
	clientColor := rl.Green
	if clientCount == 0 {
		clientStatus = "No clients"
		clientColor = rl.Red
	}
	rl.DrawTextEx(font, fmt.Sprintf("Clients: %d (%s)", clientCount, clientStatus), rl.NewVector2(20, 80), 16, 0, clientColor)

	// Live-linked objects count
	liveObjects := 0
	var objectNames []string
	for _, geom := range scene.Geometries {
		// Count objects that have been updated (not just the initial sphere)
		if geom.Name != "pSphere1" && geom.Name != "Plane" {
			liveObjects++
			objectNames = append(objectNames, geom.Name)
		}
	}

	rl.DrawTextEx(font, fmt.Sprintf("Live-linked Objects: %d", liveObjects), rl.NewVector2(20, 110), 16, 0, rl.White)

	// Object names (scrollable if too many)
	if liveObjects > 0 {
		rl.DrawTextEx(font, "Objects:", rl.NewVector2(20, 140), 12, 0, rl.Yellow)

		maxDisplay := 5 // Maximum objects to display in the panel
		startY := int32(170)
		lineHeight := int32(20)

		for i, name := range objectNames {
			if i >= maxDisplay {
				// Show "and X more..." if there are too many objects
				remaining := liveObjects - maxDisplay
				rl.DrawTextEx(font, fmt.Sprintf("... and %d more", remaining), rl.NewVector2(12, float32(startY+int32(i)*lineHeight)), 16, 0, rl.Gray)
				break
			}
			// Truncate long names
			displayName := name
			if len(name) > 20 {
				displayName = name[:17] + "..."
			}
			rl.DrawTextEx(font, fmt.Sprintf("• %s", displayName), rl.NewVector2(30, float32(startY+int32(i)*lineHeight)), 16, 0, rl.LightGray)
		}
	} else {
		rl.DrawTextEx(font, "No live-linked objects", rl.NewVector2(20, 140), 16, 0, rl.White)
	}

	// Instructions at the bottom
	instructions := "Move: WASD | Look: Mouse | Zoom: Mouse Wheel"
	textWidth := rl.MeasureText(instructions, 16)
	rl.DrawTextEx(font, instructions, rl.NewVector2(float32(rl.GetScreenWidth())/2-float32(textWidth)/2, float32(rl.GetScreenHeight())-30), 16, 0, rl.LightGray)
}
