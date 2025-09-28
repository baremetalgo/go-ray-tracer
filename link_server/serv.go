package link_server

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"go-ray-tracing/core"
	"net"
	"strings"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var lastMeshUpdate = make(map[string]time.Time)
var meshMutex sync.Mutex
var ClientConnections []net.Conn
var connectionsMutex sync.Mutex
var SERVER_STATUS = "On"

type TransformData struct {
	Name     string     `json:"name"`
	Position [3]float32 `json:"position"`
	Rotation [4]float32 `json:"rotation"` // Quaternion (x, y, z, w)
	Scale    [3]float32 `json:"scale"`
}

type MeshData struct {
	Name      string       `json:"name"`
	Vertices  [][3]float32 `json:"vertices"`
	Normals   [][3]float32 `json:"normals"`
	TexCoords [][2]float32 `json:"texCoords"`
	Indices   []int32      `json:"indices"`
}

type LiveLinkServer struct {
	host    string
	port    string
	clients map[net.Conn]bool
	mutex   sync.Mutex
	scene   *core.Scene3D

	// Channel for communicating mesh data to main thread
	meshChan chan MeshData
	// Channel for communicating transform data to main thread
	transformChan chan TransformData
}

func NewLiveLinkServer(host, port string, scene *core.Scene3D) *LiveLinkServer {
	return &LiveLinkServer{
		host:          host,
		port:          port,
		clients:       make(map[net.Conn]bool),
		scene:         scene,
		meshChan:      make(chan MeshData, 100),
		transformChan: make(chan TransformData, 100),
	}
}

func Start_Server(scene *core.Scene3D) *LiveLinkServer {
	// Start Live Link server
	liveLinkServer := NewLiveLinkServer("localhost", "8080", scene)
	go func() {
		if err := liveLinkServer.Start(); err != nil {
			SERVER_STATUS = "Off"
			rl.TraceLog(rl.LogError, "Failed to start server: %v", err)
		}
	}()

	return liveLinkServer
}

func (s *LiveLinkServer) Start() error {
	listener, err := net.Listen("tcp", s.host+":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("✅ Live Link server started on %s:%s\n", s.host, s.port)
	fmt.Printf("✅ Server is listening for connections...\n")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		fmt.Printf("✅ New client connected: %s\n", conn.RemoteAddr().String())
		s.mutex.Lock()
		s.clients[conn] = true
		s.mutex.Unlock()

		// Add to client connections list
		connectionsMutex.Lock()
		ClientConnections = append(ClientConnections, conn)
		connectionsMutex.Unlock()

		go s.handleClient(conn)
	}
}

func (s *LiveLinkServer) ProcessStream() {
	s.ProcessMeshUpdates()
	s.ProcessTransformUpdates()
}

func (s *LiveLinkServer) ProcessMeshUpdates() {
	select {
	case meshData, ok := <-s.meshChan:
		if ok {
			s.handleMeshDataMainThread(meshData)
		} else {
			fmt.Println("Mesh channel closed")
		}
	default:
		// No mesh data to process
	}
}

func (s *LiveLinkServer) ProcessTransformUpdates() {
	select {
	case transformData, ok := <-s.transformChan:
		if ok {
			s.handleTransformDataMainThread(transformData)
		} else {
			fmt.Println("Transform channel closed")
		}
	default:
		// No transform data to process
	}
}

