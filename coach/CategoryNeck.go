package main

import "coach/models"

const (
	CategoryIdNeck = "neck"
)

var (
	DrillNeckFront = models.Drill{
		ID: "NeckFront",
		Name: map[string]string{
			"fr-FR": "Cou avant",
			"en-EN": "Front Neck",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du cou vers l'avant, les deux main sur la tete",
			"en-EN": "Stretching the neck forward, both hands on the head",
		},
		CategoryID:       CategoryIdNeck,
		TargetRepetition: 3,
	}

	DrillNeckLeft = models.Drill{
		ID: "NeckLeft",
		Name: map[string]string{
			"fr-FR": "Cou gauche",
			"en-EN": "Left Neck",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du cou vers la gauche, tete basculée vers la droite, la main droite sur l'oreille gauche",
			"en-EN": "Stretching the neck to the left, head tilted to the right, right hand on the left ear",
		},
		CategoryID:       CategoryIdNeck,
		TargetRepetition: 3,
	}

	DrillNeckRight = models.Drill{
		ID: "NeckRight",
		Name: map[string]string{
			"fr-FR": "Cou droit",
			"en-EN": "Right Neck",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du cou vers la droite, tete basculée vers la gauche, la main droite sur l'oreille droite",
			"en-EN": "Stretching the neck to the right, head tilted to the left, right hand on the right ear",
		},
		CategoryID:       CategoryIdNeck,
		TargetRepetition: 3,
	}

	CategoryNeck = models.Category{
		ID: CategoryIdNeck,
		Name: map[string]string{
			"fr-FR": "Cou",
			"en-EN": "Neck",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du cou et de la tête",
			"en-EN": "Stretching of the neck and head",
		},
		Drills: map[string]models.Drill{
			DrillNeckFront.ID: DrillNeckFront,
			DrillNeckLeft.ID:  DrillNeckLeft,
			DrillNeckRight.ID: DrillNeckRight,
		},
	}
)
