package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	// ... other imports ...
	"cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

const (
	CoachAgentPortEnvVar = "COACH_AGENT_PORT"
	CoachAgentHostEnvVar = "COACH_AGENT_HOST"
	CoachAgentNameEnvVar = "COACH_AGENT_NAME"

	CoachAgentBaseURL       = "http://%s:%s"
	CoachAgentSessionPath   = "/apps/%s/users/%s/sessions/%s"
	CoachAgentStreamingPath = "/run_live"
)

var (
	coachBaseUrl   string
	coachAgentPort string
	coachAgentHost string
	coachAgentName string
	ttsClient      *texttospeech.Client
)

func main() {
	// Get environment variables
	coachAgentHost = os.Getenv(CoachAgentHostEnvVar)
	if coachAgentHost == "" {
		panic("Coach agent host is not set (" + CoachAgentHostEnvVar + ")")
	}

	coachAgentPort = os.Getenv(CoachAgentPortEnvVar)
	if coachAgentPort == "" {
		panic("Coach agent local port is not set (" + CoachAgentPortEnvVar + ")")
	}

	coachAgentName = os.Getenv(CoachAgentNameEnvVar)
	if coachAgentName == "" {
		panic("Coach agent name is not set (" + CoachAgentNameEnvVar + ")")
	}

	coachBaseUrl = fmt.Sprintf(CoachAgentBaseURL, coachAgentHost, coachAgentPort)

	var err error
	ttsClient, err = texttospeech.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create Text-to-Speech client: %v", err)
	}

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
		r.Post("/chat", handlePrompt)
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

	sessionID := getSessionID() // Get the session ID for today

	url := coachBaseUrl + fmt.Sprintf(CoachAgentSessionPath, coachAgentName, user, sessionID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create session delete request: %v", err), http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to send session delete request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Session cleaned successfully"))
		log.Printf("Session %s for user %s cleaned", sessionID, user)
	} else if resp.StatusCode == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Session not found"))
		log.Printf("Session %s for user %s not found for cleaning", sessionID, user)
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("failed to clean session, status code: %d, body: %s", resp.StatusCode, string(bodyBytes)), http.StatusInternalServerError)
		log.Printf("Failed to clean session %s for user %s, status code: %d", sessionID, user, resp.StatusCode)
	}
}

func getSessionID() (sessionID string) {
	return time.Now().UTC().Format("2006-01-02")
}

