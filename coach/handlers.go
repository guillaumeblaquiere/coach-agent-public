// C:/Users/Guillou/IdeaProjects/coach/coach/handlers.go

package main

import (
	"cloud.google.com/go/firestore"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	dailyPlansCollection = "dailyTrainingPlans"
	defaultDateLayout    = "2006-01-02"
	defaultRequestSource = "API"
)

var (
	defaultPlanTemplate = PlanTemplate{
		ID:         "default",
		Categories: Categories,
		CreatedAt:  time.Time{},
		UpdatedAt:  time.Time{},
	}
)

// API provides application-wide context, like the Firestore client.
type API struct {
	fsClient    *firestore.Client
	connManager *ConnectionManager
	// logger *log.Logger // Recommended for production
}

// NewAPI creates a new API instance.
func NewAPI(fs *firestore.Client) *API {
	return &API{fsClient: fs, connManager: NewConnectionManager()}
}

// --- Helper Functions ---

func (api *API) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user email from query parameters
	userEmail := r.URL.Query().Get("email")
	if userEmail == "" {
		http.Error(w, "User email is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection for %s: %v", userEmail, err)
		return
	}

	// Register the new connection
	api.connManager.Add(userEmail, conn)

	// Use defer to ensure the connection is cleaned up on exit
	defer func() {
		api.connManager.Remove(userEmail, conn)
		conn.Close()
	}()

	// Read loop to detect when the client closes the connection.
	// We don't need to process incoming messages here for now.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("Client %s disconnected.", userEmail)
			break // Exit the loop to trigger the defer
		}
	}
}

func (a *API) respondWithError(w http.ResponseWriter, code int, message string) {
	a.respondWithJSON(w, code, map[string]string{"error": message})
}

func (a *API) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		// Fallback if JSON marshaling fails
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to marshal JSON response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *API) parseDateParam(dateStr string) (string, error) {
	if dateStr == "today" {
		return time.Now().UTC().Format(defaultDateLayout), nil
	}
	_, err := time.Parse(defaultDateLayout, dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %s, use YYYY-MM-DD or 'today'", dateStr)
	}
	return dateStr, nil
}

// --- Category Handlers ---

func (a *API) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories := make([]Category, 0)
	for _, category := range Categories {
		category.Drills = make(map[string]Drill) // Clear drills
		categories = append(categories, category)
	}
	a.respondWithJSON(w, http.StatusOK, categories)
}

func (a *API) GetCategory(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryId")
	category, ok := Categories[categoryID]
	if !ok {
		a.respondWithError(w, http.StatusNotFound, "Category not found")
		return
	}
	a.respondWithJSON(w, http.StatusOK, category)
}

// --- Drill Handlers (Similar structure to Category Handlers) ---

func (a *API) ListDrills(w http.ResponseWriter, r *http.Request) {
	var drills []Drill
	for _, category := range Categories {
		for _, drill := range category.Drills {
			drills = append(drills, drill)
		}
	}
	a.respondWithJSON(w, http.StatusOK, drills)
}

func (a *API) GetDrill(w http.ResponseWriter, r *http.Request) {
	drillID := chi.URLParam(r, "drillId")
	var drill Drill
	for _, category := range Categories {
		for _, d := range category.Drills {
			if d.ID == drillID {
				drill = d
				break
			}
		}
	}
	if drill.ID == "" {
		a.respondWithError(w, http.StatusNotFound, "Drill not found")
		return
	}
	a.respondWithJSON(w, http.StatusOK, drill)
}

// --- Training Plan Template Handlers ---

func (a *API) ListPlanTemplates(w http.ResponseWriter, r *http.Request) {
	a.respondWithJSON(w, http.StatusOK, []string{"default"}) //only default accepted for now
}

func (a *API) GetPlanTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	if templateID != "default" {
		a.respondWithError(w, http.StatusNotFound, "Only 'default' template ID is supported currently.")
		return
	}
	a.respondWithJSON(w, http.StatusOK, defaultPlanTemplate)
}

