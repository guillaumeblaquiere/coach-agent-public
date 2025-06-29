package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("PROJECT_ID") // Automatically inferred on Cloud Run/GCP
	if projectID == "" {
		// For local development, you might set this or use emulator
		// log.Println("GOOGLE_CLOUD_PROJECT not set. Using default or emulator settings.")
		// projectID = "your-local-project-id" // Or connect to emulator
	}

	var fsClient *firestore.Client
	var err error

	// For local development with Firestore emulator:
	// if os.Getenv("FIRESTORE_EMULATOR_HOST") != "" {
	//  log.Println("Using Firestore Emulator")
	// 	fsClient, err = firestore.NewClient(ctx, projectID, option.WithEndpoint(os.Getenv("FIRESTORE_EMULATOR_HOST")))
	// } else {
	//  log.Println("Connecting to live Firestore")
	fsClient, err = firestore.NewClient(ctx, projectID)
	// }

	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer fsClient.Close()

	log.Println("Successfully connected to Firestore.")

	api := NewAPI(fsClient)
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // Chi's built-in logger
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second)) // Set a timeout

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Replace with your frontend's origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-User-Email"},
		AllowCredentials: true,
	})
	r.Use(corsMiddleware.Handler)

	// --- API Routes ---
	r.Route("/api/v1", func(r chi.Router) {
		// Categories
		//r.Post("/categories", api.CreateCategory)
		r.Get("/categories", api.ListCategories)
		r.Get("/categories/{categoryId}", api.GetCategory)
		//r.Put("/categories/{categoryId}", api.UpdateCategory)
		//r.Delete("/categories/{categoryId}", api.DeleteCategory)

		// Drills
		//r.Post("/drills", api.CreateDrill)
		r.Get("/drills", api.ListDrills) // Add query param ?categoryId=XYZ
		r.Get("/drills/{categoryId}", api.GetDrill)
		//r.Put("/drills/{drillId}", api.UpdateDrill)
		//r.Delete("/drills/{drillId}", api.DeleteDrill)

		// Training Plan Templates
		//r.Post("/plan-templates", api.CreatePlanTemplate)
		r.Get("/plan-templates", api.ListPlanTemplates)
		r.Get("/plan-templates/{templateId}", api.GetPlanTemplate)
		//r.Put("/plan-templates/{templateId}", api.UpdatePlanTemplate)
		//r.Delete("/plan-templates/{templateId}", api.DeletePlanTemplate)
		//r.Post("/plan-templates/{templateId}/set-default", api.SetDefaultPlanTemplate)

		// Daily Training Plans
		r.Post("/daily-plans/initiate", api.InitiateDailyPlan) // Body: { "date": "YYYY-MM-DD" (opt), "templateId": "XYZ" (opt) }
		r.Get("/daily-plans/{date}", api.GetDailyPlan)         // {date} can be "YYYY-MM-DD" or "today"
		r.Put("/daily-plans/today", api.UpdateTodayDailyPlan)  // Body: Full DailyTrainingPlan object with updates

		// WebSockets
		r.Get("/ws", api.handleWebSocket)
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// ConnectionManager manages active WebSocket connections.
type ConnectionManager struct {
	connections map[string][]*websocket.Conn
	mu          sync.Mutex
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string][]*websocket.Conn),
	}
}

// Add registers a new connection for a user.
func (cm *ConnectionManager) Add(userEmail string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.connections[userEmail]; !ok {
		cm.connections[userEmail] = make([]*websocket.Conn, 0)
	}
	cm.connections[userEmail] = append(cm.connections[userEmail], conn)
	log.Printf("WebSocket connection added for user: %s. Conn length is now %d \n", userEmail, len(cm.connections[userEmail]))
}

// Remove deletes a connection.
func (cm *ConnectionManager) Remove(userEmail string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.connections[userEmail]; ok {
		for i, c := range cm.connections[userEmail] {
			if c == conn {
				cm.connections[userEmail] = append(cm.connections[userEmail][:i], cm.connections[userEmail][i+1:]...)
				break
			}
		}
		log.Printf("WebSocket connection removed for user: %s", userEmail)
	}
}

// WebSocketMessage is the structure for our messages.
type WebSocketMessage struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
	Source string      `json:"source"`
}

// SendMessage sends a message to a specific user.
func (cm *ConnectionManager) SendMessage(userEmail string, message WebSocketMessage) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	connections, ok := cm.connections[userEmail]
	if !ok {
		log.Printf("No active WebSocket connection for user: %s", userEmail)
		return
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling WebSocket message for %s: %v", userEmail, err)
		return
	}

	// Send the message to all the connections for this user
	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			log.Printf("Error sending WebSocket message to %s: %v", userEmail, err)
			// We might want to delete the connection if writing fails.
			conn.Close()
			// Remove the connection from the list
			for i, c := range cm.connections[userEmail] {
				if c == conn {
					cm.connections[userEmail] = append(cm.connections[userEmail][:i], cm.connections[userEmail][i+1:]...)
					break
				}
			}

		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Put stricter logic here for production!
		return true
	},
}