func handlePrompt(w http.ResponseWriter, r *http.Request) {
	// Get the user
	user, err := getUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get the prompt mime type
	mimeType := r.Header.Get("Content-Type")
	if mimeType != "application/json" {
		http.Error(w, "Invalid mime type", http.StatusBadRequest)
		return
	}

	// Parse the request
	var req WrapperRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	//fmt.Printf("req is: %+v\n", req)
	//TODO check supported mime type of the REQ inline data

	wrapperResponse, err := AskAgent(user, getSessionID(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Add audio version of the agent text response with Text to speech API
	if wrapperResponse.Part.Text != "" {
		audioContent, err := textToSpeech(wrapperResponse.Part.Text)
		if err != nil {
			log.Printf("Error converting text to speech: %v", err)
			// It's up to you how to handle this. You might still send the text response,
			// or return an error to the client. Here, we'll just log and continue with only the text:
		} else {
			wrapperResponse.Part.InlineData = &InlineData{
				MimeType: "audio/mp3", // or "audio/wav" depending on your settings
				Data:     audioContent,
			}
		}
	}

	// Send the response back to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(wrapperResponse)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}

func textToSpeech(text string) (string, error) {
	// Remove special character from text
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "fr-FR", // Or whatever language you need.  Consider making this configurable.
			Name:         "fr-FR-Chirp3-HD-Sulafat",
			//SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL, // Or other gender options. Also consider configuration.
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3, // or WAV, LINEAR16, etc.
			SpeakingRate:  1.1,
			//  You can adjust sample rate, speaking rate, pitch, etc. here.  See the docs.
		},
	}

	resp, err := ttsClient.SynthesizeSpeech(context.Background(), &req)
	if err != nil {
		return "", fmt.Errorf("failed to synthesize text to speech: %w", err)
	}

	// The encode the audio in b64
	return base64.StdEncoding.EncodeToString(resp.AudioContent), nil
}

func AskAgent(user string, sessionID string, req WrapperRequest) (wrapperResponse WrapperResponse, err error) {

	status, err := initSession(user, sessionID)
	if err != nil {
		wrapperResponse = WrapperResponse{
			Status: "error",
			Error:  fmt.Sprintf("failed to initialize session: %v", err),
		}
		return wrapperResponse, fmt.Errorf("failed to initialize session: %w", err)
	}

	parts := []Part{Part(req)}
	if status == http.StatusCreated {
		parts = []Part{
			{
				Text: " ",
			},
			Part(req),
		}
	}

	// Create the ADK request
	adkReq := AdkRequest{
		AppName:   coachAgentName,
		UserId:    user,
		SessionId: sessionID,
		NewMessage: NewMessage{
			Role:  "user",
			Parts: parts,
		},
	}

	// Send the request to the ADK agent
	adkReqBody, err := json.Marshal(adkReq)
	if err != nil {
		wrapperResponse = WrapperResponse{
			Status: "error",
			Error:  fmt.Sprintf("failed to marshal ADK request: %v", err),
		}
		return wrapperResponse, fmt.Errorf("failed to marshal ADK request: %w", err)
	}

	//fmt.Printf("ADK request: %s\n", string(adkReqBody))

	adkResponse, err := http.Post("http://localhost:"+coachAgentPort+"/run", "application/json", bytes.NewBuffer(adkReqBody))
	if err != nil {
		wrapperResponse = WrapperResponse{
			Status: "error",
			Error:  fmt.Sprintf("failed to send request to ADK agent: %v", err),
		}
		return wrapperResponse, fmt.Errorf("failed to send request to ADK agent: %w", err)

	}
	defer adkResponse.Body.Close()

	// Check the return code
	if adkResponse.StatusCode != http.StatusOK {
		adkResponseBody, _ := io.ReadAll(adkResponse.Body)
		wrapperResponse = WrapperResponse{
			Status: "error",
			Error:  fmt.Sprintf("ADK agent returned status code %d: %s", adkResponse.StatusCode, string(adkResponseBody)),
		}
		return wrapperResponse, fmt.Errorf("ADK agent returned status code %d: %s", adkResponse.StatusCode, string(adkResponseBody))
	}

	// Read the ADK response
	adkResponseBody, err := io.ReadAll(adkResponse.Body)
	if err != nil {
		wrapperResponse = WrapperResponse{
			Status: "error",
			Error:  fmt.Sprintf("failed to read ADK response body: %v", err),
		}
		return wrapperResponse, fmt.Errorf("failed to read ADK response body: %w", err)
	}

	//fmt.Printf("ADK response: %s\n", string(adkResponseBody))

	// Parse the ADK response
	var adkResp []AdkResponse
	err = json.Unmarshal(adkResponseBody, &adkResp)
	if err != nil {
		wrapperResponse = WrapperResponse{
			Status: "error",
			Error:  fmt.Sprintf("failed to unmarshal ADK response: %v", err),
		}
		return wrapperResponse, fmt.Errorf("failed to unmarshal ADK response: %w", err)
	}

	// Extract the text from the ADK response
	var responseJson strings.Builder
	if adkResp != nil && len(adkResp) > 0 {
		for _, respPart := range adkResp {
			if respPart.Content.Parts != nil && len(respPart.Content.Parts) > 0 {
				if respPart.Content.Parts[0].Text != "" {
					responseJson.WriteString(respPart.Content.Parts[0].Text)
				}
			}
		}
	}

	// Create the wrapper response
	wrapperResponse = WrapperResponse{
		Status: "success",
		Part: Part{
			Text: responseJson.String(),
		},
	}

	//fmt.Printf("wrapperResponse is: %+v\n", wrapperResponse.Part.Text)

	return
}

func initSession(user string, sessionID string) (status int, err error) {
	url := coachBaseUrl + fmt.Sprintf(CoachAgentSessionPath, coachAgentName, user, sessionID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to create session request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to send session request: %v\n", err)
		return http.StatusInternalServerError, fmt.Errorf("failed to send session request: %w", err)
	}
	defer resp.Body.Close()

	// If the session doesn't exist (404), we don't need to do anything,
	// the ADK agent will create it on the first /run call.
	if resp.StatusCode == http.StatusNotFound {
		// make a post request to create the session
		req, err = http.NewRequest("POST", url, nil)
		if err != nil {
			fmt.Printf("failed to create session creation request: %v\n", err)
			return http.StatusInternalServerError, fmt.Errorf("failed to create session creation request: %w", err)
		}
		resp, err = client.Do(req)
		if err != nil {
			fmt.Printf("failed to send session creation request: %v\n", err)
			return http.StatusInternalServerError, fmt.Errorf("failed to send session creation request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("failed to create session, status code: %d\n", resp.StatusCode)
			return http.StatusInternalServerError, fmt.Errorf("failed to create session, status code: %d", resp.StatusCode)
		}
		fmt.Printf("session created\n")
		return http.StatusCreated, nil
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("session already exists\n")
		return http.StatusFound, nil
	} else {
		fmt.Printf("failed to initialize session, status code: %d\n", resp.StatusCode)
		return http.StatusInternalServerError, fmt.Errorf("failed to initialize session, status code: %d", resp.StatusCode)
	}

}

func getUser(r *http.Request) (user string, err error) {
	// TODO
	return "guillaume.blaquiere@gmail.com", nil
}

//**************************
//			Models
//**************************

type WrapperRequest Part

type WrapperResponse struct {
	Part   Part   `json:"part"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type AdkRequest struct {
	AppName    string     `json:"appName"`
	UserId     string     `json:"userId"`
	SessionId  string     `json:"sessionId"`
	NewMessage NewMessage `json:"newMessage"`
}

type AdkResponseContent struct {
	Parts []struct {
		Text             string            `json:"text"`
		FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	} `json:"parts"`
	Role string `json:"role"`
}

type FunctionResponse struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type AdkResponse struct {
	Content       AdkResponseContent       `json:"content"`
	UsageMetadata AdkResponseUsageMetadata `json:"usageMetadata"`
	InvocationId  string                   `json:"invocationId"`
	Author        string                   `json:"author"`
	Actions       AdkResponseActions       `json:"actions"`
	Id            string                   `json:"id"`
	Timestamp     float64                  `json:"timestamp"`
}

type AdkResponseUsageMetadata struct {
	CandidatesTokenCount    int `json:"candidatesTokenCount"`
	CandidatesTokensDetails []struct {
		Modality   string `json:"modality"`
		TokenCount int    `json:"tokenCount"`
	} `json:"candidatesTokensDetails"`
	PromptTokenCount    int `json:"promptTokenCount"`
	PromptTokensDetails []struct {
		Modality   string `json:"modality"`
		TokenCount int    `json:"tokenCount"`
	} `json:"promptTokensDetails"`
	TotalTokenCount int    `json:"totalTokenCount"`
	TrafficType     string `json:"trafficType"`
}

type AdkResponseActions struct {
	StateDelta           struct{} `json:"stateDelta"`
	ArtifactDelta        struct{} `json:"artifactDelta"`
	RequestedAuthConfigs struct{} `json:"requestedAuthConfigs"`
}

type InlineData struct {
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`
}

type NewMessage struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// For development, accept everything.
		// In production, use strict origin validation.
		return true
	},
}