func (s *LiveLinkServer) handleClient(conn net.Conn) {
	defer func() {
		s.mutex.Lock()
		delete(s.clients, conn)
		s.mutex.Unlock()

		// Remove from client connections list
		connectionsMutex.Lock()
		for i, clientConn := range ClientConnections {
			if clientConn == conn {
				ClientConnections = append(ClientConnections[:i], ClientConnections[i+1:]...)
				break
			}
		}
		connectionsMutex.Unlock()

		conn.Close()
		fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
	}()

	fmt.Printf("Client connected: %s\n", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)

	for {
		// Read message length (4 bytes)
		lengthBytes := make([]byte, 4)
		n, err := reader.Read(lengthBytes)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("Client %s disconnected gracefully\n", conn.RemoteAddr().String())
				return // Normal disconnect
			} else if strings.Contains(err.Error(), "forcibly closed") {
				fmt.Printf("Client %s closed the connection\n", conn.RemoteAddr().String())
				return // Client closed connection
			} else {
				fmt.Printf("Error reading message length from %s: %v\n", conn.RemoteAddr().String(), err)
				return
			}
		}

		if n != 4 {
			fmt.Printf("Incomplete length data from %s: expected 4 bytes, got %d\n", conn.RemoteAddr().String(), n)
			return
		}

		length := binary.BigEndian.Uint32(lengthBytes)

		// Read message data
		data := make([]byte, length)
		bytesRead := 0
		for bytesRead < int(length) {
			n, err := reader.Read(data[bytesRead:])
			if err != nil {
				fmt.Printf("Error reading message data from %s: %v\n", conn.RemoteAddr().String(), err)
				return
			}
			bytesRead += n
		}

		// Parse message type (first 4 bytes)
		if len(data) < 4 {
			fmt.Printf("Message too short from %s: expected at least 4 bytes, got %d\n", conn.RemoteAddr().String(), len(data))
			continue
		}

		messageType := string(data[:4])
		messageData := data[4:]

		fmt.Printf("Received message type: %s, length: %d\n", messageType, len(messageData))

		switch messageType {
		case "TRNS": // Transform data
			var transformData TransformData
			if err := json.Unmarshal(messageData, &transformData); err != nil {
				fmt.Printf("Error parsing transform data from %s: %v\n", conn.RemoteAddr().String(), err)
				continue
			}
			fmt.Printf("Received transform for: %s\n", transformData.Name)
			// Send to main thread via channel
			s.transformChan <- transformData

		case "MESH": // Mesh data
			var meshData MeshData
			if err := json.Unmarshal(messageData, &meshData); err != nil {
				fmt.Printf("Error parsing mesh data from %s: %v\n", conn.RemoteAddr().String(), err)
				continue
			}
			fmt.Printf("Received mesh for: %s, vertices: %d\n", meshData.Name, len(meshData.Vertices))
			// Send to main thread via channel
			s.meshChan <- meshData

		case "PING": // Ping message for testing
			fmt.Printf("Ping received from %s\n", conn.RemoteAddr().String())
			response := []byte("PONG")
			s.sendMessage(conn, "PONG", response)

		default:
			fmt.Printf("Unknown message type from %s: %s\n", conn.RemoteAddr().String(), messageType)
		}
	}
}

func (s *LiveLinkServer) sendMessage(conn net.Conn, messageType string, data []byte) {
	message := make([]byte, 4+len(data))
	copy(message[:4], []byte(messageType))
	copy(message[4:], data)

	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(message)))

	// Send message length
	if _, err := conn.Write(lengthBytes); err != nil {
		fmt.Printf("Error sending message length: %v\n", err)
		return
	}

	// Send message
	if _, err := conn.Write(message); err != nil {
		fmt.Printf("Error sending message: %v\n", err)
		return
	}
}

// In server/live_link_server.go, update the handleTransformData function:
func (s *LiveLinkServer) handleTransformDataMainThread(data TransformData) {

	// Rate limiting - don't process the same mesh too frequently
	meshMutex.Lock()
	lastUpdate, exists := lastMeshUpdate[data.Name]
	if exists && time.Since(lastUpdate) < 100*time.Millisecond {
		meshMutex.Unlock()
		fmt.Printf("Skipping rapid mesh update for: %s\n", data.Name)
		return
	}
	lastMeshUpdate[data.Name] = time.Now()
	meshMutex.Unlock()

	fmt.Printf("Received transform for: %s - Pos: %.2f,%.2f,%.2f - Rot: %.2f,%.2f,%.2f,%.2f - Scale: %.2f,%.2f,%.2f\n",
		data.Name,
		data.Position[0], data.Position[1], data.Position[2],
		data.Rotation[0], data.Rotation[1], data.Rotation[2], data.Rotation[3],
		data.Scale[0], data.Scale[1], data.Scale[2])

	// Find the geometry with the matching name
	found := false
	for _, geom := range s.scene.Geometries {
		if geom.Name == data.Name {
			found = true
			// Update position
			geom.Position = rl.NewVector3(
				data.Position[0],
				data.Position[1],
				data.Position[2],
			)

			// Update rotation (quaternion)
			geom.SetQuaternion(
				data.Rotation[0],
				data.Rotation[1],
				data.Rotation[2],
				data.Rotation[3],
			)

			// Update scale
			geom.Scale = rl.NewVector3(
				data.Scale[0],
				data.Scale[1],
				data.Scale[2],
			)
			fmt.Printf("Updated geometry %v: Position: %v, Rotation: %v, Scale %v\n", geom.Name, data.Position, data.Rotation, data.Scale)

			break
		}
	}

	if !found {
		fmt.Printf("Warning: Geometry not found with name: %s. Available geometries:\n", data.Name)
		for _, geom := range s.scene.Geometries {
			fmt.Printf("  - %s\n", geom.Name)
		}
	}
}

