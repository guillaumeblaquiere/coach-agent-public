package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gblaquiere.dev/wrapper_agent/models"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	CoachAgentBaseURL                = "http://%s:%s"
	CoachAgentSessionPath            = "/apps/%s/users/%s/sessions/%s"
	CoachAgentSessionAgentEnginePath = "/apps/%s/users/%s/agentEngineSessions"
	CoachAgentStreamingPath          = "/run_live"
)

type CoachAgentService struct {
	coachAgentBaseURL   string
	coachAgentName      string
	coachBackendService *CoachBackendService
	sessionMapUserId    map[string]string
}

func NewCoachAgentService(host, port, name string, coachBackendService *CoachBackendService) *CoachAgentService {
	coachBaseUrl := fmt.Sprintf(CoachAgentBaseURL, host, port)

	return &CoachAgentService{
		coachAgentBaseURL:   coachBaseUrl,
		coachAgentName:      name,
		coachBackendService: coachBackendService,
		sessionMapUserId:    make(map[string]string),
	}
}

func (cas *CoachAgentService) InitSession(user string) (sessionId string, err error) {
	sessionId, err = cas.GetSessionId(user)
	if err != nil {
		log.Printf("Error getting session ID: %v", err)
		return
	}

	if sessionId == "" {
		// Create the session and save it
		sessionId, err = cas.createSession(user)
		if err != nil {
			log.Printf("Error creating session: %v", err)
			return
		}
		cas.sessionMapUserId[user] = sessionId
		err = cas.coachBackendService.UpdateSessionID(user, sessionId)
		if err != nil {
			log.Printf("Error updating daily plan: %v", err)
			return
		}
	}
	return
}

