package main

import (
	"cloud.google.com/go/firestore"
	"coach/models"
	"encoding/json"
	"fmt"
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
	defaultLocale        = "en-EN"
)

var (
	defaultPlanTemplate = models.PlanTemplate{
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
}

// NewAPI creates a new API instance.
func NewAPI(fs *firestore.Client) *API {
	return &API{fsClient: fs, connManager: NewConnectionManager()}
}

// --- Helper Functions ---

func (a *API) getLocale(r *http.Request) string {
	locale := r.URL.Query().Get("locale")
	if locale != "fr-FR" && locale != "en-EN" {
		return defaultLocale
	}
	return locale
}

func toLocalizedDrill(drill models.Drill, locale string) models.LocalizedDrill {
	return models.LocalizedDrill{
		ID:               drill.ID,
		Name:             drill.Name[locale],
		Description:      drill.Description[locale],
		CategoryID:       drill.CategoryID,
		TargetRepetition: drill.TargetRepetition,
		CreatedAt:        drill.CreatedAt,
		UpdatedAt:        drill.UpdatedAt,
	}
}

func toLocalizedCategory(category models.Category, locale string, includeDrills bool) models.LocalizedCategory {
	localizedCategory := models.LocalizedCategory{
		ID:          category.ID,
		Name:        category.Name[locale],
		Description: category.Description[locale],
	}
	if includeDrills {
		localizedCategory.Drills = make(map[string]models.LocalizedDrill)
		for drillID, drill := range category.Drills {
			localizedCategory.Drills[drillID] = toLocalizedDrill(drill, locale)
		}
	}
	return localizedCategory
}

func (api *API) handleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	api.connManager.Add(userEmail, conn)
	defer func() {
		api.connManager.Remove(userEmail, conn)
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("Client %s disconnected.", userEmail)
			break
		}
	}
}

func (a *API) respondWithError(w http.ResponseWriter, code int, message string) {
	a.respondWithJSON(w, code, map[string]string{"error": message})
}

func (a *API) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
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
	locale := a.getLocale(r)
	localizedCategories := make([]models.LocalizedCategory, 0, len(Categories))
	for _, category := range Categories {
		localizedCategories = append(localizedCategories, toLocalizedCategory(category, locale, false))
	}
	a.respondWithJSON(w, http.StatusOK, localizedCategories)
}

func (a *API) GetCategory(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryId")
	category, ok := Categories[categoryID]
	if !ok {
		a.respondWithError(w, http.StatusNotFound, "Category not found")
		return
	}
	locale := a.getLocale(r)
	a.respondWithJSON(w, http.StatusOK, toLocalizedCategory(category, locale, true))
}

// --- Drill Handlers ---

func (a *API) ListDrills(w http.ResponseWriter, r *http.Request) {
	locale := a.getLocale(r)
	var drills []models.LocalizedDrill
	for _, category := range Categories {
		for _, drill := range category.Drills {
			drills = append(drills, toLocalizedDrill(drill, locale))
		}
	}
	a.respondWithJSON(w, http.StatusOK, drills)
}

func (a *API) GetDrill(w http.ResponseWriter, r *http.Request) {
	drillID := chi.URLParam(r, "drillId")
	for _, category := range Categories {
		if drill, ok := category.Drills[drillID]; ok {
			locale := a.getLocale(r)
			a.respondWithJSON(w, http.StatusOK, toLocalizedDrill(drill, locale))
			return
		}
	}
	a.respondWithError(w, http.StatusNotFound, "Drill not found")
}

// --- Training Plan Template Handlers ---

func (a *API) ListPlanTemplates(w http.ResponseWriter, r *http.Request) {
	a.respondWithJSON(w, http.StatusOK, []string{"default"})
}

func (a *API) GetPlanTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	if templateID != "default" {
		a.respondWithError(w, http.StatusNotFound, "Only 'default' template ID is supported currently.")
		return
	}

	locale := a.getLocale(r)
	localizedPlan := models.LocalizedPlanTemplate{
		ID:         defaultPlanTemplate.ID,
		Categories: make(map[string]models.LocalizedCategory),
		CreatedAt:  defaultPlanTemplate.CreatedAt,
		UpdatedAt:  defaultPlanTemplate.UpdatedAt,
	}

	for catID, category := range defaultPlanTemplate.Categories {
		localizedPlan.Categories[catID] = toLocalizedCategory(category, locale, true)
	}

	a.respondWithJSON(w, http.StatusOK, localizedPlan)
}

// --- Daily Training Plan Handlers ---

func (a *API) createDefaultDailyPlan() models.DailyTrainingPlan {
	now := time.Now().UTC()
	plan := models.DailyTrainingPlan{
		ID:          "default",
		TemplateID:  defaultPlanTemplate.ID,
		Date:        now.Format(defaultDateLayout),
		CreatedAt:   now,
		UpdatedAt:   now,
		Repetitions: make(map[string]map[string]models.Achievement),
	}

	for catID, category := range defaultPlanTemplate.Categories {
		plan.Repetitions[catID] = make(map[string]models.Achievement)
		for drillID := range category.Drills {
			plan.Repetitions[catID][drillID] = models.Achievement{Repetition: 0, Note: "", CreatedAt: now, UpdatedAt: now}
		}
	}
	return plan
}