func (a *API) createDefaultDailyPlan() (plan DailyTrainingPlan) {
	now := time.Now().UTC()
	plan = DailyTrainingPlan{
		ID:         "default",
		TemplateID: defaultPlanTemplate.ID,
		Date:       now.Format(defaultDateLayout),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	repetitions := make(map[string]map[string]Achievement)
	for _, category := range defaultPlanTemplate.Categories {
		repetitions[category.ID] = make(map[string]Achievement)
		for _, drill := range category.Drills {
			repetitions[category.ID][drill.ID] = Achievement{
				Repetition: 0,
				Note:       "",
				CreatedAt:  now,
				UpdatedAt:  now,
			}
		}
	}
	plan.Repetitions = repetitions
	return plan
}

// --- Daily Training Plan Handlers ---
// TODO add templateID
func (a *API) InitiateDailyPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dateStr := time.Now().UTC().Format(defaultDateLayout)
	userEmail, err := a.getEmailFromHeader(r)
	if err != nil {
		fmt.Printf("Error getting user email: %v\n", err)
		a.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	docID, err := a.getDocID(r, userEmail, dateStr)
	if err != nil {
		a.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	planDocRef := a.fsClient.Collection(dailyPlansCollection).Doc(docID)

	// Create a default plan
	dailyPlan := a.createDefaultDailyPlan()
	dailyPlan.ID = docID

	// Save the plan
	_, err = planDocRef.Set(ctx, dailyPlan) // Use Set with specific ID (dateStr)
	if err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Failed to create daily plan: "+err.Error())
		return
	}
	a.respondWithJSON(w, http.StatusCreated, dailyPlan)
}

func (a *API) GetDailyPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dateParam := chi.URLParam(r, "date") // "YYYY-MM-DD" or "today"

	dateStr, err := a.parseDateParam(dateParam)
	if err != nil {
		a.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userEmail, err := a.getEmailFromHeader(r)
	if err != nil {
		fmt.Printf("Error getting user email: %v\n", err)
		a.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	docID, err := a.getDocID(r, userEmail, dateStr)
	if err != nil {
		a.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	docSnap, err := a.fsClient.Collection(dailyPlansCollection).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// If dateStr is today, create the daily plan
			if dateStr == time.Now().UTC().Format(defaultDateLayout) {
				fmt.Sprintf(`"Daily plan for %s not found. Auto-initiating today's plan...", docID`)
				// Create the default plan for today
				dailyPlan := a.createDefaultDailyPlan()
				dailyPlan.ID = docID
				_, createErr := a.fsClient.Collection(dailyPlansCollection).Doc(docID).Set(ctx, dailyPlan)
				if createErr != nil {
					a.respondWithError(w, http.StatusInternalServerError, "Failed to auto-initiate today's plan: "+createErr.Error())
					return
				}
				// Return the newly created plan
				a.respondWithJSON(w, http.StatusOK, dailyPlan)
				return
			}
			// For historical dates not found, return 404
			a.respondWithError(w, http.StatusNotFound, fmt.Sprintf("Daily plan for %s not found. Initiate it first.", docID))
		} else {
			a.respondWithError(w, http.StatusInternalServerError, "Failed to get daily plan: "+err.Error())
		}
		return
	}
	var plan DailyTrainingPlan
	if err := docSnap.DataTo(&plan); err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Failed to parse daily plan data: "+err.Error())
		return
	}
	a.respondWithJSON(w, http.StatusOK, plan)
}

func (a *API) UpdateTodayDailyPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now().UTC()
	todayStr := now.Format(defaultDateLayout)

	source := a.getRequestSource(r)
	fmt.Printf("Received request from source: %s\n", source)

	// Body contain only the updated drills
	var updatedPlan DailyTrainingPlan

	body, _ := io.ReadAll(r.Body)
	fmt.Printf("body: %s\n", body) // Keep for debugging if needed

	if err := json.Unmarshal(body, &updatedPlan); err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	// fmt.Printf("Received updated plan: %+v\n", updatedPlan) // Keep for debugging if needed

	// Ensure the update is for today and ID matches
	if updatedPlan.Date != "" && updatedPlan.Date != todayStr {
		a.respondWithError(w, http.StatusBadRequest, "Can only update today's plan via this endpoint. Date in payload mismatches today.")
		return
	}

	updates := []firestore.Update{
		{Path: "updatedAt", Value: now},
	}
	for categoryId, drills := range updatedPlan.Repetitions {
		for drillId, achievement := range drills {
			fieldPathBase := fmt.Sprintf("repetitions.%s.%s.", categoryId, drillId)
			updates = append(updates,
				firestore.Update{
					Path:  fieldPathBase + "repetition",
					Value: achievement.Repetition,
				},
				firestore.Update{
					Path:  fieldPathBase + "note",
					Value: achievement.Note,
				},
				firestore.Update{
					Path:  fieldPathBase + "updatedAt",
					Value: now,
				})
		}
	}

	userEmail, err := a.getEmailFromHeader(r)
	if err != nil {
		fmt.Printf("Error getting user email: %v\n", err)
		a.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	docID, err := a.getDocID(r, userEmail, todayStr)
	if err != nil {
		a.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_, err = a.fsClient.Collection(dailyPlansCollection).Doc(docID).Update(ctx, updates)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			a.respondWithError(w, http.StatusNotFound, "Today's plan not found to update. Initiate it first.")
		} else {
			a.respondWithError(w, http.StatusInternalServerError, "Failed to update today's plan: "+err.Error())
		}
		return
	}

	// Fetch and return the updated plan
	updatedDocSnap, err := a.fsClient.Collection(dailyPlansCollection).Doc(docID).Get(ctx)
	if err != nil {
		// This shouldn't happen if the update succeeded, but handle defensively
		a.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve updated plan: "+err.Error())
		return
	}
	var finalPlan DailyTrainingPlan
	if err := updatedDocSnap.DataTo(&finalPlan); err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Failed to parse updated plan data: "+err.Error())
		return
	}

	if finalPlan.ID != "" { // Ensure the update succeeded
		log.Printf("Sending WebSocket update to user: %s", userEmail)
		message := WebSocketMessage{
			Action: "PLAN_UPDATED",
			// FIX: Send the full, final plan object, not the partial update from the request.
			Data:   finalPlan,
			Source: source,
		}
		a.connManager.SendMessage(userEmail, message)
	}

	a.respondWithJSON(w, http.StatusOK, finalPlan)
}

func (a *API) getDocID(r *http.Request, userEmail, dateStr string) (docID string, err error) {
	return fmt.Sprintf("%s-%s", userEmail, dateStr), nil
}

// Get email from header
func (a *API) getEmailFromHeader(r *http.Request) (string, error) {
	// In a real application, this would involve authenticating the user,
	// e.g., via a JWT token in the request headers.
	// For this example, we'll return a placeholder user.
	userEmail := "guillaume.blaquiere@gmail.com"
	return userEmail, nil
}

func (a *API) getRequestSource(r *http.Request) (source string) {
	// Get the source query parameter or set API as default one
	source = r.URL.Query().Get("source")
	if source == "" {
		source = defaultRequestSource
	}
	return strings.ToLower(strings.TrimSpace(source))
}

func isValidEmail(email string) bool {
	// Very basic email format check for demonstration
	return len(email) > 0 && strings.Contains(email, "@")

}
