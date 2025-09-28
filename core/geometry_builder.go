// geometry_builder.go
package core

import (
	"fmt"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func CreateMeshFromData(data *GeoData) rl.Mesh {
	mesh := rl.Mesh{}

	// Create mesh with proper allocation
	mesh = rl.GenMeshPoly(0, 0) // Create an empty mesh as base
	rl.UnloadMesh(&mesh)        // Unload the empty mesh but keep the struct

	// Set counts
	mesh.VertexCount = int32(len(data.Vertices))
	mesh.TriangleCount = int32(len(data.Indices) / 3)

	// Allocate memory using raylib's internal functions
	if len(data.Vertices) > 0 {
		verts := make([]float32, len(data.Vertices)*3)
		for i, v := range data.Vertices {
			verts[i*3] = v.X
			verts[i*3+1] = v.Y
			verts[i*3+2] = v.Z
		}
		// Use raylib's mesh allocation function
		mesh.Vertices = (*float32)(rl.MemAlloc(uint32(uintptr(len(verts)) * unsafe.Sizeof(float32(0)))))
		copy(unsafe.Slice(mesh.Vertices, len(verts)), verts)
	}

	// Normals
	if len(data.Normals) > 0 {
		norms := make([]float32, len(data.Normals)*3)
		for i, n := range data.Normals {
			norms[i*3] = n.X
			norms[i*3+1] = n.Y
			norms[i*3+2] = n.Z
		}
		mesh.Normals = (*float32)(rl.MemAlloc(uint32(uintptr(len(norms)) * unsafe.Sizeof(float32(0)))))
		copy(unsafe.Slice(mesh.Normals, len(norms)), norms)
	}

	// Texcoords
	if len(data.TexCoords) > 0 {
		tex := make([]float32, len(data.TexCoords)*2)
		for i, t := range data.TexCoords {
			tex[i*2] = t.X
			tex[i*2+1] = t.Y
		}
		mesh.Texcoords = (*float32)(rl.MemAlloc(uint32(uintptr(len(tex)) * unsafe.Sizeof(float32(0)))))
		copy(unsafe.Slice(mesh.Texcoords, len(tex)), tex)
	}

	// Indices
	if len(data.Indices) > 0 {
		inds := make([]uint16, len(data.Indices))
		for i, idx := range data.Indices {
			inds[i] = uint16(idx)
		}
		mesh.Indices = (*uint16)(rl.MemAlloc(uint32(uintptr(len(inds)) * unsafe.Sizeof(uint16(0)))))
		copy(unsafe.Slice(mesh.Indices, len(inds)), inds)
	}

	// Upload mesh data to GPU
	rl.UploadMesh(&mesh, false)
	fmt.Printf("New Mesh Created : VertexCount=%d, TriangleCount=%d\n", mesh.VertexCount, mesh.TriangleCount)
	return mesh
}

// CreateModelFromMeshData creates a Geometry wrapper from GeoData
func CreateModelFromMeshData(data *GeoData, name string) *Geometry {
	mesh := CreateMeshFromData(data)
	model := rl.LoadModelFromMesh(mesh)
	fmt.Printf("New Model Created : %v\n", name)
	return NewGeometry(&model, name)
}

func UpdateGeometryFromMeshData(geom *Geometry, data *GeoData) {
	// Free the old model and mesh memory
	geom.Cleanup()

	// Upload new data
	mesh := CreateMeshFromData(data)
	model := rl.LoadModelFromMesh(mesh)

	// Update geometry
	geom.Model = model
	fmt.Printf("Updated Geometry with received mesh data : %v\n", geom.Name)
}
