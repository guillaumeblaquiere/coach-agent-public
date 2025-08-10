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
)

type CoachBackendService struct {
	backendURL string
}

func NewCoachBackend(backendURL string) *CoachBackendService {
	return &CoachBackendService{
		backendURL: backendURL,
	}
}

func (cb *CoachBackendService) GetSessionID(user string) (sessionId string, err error) {
	// Get the daily plan
	dailyPlan, err := cb.getDailyPlan(user)
	if err != nil {
		log.Printf("Error getting daily plan: %v", err)
		return
	}
	return dailyPlan.SessionID, nil
}

func (cb *CoachBackendService) getDailyPlan(user string) (dailyPlan *models.DailyTrainingPlan, err error) {
	// create a request and add the user in the header
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/daily-plans/today?source=agent", cb.backendURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-User-Email", user) // Set the user email in the header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily plan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get daily plan: status code %d", resp.StatusCode)
	}

	// Load the body into the dailyPlan struct
	dailyPlan = &models.DailyTrainingPlan{}
	if err = json.NewDecoder(resp.Body).Decode(dailyPlan); err != nil {
		return nil, fmt.Errorf("failed to decode daily plan: %w", err)
	}

	return dailyPlan, nil
}

func (cb *CoachBackendService) CreateBackendConn(user string) (backendConn *websocket.Conn, err error) {
	backendWsURL, err := url.Parse(cb.backendURL)
	if err != nil {
		log.Printf("Error parsing backend URL: %v", err)
		return
	}
	fmt.Printf("backendWsURL is: %+v\n", backendWsURL)
	backendWsURL.Scheme = "ws"
	backendWsURL.Path = "/api/v1/ws"
	backendWsQuery := backendWsURL.Query()
	backendWsQuery.Set("email", user)
	backendWsURL.RawQuery = backendWsQuery.Encode()

	log.Printf("Connecting to backend event stream at: %s", backendWsURL.String())
	backendConn, _, err = websocket.DefaultDialer.Dial(backendWsURL.String(), nil)
	return
}

func (cb *CoachBackendService) UpdateSessionID(user string, sessionId string) (err error) {
	// Get the daily plan
	plan, err := cb.getDailyPlan(user)
	if err != nil {
		return fmt.Errorf("failed to get daily plan: %w", err)
	}

	// Optimize the data shared and only send the plan ID and session ID
	newPlan := &models.DailyTrainingPlan{
		ID:        plan.ID,
		SessionID: sessionId,
		Date:      plan.Date,
	}

	// Perform a Put to the backend API with the user in the header
	requestBody, err := json.Marshal(newPlan)
	if err != nil {
		return fmt.Errorf("failed to marshal daily plan: %w", err)
	}

	//fmt.Printf("requestBody %s", string(requestBody))

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/daily-plans/today?source=agent", cb.backendURL), io.NopCloser(bytes.NewBuffer(requestBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Email", user) // Set the user email in the header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update daily plan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update daily plan: status code %d", resp.StatusCode)
	}

	return nil
}
