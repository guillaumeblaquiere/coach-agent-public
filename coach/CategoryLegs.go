package main

const (
	CategoryIdLeg = "leg"
)

var (
	DrillHamstringLeft = Drill{
		ID:               "HamstringLeft",
		Name:             "Ischio Gauche",
		Description:      "Etirement de l'ischio gauche, position debout la jambe sur-élevée, pointe du pied vers le haut",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillHamstringRight = Drill{
		ID:               "HamstringRight",
		Name:             "Ischio Droit",
		Description:      "Etirement de l'ischio droit, position debout la jambe sur-élevée, pointe du pied vers le haut",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillHamstringBoth = Drill{
		ID:               "HamstringBoth",
		Name:             "Ischios",
		Description:      "Etirement des ischios, position assise au sol, jambes tendues, corps basculé en avant",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillQuadricepsLeft = Drill{
		ID:               "QuadricepsLeft",
		Name:             "Quadriceps gauche",
		Description:      "Etirement du quadriceps gauche, position debout, jambe pliée arrière",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillQuadricepsRight = Drill{
		ID:               "QuadricepsRight",
		Name:             "Quadriceps droit",
		Description:      "Etirement du quadriceps droit, position debout, jambe pliée arrière",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillCalfLeft = Drill{
		ID:               "CalfLeft",
		Name:             "Mollet Gauche",
		Description:      "Etirement du mollet gauche, position debout sur une marche",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillCalfRight = Drill{
		ID:               "CalfRight",
		Name:             "Mollet Droit",
		Description:      "Etirement du mollet droit, position debout sur une marche",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillAdductorSplits = Drill{
		ID:               "AdductorSplits",
		Name:             "Adducteur grand écart",
		Description:      "Etirement des adducteurs, position assise au sol, jambes écartées, corps basculé en avant",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillAdductorButterfly = Drill{
		ID:               "AdductorButterfly",
		Name:             "Adducteur papillon",
		Description:      "Etirement des adducteurs, position assise au sol, jambes pliées, plantes des pieds jointes",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillAdductorLeft = Drill{
		ID:               "AdductorLeft",
		Name:             "Adducteur gauche",
		Description:      "Etirement de l'adducteur gauche, position assise, jambe gauche pliée, jambe droite tendue",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	DrillAdductorRight = Drill{
		ID:               "AdductorRight",
		Name:             "Adducteur droit",
		Description:      "Etirement de l'adducteur droit, position assise, jambe droite pliée, jambe gauche tendue",
		CategoryID:       CategoryIdLeg,
		TargetRepetition: 3,
	}

	CategoryLeg = Category{
		ID:          CategoryIdLeg,
		Name:        "Jambe",
		Description: "Etirement bas du corps et des jambes, position assise et debout",
		Drills: map[string]Drill{
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
