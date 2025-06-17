package main

const (
	CategoryIdNeck = "neck"
)

var (
	DrillNeckFront = Drill{
		ID:               "NeckFront",
		Name:             "Cou avant",
		Description:      "Etirement du cou vers l'avant, les deux main sur la tete",
		CategoryID:       CategoryIdNeck,
		TargetRepetition: 3,
	}

	DrillNeckLeft = Drill{
		ID:               "NeckLeft",
		Name:             "Cou gauche",
		Description:      "Etirement du cou vers la gauche, tete basculée vers la droite, la main droite sur l'oreille gauche",
		CategoryID:       CategoryIdNeck,
		TargetRepetition: 3,
	}

	DrillNeckRight = Drill{
		ID:               "NeckRight",
		Name:             "Cou droit",
		Description:      "Etirement du cou vers la droite, tete basculée vers la gauche, la main droite sur l'oreille droite",
		CategoryID:       CategoryIdNeck,
		TargetRepetition: 3,
	}

	CategoryNeck = Category{
		ID:          CategoryIdNeck,
		Name:        "Cou",
		Description: "Etirement du cou et de la tête",
		Drills: map[string]Drill{
			DrillNeckFront.ID: DrillNeckFront,
			DrillNeckLeft.ID:  DrillNeckLeft,
			DrillNeckRight.ID: DrillNeckRight,
		},
	}
)
