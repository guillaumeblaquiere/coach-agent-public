package main

import (
	"context"
	"gblaquiere.dev/wrapper_agent/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	CoachAgentPortEnvVar  = "COACH_AGENT_PORT"
	CoachAgentHostEnvVar  = "COACH_AGENT_HOST"
	CoachAgentNameEnvVar  = "COACH_AGENT_NAME"
	CoachBackendURLEnvVar = "COACH_BACKEND_URL"
)

var (
	coachBackendService *services.CoachBackendService
	coachAgentService   *services.CoachAgentService
)

func main() {
	// Get environment variables
	coachAgentHost := os.Getenv(CoachAgentHostEnvVar)
	if coachAgentHost == "" {
		panic("Coach agent host is not set (" + CoachAgentHostEnvVar + ")")
	}

	coachAgentPort := os.Getenv(CoachAgentPortEnvVar)
	if coachAgentPort == "" {
		panic("Coach agent local port is not set (" + CoachAgentPortEnvVar + ")")
	}

	coachAgentName := os.Getenv(CoachAgentNameEnvVar)
	if coachAgentName == "" {
		panic("Coach agent name is not set (" + CoachAgentNameEnvVar + ")")
	}

	coachBackendUrl := os.Getenv(CoachBackendURLEnvVar)
	if coachBackendUrl == "" {
		panic("Coach backend URL is not set (" + CoachBackendURLEnvVar + ")")
	}

	coachBackendService = services.NewCoachBackend(coachBackendUrl)
	coachAgentService = services.NewCoachAgentService(coachAgentHost, coachAgentPort, coachAgentName, coachBackendService)

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

	r.Route("/api/v1", func(r chi.Router) {
		//r.Post("/chat", handlePrompt)
		r.Delete("/chat", cleanSession)
		r.Get("/chat/stream", handleChatStream) // NEW WEBSOCKET ROUTE
	})

	// Get the port from the env var
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
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

func cleanSession(w http.ResponseWriter, r *http.Request) {
	// Call the delete method of the session endpoint
	user, err := getUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = coachAgentService.CleanSession(user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Session cleaned successfully"))
}

func getUser(r *http.Request) (user string, err error) {
	// TODO
	return "guillaume.blaquiere@gmail.com", nil
}

//**************************
//			Models
//**************************

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// For development, accept everything.
		// In production, use strict origin validation.
		return true
	},
}

// handleChatStream manages the WebSocket connection and proxies it to the Python agent.
func handleChatStream(w http.ResponseWriter, r *http.Request) {
	browserConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading client WebSocket: %v", err)
		return
	}
	defer browserConn.Close()
	log.Println("Client (browser) connected via WebSocket.")

	user, _ := getUser(r) // Get the user (even if it's hardcoded for now)

	err = coachAgentService.StreamSession(user, browserConn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