func (cas *CoachAgentService) createSession(user string) (sessionId string, err error) {

	coachUrl := cas.coachAgentBaseURL + fmt.Sprintf(CoachAgentSessionAgentEnginePath, cas.coachAgentName, user)

	initialState := cas.getSessionStateDetail(user)

	jsonData, err := json.Marshal(initialState)
	if err != nil {
		fmt.Printf("failed to marshal initial state: %v\n", err)
		return "", fmt.Errorf("failed to marshal initial state: %w", err)
	}

	req, err := http.NewRequest("POST", coachUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("failed to create session creation request: %v\n", err)
		return "", fmt.Errorf("failed to send session creation request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("failed to send session creation request: %v\n", err)
		return "", fmt.Errorf("failed to send session creation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("failed to create session, status code: %d\n", resp.StatusCode)
		return "", fmt.Errorf("failed to create session, status code: %d", resp.StatusCode)
	}
	fmt.Printf("session created\n")

	var adkSession models.ADKSession
	err = json.NewDecoder(resp.Body).Decode(&adkSession)
	if err != nil {
		fmt.Printf("failed to decode session creation response: %v\n", err)
		return "", fmt.Errorf("failed to decode session creation response: %w", err)
	}

	return adkSession.Id, nil
}

// Delete the old session and generate a new one. Save the new ID to the backend and map
func (cas *CoachAgentService) CleanSession(user string) (err error) {
	sessionId, err := cas.GetSessionId(user)
	if err != nil {
		log.Printf("Error getting session ID: %v", err)
		return
	}

	if sessionId != "" {

		// Delete the session
		url := cas.coachAgentBaseURL + fmt.Sprintf(CoachAgentSessionPath, cas.coachAgentName, user, sessionId)
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create session delete request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send session delete request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("Session %s for user %s cleaned", sessionId, user)
		} else if resp.StatusCode == http.StatusNotFound {
			log.Printf("Session %s for user %s not found for cleaning", sessionId, user)
		} else {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("Failed to clean session %s for user %s, status code: %d", sessionId, user, resp.StatusCode)
			return fmt.Errorf("failed to clean session, status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
		}
	}

	// Init a new session
	sessionId, err = cas.createSession(user)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		return
	}
	cas.sessionMapUserId[user] = sessionId
	err = cas.coachBackendService.UpdateSessionID(user, sessionId)
	if err != nil {
		log.Printf("Error updating daily plan: %v", err)
		return
	}

	return
}

func (cas *CoachAgentService) StreamSession(user string, browserConn *websocket.Conn) (err error) {

	// Create the session, to be sure it exists
	sessionId, err := cas.InitSession(user)
	if err != nil {
		log.Printf("Error during session creation: %v", err)
		return
	}

	// Build the URL for the Python agent
	agentURL := url.URL{
		Scheme: "ws",
		Host:   strings.ReplaceAll(cas.coachAgentBaseURL, "http://", ""),
		Path:   CoachAgentStreamingPath,
	}
	q := agentURL.Query()
	q.Set("app_name", cas.coachAgentName)
	q.Set("user_id", user)
	q.Set("session_id", sessionId)
	q.Set("modalities", "AUDIO")
	agentURL.RawQuery = q.Encode()

	log.Printf("Connecting to backend agent at: %s", agentURL.String())

	// Connect to the Python agent
	agentConn, _, err := websocket.DefaultDialer.Dial(agentURL.String(), nil)
	if err != nil {
		log.Printf("Error connecting to Python WebSocket agent: %v", err)
		// Inform the client that a server-side error occurred
		browserConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Could not reach the agent service"))
		return
	}
	defer agentConn.Close()
	log.Println("Successfully connected to Python WebSocket agent.")

	backendConn, err := cas.coachBackendService.CreateBackendConn(user)
	if err != nil {
		log.Printf("Error connecting to backend event stream: %v", err)
	} else {
		defer backendConn.Close()
		log.Println("Successfully connected to backend event stream.")
		// Start the backend->agent forwarder only if connection succeeded
		go cas.forwardBackendToAgent(backendConn, agentConn)
	}

	// Set up the bidirectional proxy
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine to forward messages from client to agent
	go func() {
		defer wg.Done()
		for {
			messageType, p, err := browserConn.ReadMessage()
			if err != nil {
				log.Printf("Read error (client->agent): %v", err)
				agentConn.WriteMessage(websocket.CloseMessage, []byte{})
				browserConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Could not reach the agent service"))
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
			if err != nil {
				log.Printf("Read error (agent->client): %v", err)
				browserConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := browserConn.WriteMessage(messageType, p); err != nil {
				log.Printf("Write error (agent->client, original): %v", err)
				return
			}
		}
	}()

	wg.Wait()
	log.Println("WebSocket proxy finished.")
	return
}

func (cas *CoachAgentService) forwardBackendToAgent(backendConn, agentConn *websocket.Conn) {
	for {
		_, p, err := backendConn.ReadMessage()
		if err != nil {
			log.Printf("Read error (backend->agent): %v", err)
			return
		}

		var backendMsg struct {
			Action string      `json:"action"`
			Data   interface{} `json:"data"`
			Source string      `json:"source"`
		}
		if err := json.Unmarshal(p, &backendMsg); err != nil {
			log.Printf("Error unmarshalling backend event: %v", err)
			continue
		}

		if backendMsg.Action == "PLAN_UPDATED" && backendMsg.Source != "agent" {
			log.Println("Forwarding PLAN_UPDATED event to agent.")
			agentEvent := map[string]interface{}{
				"mime_type":    "application/json",
				"event_source": backendMsg.Source,
				"event_type":   "plan_updated",
				"data":         backendMsg.Data,
			}
			agentEventBytes, _ := json.Marshal(agentEvent)

			if err := agentConn.WriteMessage(websocket.TextMessage, agentEventBytes); err != nil {
				log.Printf("Write error (backend->agent): %v", err)
				return
			}
		}
	}
}

func (cas *CoachAgentService) getSessionStateDetail(user string) SessionState {
	//TODO get in database the user attribute
	return map[string]interface{}{
		"user:first_name": "Guillaume",
	}
}

func (cas *CoachAgentService) GetSessionId(user string) (sessionId string, err error) {
	ok := false
	sessionId, ok = cas.sessionMapUserId[user]
	if ok {
		return
	}

	sessionId, err = cas.coachBackendService.GetSessionID(user)
	if err != nil {
		log.Printf("Error getting session ID: %v", err)
		return
	}
	return
}

type SessionState map[string]interface{}