func (s *LiveLinkServer) handleMeshDataMainThread(data MeshData) {
	// Add a safety check at the beginning
	if s.scene == nil {
		fmt.Printf("Warning: Scene is nil, cannot process mesh data for %s\n", data.Name)
		return
	}

	fmt.Printf("Received mesh data for: %s - Vertices: %d, Indices: %d\n",
		data.Name, len(data.Vertices), len(data.Indices))

	// Validate mesh data
	if len(data.Vertices) == 0 {
		fmt.Printf("Warning: No vertices in mesh data for %s\n", data.Name)
		return
	}

	if len(data.Indices) == 0 || len(data.Indices)%3 != 0 {
		fmt.Printf("Warning: Invalid indices count for %s: %d (must be multiple of 3)\n",
			data.Name, len(data.Indices))
		return
	}

	// Convert the received data to scene.GeoData format
	meshData := core.GeoData{
		Vertices:  make([]rl.Vector3, len(data.Vertices)),
		Normals:   make([]rl.Vector3, len(data.Normals)),
		TexCoords: make([]rl.Vector2, len(data.TexCoords)),
		Indices:   data.Indices,
	}

	// Convert vertices
	for i, v := range data.Vertices {
		meshData.Vertices[i] = rl.NewVector3(v[0], v[1], v[2])
	}

	// Convert normals (if provided, otherwise generate)
	if len(data.Normals) > 0 && len(data.Normals) == len(data.Vertices) {
		for i, n := range data.Normals {
			meshData.Normals[i] = rl.NewVector3(n[0], n[1], n[2])
		}
	} else {
		// Generate simple normals if not provided
		meshData.Normals = make([]rl.Vector3, len(data.Vertices))
		for i := range meshData.Normals {
			meshData.Normals[i] = rl.NewVector3(0, 1, 0) // Default up normal
		}
	}

	// Convert texture coordinates (if provided)
	if len(data.TexCoords) > 0 && len(data.TexCoords) == len(data.Vertices) {
		for i, t := range data.TexCoords {
			meshData.TexCoords[i] = rl.NewVector2(t[0], t[1])
		}
	} else {
		// Generate default UVs if not provided
		meshData.TexCoords = make([]rl.Vector2, len(data.Vertices))
		for i := range meshData.TexCoords {
			meshData.TexCoords[i] = rl.NewVector2(0, 0)
		}
	}

	// Create or update the geometry
	geom := s.findOrCreateGeometry(data.Name)
	if geom != nil {
		// Update the geometry with the new mesh data
		core.UpdateGeometryFromMeshData(geom, &meshData)

		// Add safety checks before accessing the model
		if s.scene.DefaultShader != nil && geom.Model.MeshCount > 0 {
			geom.Model.Materials.Shader = *s.scene.DefaultShader
		}
		fmt.Printf("Mesh updated successfully for: %s\n", data.Name)
	}
}

func (s *LiveLinkServer) findOrCreateGeometry(name string) *core.Geometry {
	// Try to find existing geometry
	for _, geom := range s.scene.Geometries {
		if geom.Name == name {
			return geom
		}
	}

	// Create new geometry if not found - use a simple placeholder for now
	// The actual mesh will be updated in the next step
	mesh := rl.GenMeshCube(1, 1, 1) // Default mesh
	model := rl.LoadModelFromMesh(mesh)
	geom := core.NewGeometry(&model, name)
	(geom.Model.Materials).Shader = *s.scene.DefaultShader

	s.scene.Geometries = append(s.scene.Geometries, geom)
	return geom
}

func (s *LiveLinkServer) BroadcastMessage(messageType string, data []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	message := make([]byte, 4+len(data))
	copy(message[:4], []byte(messageType))
	copy(message[4:], data)

	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(message)))

	for conn := range s.clients {
		// Send message length
		conn.Write(lengthBytes)

		// Send message
		conn.Write(message)
	}
}

func (s *LiveLinkServer) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for conn := range s.clients {
		conn.Close()
	}

	s.clients = make(map[net.Conn]bool)
	// close(s.meshChan)
	// close(s.transformChan)
}
