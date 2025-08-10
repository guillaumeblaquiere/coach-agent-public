package models

import (
	"time"
)

// Category represents a category for drills (e.g., Strength, Cardio).
type Category struct {
	ID          string            `firestore:"id" json:"id"` // Firestore document ID, set after retrieval/creation
	Name        map[string]string `firestore:"name" json:"name"`
	Description map[string]string `firestore:"description,omitempty" json:"description,omitempty"`
	Drills      map[string]Drill  `firestore:"drills,omitempty" json:"drills,omitempty"`
}

// Drill represents a specific exercise.
type Drill struct {
	ID               string            `firestore:"id" json:"id"`
	Name             map[string]string `firestore:"name" json:"name"`
	Description      map[string]string `firestore:"description,omitempty" json:"description,omitempty"`
	CategoryID       string            `firestore:"categoryId" json:"categoryId"`
	TargetRepetition int               `firestore:"targetRepetition,omitempty" json:"targetRepetition,omitempty"`
	CreatedAt        time.Time         `firestore:"createdAt" json:"-"`
	UpdatedAt        time.Time         `firestore:"updatedAt" json:"-"`
}

type PlanTemplate struct {
	ID         string              `firestore:"id" json:"id"`
	Categories map[string]Category `firestore:"categories,omitempty" json:"categories,omitempty"`
	CreatedAt  time.Time           `firestore:"createdAt" json:"-"`
	UpdatedAt  time.Time           `firestore:"updatedAt" json:"-"`
}

// DailyTrainingPlan is an instance of a training plan for a specific day.
// The Firestore document ID for this could be the date string "YYYY-MM-DD".
type DailyTrainingPlan struct {
	ID          string                            `firestore:"id" json:"id"`
	TemplateID  string                            `firestore:"templateId,omitempty" json:"templateId,omitempty"`
	Date        string                            `firestore:"date" json:"date"` // "YYYY-MM-DD"
	SessionID   string                            `firestore:"sessionId,omitempty" json:"sessionId,omitempty"`
	Repetitions map[string]map[string]Achievement `firestore:"repetitions,omitempty" json:"repetitions,omitempty"`
	CreatedAt   time.Time                         `firestore:"createdAt" json:"-"`
	UpdatedAt   time.Time                         `firestore:"updatedAt" json:"-"`
}

type Achievement struct {
	Repetition int       `firestore:"repetition" json:"repetition"`
	Note       string    `firestore:"note,omitempty" json:"note,omitempty"`
	CreatedAt  time.Time `firestore:"createdAt" json:"-"`
	UpdatedAt  time.Time `firestore:"updatedAt" json:"-"`
}

// LocalizedCategory is a temporary struct to hold localized data for a category.
type LocalizedCategory struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Drills      map[string]LocalizedDrill `json:"drills,omitempty"`
}

// LocalizedDrill is a temporary struct to hold localized data for a drill.
type LocalizedDrill struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	CategoryID       string    `json:"categoryId"`
	TargetRepetition int       `json:"targetRepetition,omitempty"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

// LocalizedPlanTemplate is a temporary struct to hold localized data for a plan template.
type LocalizedPlanTemplate struct {
	ID         string                       `json:"id"`
	Categories map[string]LocalizedCategory `json:"categories,omitempty"`
	CreatedAt  time.Time                    `json:"-"`
	UpdatedAt  time.Time                    `json:"-"`
}
