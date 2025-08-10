package main

import "coach/models"

const (
	CategoryIdLeg = "leg"
)

var (
	DrillHamstringLeft = models.Drill{
		ID: "HamstringLeft",
		Name: map[string]string{
			"fr-FR": "Ischio Gauche",
			"en-EN": "Left Hamstring",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'ischio gauche, position debout la jambe sur-élevée, pointe du pied vers le haut",
			"en-EN": "Stretching of the left hamstring, standing position with the leg elevated, toe pointing upwards",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillHamstringRight = models.Drill{
		ID: "HamstringRight",
		Name: map[string]string{
			"fr-FR": "Ischio Droit",
			"en-EN": "Right Hamstring",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'ischio droit, position debout la jambe sur-élevée, pointe du pied vers le haut",
			"en-EN": "Stretching of the right hamstring, standing position with the leg elevated, toe pointing upwards",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillHamstringBoth = models.Drill{
		ID: "HamstringBoth",
		Name: map[string]string{
			"fr-FR": "Ischios",
			"en-EN": "Hamstrings",
		},
		Description: map[string]string{
			"fr-FR": "Etirement des ischios, position assise au sol, jambes tendues, corps basculé en avant",
			"en-EN": "Stretching of the hamstrings, sitting on the ground, legs extended, body tilted forward",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillQuadricepsLeft = models.Drill{
		ID: "QuadricepsLeft",
		Name: map[string]string{
			"fr-FR": "Quadriceps gauche",
			"en-EN": "Left Quadriceps",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du quadriceps gauche, position debout, jambe pliée arrière",
			"en-EN": "Stretching of the left quadriceps, standing position, leg bent back",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillQuadricepsRight = models.Drill{
		ID: "QuadricepsRight",
		Name: map[string]string{
			"fr-FR": "Quadriceps droit",
			"en-EN": "Right Quadriceps",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du quadriceps droit, position debout, jambe pliée arrière",
			"en-EN": "Stretching of the right quadriceps, standing position, leg bent back",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillCalfLeft = models.Drill{
		ID: "CalfLeft",
		Name: map[string]string{
			"fr-FR": "Mollet Gauche",
			"en-EN": "Left Calf",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du mollet gauche, position debout sur une marche",
			"en-EN": "Stretching of the left calf, standing position on a step",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillCalfRight = models.Drill{
		ID: "CalfRight",
		Name: map[string]string{
			"fr-FR": "Mollet Droit",
			"en-EN": "Right Calf",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du mollet droit, position debout sur une marche",
			"en-EN": "Stretching of the right calf, standing position on a step",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillAdductorSplits = models.Drill{
		ID: "AdductorSplits",
		Name: map[string]string{
			"fr-FR": "Adducteur grand écart",
			"en-EN": "Adductor Splits",
		},
		Description: map[string]string{
			"fr-FR": "Etirement des adducteurs, position assise au sol, jambes écartées, corps basculé en avant",
			"en-EN": "Stretching of the adductors, sitting on the ground, legs apart, body tilted forward",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillAdductorButterfly = models.Drill{
		ID: "AdductorButterfly",
		Name: map[string]string{
			"fr-FR": "Adducteur papillon",
			"en-EN": "Adductor Butterfly",
		},
		Description: map[string]string{
			"fr-FR": "Etirement des adducteurs, position assise au sol, jambes pliées, plantes des pieds jointes",
			"en-EN": "Stretching of the adductors, sitting on the ground, legs bent, soles of the feet together",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillAdductorLeft = models.Drill{
		ID: "AdductorLeft",
		Name: map[string]string{
			"fr-FR": "Adducteur gauche",
			"en-EN": "Left Adductor",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'adducteur gauche, position assise, jambe gauche pliée, jambe droite tendue",
			"en-EN": "Stretching of the left adductor, sitting position, left leg bent, right leg extended",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillAdductorRight = models.Drill{
		ID: "AdductorRight",
		Name: map[string]string{
			"fr-FR": "Adducteur droit",
			"en-EN": "Right Adductor",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'adducteur droit, position assise, jambe droite pliée, jambe gauche tendue",
			"en-EN": "Stretching of the right adductor, sitting position, right leg bent, left leg extended",
		},
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	CategoryLeg = models.Category{
		ID: CategoryIdLeg,
		Name: map[string]string{
			"fr-FR": "Jambe",
			"en-EN": "Leg",
		},
		Description: map[string]string{
			"fr-FR": "Etirement bas du corps et des jambes, position assise et debout",
			"en-EN": "Stretching of the lower body and legs, sitting and standing position",
		},
		Drills: map[string]models.Drill{
			DrillHamstringLeft.ID:     DrillHamstringLeft,
			DrillHamstringRight.ID:    DrillHamstringRight,
			DrillHamstringBoth.ID:     DrillHamstringBoth,
			DrillQuadricepsLeft.ID:    DrillQuadricepsLeft,
			DrillQuadricepsRight.ID:   DrillQuadricepsRight,
			DrillCalfLeft.ID:          DrillCalfLeft,
			DrillCalfRight.ID:         DrillCalfRight,
			DrillAdductorSplits.ID:    DrillAdductorSplits,
			DrillAdductorButterfly.ID: DrillAdductorButterfly,
			DrillAdductorLeft.ID:      DrillAdductorLeft,
			DrillAdductorRight.ID:     DrillAdductorRight,
		},
	}
)