func (a *API) InitiateDailyPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dateStr := time.Now().UTC().Format(defaultDateLayout)
	userEmail, err := a.getEmailFromHeader(r)
	if err != nil {
		a.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	docID := a.getDocID(userEmail, dateStr)

	// Create a default plan
	dailyPlan := a.createDefaultDailyPlan()
	dailyPlan.ID = docID

	// Save the plan
	if _, err := a.fsClient.Collection(dailyPlansCollection).Doc(docID).Set(ctx, dailyPlan); err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Failed to create daily plan: "+err.Error())
		return
	}
	a.respondWithJSON(w, http.StatusCreated, dailyPlan)
}

func (a *API) GetDailyPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dateParam := chi.URLParam(r, "date")
	dateStr, err := a.parseDateParam(dateParam)
	if err != nil {
		a.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userEmail, err := a.getEmailFromHeader(r)
	if err != nil {
		a.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	docID := a.getDocID(userEmail, dateStr)

	docSnap, err := a.fsClient.Collection(dailyPlansCollection).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			if dateStr == time.Now().UTC().Format(defaultDateLayout) {
				dailyPlan := a.createDefaultDailyPlan()
				dailyPlan.ID = docID
				if _, createErr := a.fsClient.Collection(dailyPlansCollection).Doc(docID).Set(ctx, dailyPlan); createErr != nil {
					a.respondWithError(w, http.StatusInternalServerError, "Failed to auto-initiate today's plan: "+createErr.Error())
					return
				}
				a.respondWithJSON(w, http.StatusOK, dailyPlan)
				return
			}
			a.respondWithError(w, http.StatusNotFound, fmt.Sprintf("Daily plan for %s not found. Initiate it first.", docID))
		} else {
			a.respondWithError(w, http.StatusInternalServerError, "Failed to get daily plan: "+err.Error())
		}
		return
	}

	var plan models.DailyTrainingPlan
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

	var updatedPlan models.DailyTrainingPlan
	if err := json.NewDecoder(r.Body).Decode(&updatedPlan); err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	fmt.Printf("Received payload: %+v\n", updatedPlan)

	if updatedPlan.Date != "" && updatedPlan.Date != todayStr {
		a.respondWithError(w, http.StatusBadRequest, "Can only update today's plan. Date in payload mismatches today.")
		return
	}

	userEmail, err := a.getEmailFromHeader(r)
	if err != nil {
		a.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	docID := a.getDocID(userEmail, todayStr)

	updates := []firestore.Update{{Path: "updatedAt", Value: now}}
	for categoryId, drills := range updatedPlan.Repetitions {
		for drillId, achievement := range drills {
			fieldPathBase := fmt.Sprintf("repetitions.%s.%s.", categoryId, drillId)
			updates = append(updates, firestore.Update{Path: fieldPathBase + "repetition", Value: achievement.Repetition})
			updates = append(updates, firestore.Update{Path: fieldPathBase + "note", Value: achievement.Note})
			updates = append(updates, firestore.Update{Path: fieldPathBase + "updatedAt", Value: now})
		}
	}

	// Add the session if provided
	if updatedPlan.SessionID != "" {
		updates = append(updates, firestore.Update{Path: "sessionId", Value: updatedPlan.SessionID})
	}

	if _, err := a.fsClient.Collection(dailyPlansCollection).Doc(docID).Update(ctx, updates); err != nil {
		if status.Code(err) == codes.NotFound {
			a.respondWithError(w, http.StatusNotFound, "Today's plan not found to update. Initiate it first.")
		} else {
			a.respondWithError(w, http.StatusInternalServerError, "Failed to update today's plan: "+err.Error())
		}
		return
	}

	updatedDocSnap, err := a.fsClient.Collection(dailyPlansCollection).Doc(docID).Get(ctx)
	if err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve updated plan: "+err.Error())
		return
	}

	var finalPlan models.DailyTrainingPlan
	if err := updatedDocSnap.DataTo(&finalPlan); err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Failed to parse updated plan data: "+err.Error())
		return
	}

	if finalPlan.ID != "" {
		log.Printf("Sending WebSocket update to user: %s", userEmail)
		message := WebSocketMessage{Action: "PLAN_UPDATED", Data: finalPlan, Source: source}
		a.connManager.SendMessage(userEmail, message)
	}

	a.respondWithJSON(w, http.StatusOK, finalPlan)
}

func (a *API) getDocID(userEmail, dateStr string) string {
	return fmt.Sprintf("%s-%s", userEmail, dateStr)
}

func (a *API) getEmailFromHeader(r *http.Request) (string, error) {
	userEmail := "guillaume.blaquiere@gmail.com"
	return userEmail, nil
}

func (a *API) getRequestSource(r *http.Request) string {
	source := r.URL.Query().Get("source")
	if source == "" {
		return defaultRequestSource
	}
	return strings.ToLower(strings.TrimSpace(source))
}
