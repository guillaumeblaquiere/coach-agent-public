package main

const (
	CategoryIdBack = "back"
)

var (
	DrillHipStretchedLeft = Drill{
		ID:               "HipStretchedLeft",
		Name:             "Etirement Hanche Gauche",
		Description:      "Position de fente au sol, jambe gauche pliée sur le coté, jambe droite tendue, genou au sol, bassin basculé à l'avant",
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillHipStretchedRight = Drill{
		ID:               "HipStretchedRight",
		Name:             "Etirement Hanche Droit",
		Description:      "Position de fente au sol, jambe droite pliée sur le coté, jambe gauche tendue, genou au sol, bassin basculé à l'avant",
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillLumbarRotationLeft = Drill{
		ID:               "LumbarRotationLeft",
		Name:             "Rotation Lombaire Gauche",
		Description:      "Position au sol, jambe gauche pliée, basculée sur le coté droit, jambe droite tendue, épaules au sol",
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillLumbarRotationRight = Drill{
		ID:               "LumbarRotationRight",
		Name:             "Rotation Lombaire Droit",
		Description:      "Position au sol, jambe droite pliée, basculée sur le coté gauche, jambe gauche tendue, épaules au sol",
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillLumbarStretched = Drill{
		ID:               "LumbarStretched",
		Name:             "Etirement Lombaire",
		Description:      "Position au sol, genoux sur la poitrine, rouler",
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillGluteusLeft = Drill{
		ID:               "GluteusLeft",
		Name:             "Fessier Gauche",
		Description:      "Position au sol, jambe gauche pliée, la basculer vers la poitrine, jambe droite tendue",
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillGluteusRight = Drill{
		ID:               "GluteusRight",
		Name:             "Fessier Droit",
		Description:      "Position au sol, jambe droite pliée, la basculer vers la poitrine, jambe gauche tendue",
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillAbs = Drill{
		ID:               "Abs",
		Name:             "Etirement abdominaux",
		Description:      "Position au allongé face au sol, bras tendu, tête vers le haut",
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillUpperBack = Drill{
		ID:               "UpperBack",
		Name:             "Etirement haut du dos",
		Description:      "Position au sol, dos arrondit, bras tendus à l'avant",
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	CategoryBack = Category{
		ID:          CategoryIdBack,
		Name:        "Dos",
		Description: "Etirement des hanches et du dos, position au sol",
		Drills: map[string]Drill{
			DrillHipStretchedLeft.ID:    DrillHipStretchedLeft,
			DrillHipStretchedRight.ID:   DrillHipStretchedRight,
			DrillLumbarRotationLeft.ID:  DrillLumbarRotationLeft,
			DrillLumbarRotationRight.ID: DrillLumbarRotationRight,
			DrillLumbarStretched.ID:     DrillLumbarStretched,
			DrillGluteusLeft.ID:         DrillGluteusLeft,
			DrillGluteusRight.ID:        DrillGluteusRight,
			DrillAbs.ID:                 DrillAbs,
			DrillUpperBack.ID:           DrillUpperBack,
		},
	}
)