// handleChatStream manages the WebSocket connection and proxies it to the Python agent.
func handleChatStream(w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading client WebSocket: %v", err)
		return
	}
	defer clientConn.Close()
	log.Println("Client (browser) connected via WebSocket.")

	user, _ := getUser(r) // Get the user (even if it's hardcoded for now)
	sessionID := getSessionID()

	// Create the session, to be sure it exists
	_, err = initSession(user, sessionID)
	if err != nil {
		log.Printf("Error during session creation: %v", err)
		return
	}

	// Build the URL for the Python agent
	agentURL := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%s", coachAgentHost, coachAgentPort), // Your Python agent's port
		Path:   CoachAgentStreamingPath,
	}
	q := agentURL.Query()
	q.Set("app_name", coachAgentName)
	q.Set("user_id", user)
	q.Set("session_id", sessionID)
	q.Set("modalities", "AUDIO")
	agentURL.RawQuery = q.Encode()

	log.Printf("Connecting to backend agent at: %s", agentURL.String())

	// Connect to the Python agent
	agentConn, _, err := websocket.DefaultDialer.Dial(agentURL.String(), nil)
	if err != nil {
		log.Printf("Error connecting to Python WebSocket agent: %v", err)
		// Inform the client that a server-side error occurred
		clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Could not reach the agent service"))
		return
	}
	defer agentConn.Close()
	log.Println("Successfully connected to Python WebSocket agent.")

	// Set up the bidirectional proxy
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine to forward messages from client to agent
	go func() {
		defer wg.Done()
		for {
			messageType, p, err := clientConn.ReadMessage()
			if err != nil {
				log.Printf("Read error (client->agent): %v", err)
				agentConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			/*			// --- NEW: Debugging log ---
						if messageType == websocket.TextMessage {
							var msg map[string]interface{}
							if json.Unmarshal(p, &msg) == nil {
								// It's a valid JSON
								mimeType, _ := msg["mime_type"].(string)
								data, _ := msg["data"].(string)
								if mimeType != "audio/pcm" {
									log.Printf("[PROXY: Client->Agent] Received JSON message. Mime-Type: %s, Data length: %d", mimeType, len(data))
								}
							} else {
								// Not a JSON, or unexpected structure
								log.Printf("[PROXY: Client->Agent] Received Text message, length: %d", len(p))
							}
						} else {
							log.Printf("[PROXY: Client->Agent] Received Binary message, length: %d", len(p))
						}
						// --- END NEW ---*/

			if err := agentConn.WriteMessage(messageType, p); err != nil {
				log.Printf("Write error (client->agent): %v", err)
				return
			}
		}
	}()

	// Goroutine to forward messages from agent to client
	go func() {
		defer wg.Done()
		for {
			messageType, p, err := agentConn.ReadMessage()
			/*			// --- NEW: Debugging log ---
						if messageType == websocket.TextMessage {
							var msg map[string]interface{}
							if json.Unmarshal(p, &msg) == nil {
								// It's a valid JSON
								mimeType, _ := msg["mime_type"].(string)
								data, _ := msg["data"].(string)
								log.Printf("[PROXY: Agent->Client] Received JSON message. Mime-Type: %s, Data length: %d", mimeType, len(data))
							} else {
								// Not a JSON, or unexpected structure
								log.Printf("[PROXY: Agent->Client] Received Text message, length: %d", len(p))
							}
						} else {
							log.Printf("[PROXY: Agent->Client] Received Binary message, length: %d", len(p))
						}
						// --- END NEW ---*/

			if err != nil {
				log.Printf("Read error (agent->client): %v", err)
				clientConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// --- BUG FIX: Was writing to agentConn, now correctly writes to clientConn ---
			if err := clientConn.WriteMessage(messageType, p); err != nil {
				log.Printf("Write error (agent->client): %v", err)
				return // No need to do anything else, the client connection is likely broken
			}
		}
	}()

	wg.Wait()
	log.Println("WebSocket proxy finished.")
}
