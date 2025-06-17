package main

const (
	CategoryIdShoulder = "shoulder"
)

var (
	DrillTricepsLeft = Drill{
		ID:               "TricepsLeft",
		Name:             "Triceps gauche",
		Description:      "Etirement du triceps gauche, coude au dessus de la tete, l'autre main tire le coude",
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillTricepsRight = Drill{
		ID:               "TricepsRight",
		Name:             "Triceps droit",
		Description:      "Etirement du triceps droit, coude au dessus de la tete, l'autre main tire le coude",
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillShoulderCrossLeft = Drill{
		ID:               "ShoulderCrossLeft",
		Name:             "Epaule croisée gauche",
		Description:      "Etirement de l'épaule gauche, bras gauche tendu devant soi allant vers la droite, l'autre main tire le bras",
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillShoulderCrossRight = Drill{
		ID:               "ShoulderCrossRight",
		Name:             "Epaule croisée droite",
		Description:      "Etirement de l'épaule droite, bras droit tendu devant soi allant vers la gauche, l'autre main tire le bras",
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillShoulderCrossRightBack = Drill{
		ID:               "ShoulderCrossRightBack",
		Name:             "Epaule croisée droite arrière",
		Description:      "Etirement de l'épaule droite, bras droit tendu appuyé sur un montant fixe (porte, poteau) vers l'arrière. Le corps effectue une rotation pour étirer l'épaule",
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillShoulderCrossLeftBack = Drill{
		ID:               "ShoulderCrossLeftBack",
		Name:             "Epaule croisée gauche arrière",
		Description:      "Etirement de l'épaule gauche, bras gauche tendu appuyé sur un montant fixe (porte, poteau) vers l'arrière. Le corps effectue une rotation pour étirer l'épaule",
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillObliqueLeft = Drill{
		ID:               "ObliqueLeft",
		Name:             "Oblique gauche",
		Description:      "Etirement de l'oblique gauche, bras gauche le long du corps, bras droit au dessus de la tete, penché vers la gauche",
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillObliqueRight = Drill{
		ID:               "ObliqueRight",
		Name:             "Oblique droit",
		Description:      "Etirement de l'oblique droit, bras droit le long du corps, bras gauche au dessus de la tete, penché vers la droite",
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	CategoryShoulder = Category{
		ID:          CategoryIdShoulder,
		Name:        "Epaule",
		Description: "Etirement des épaules et triceps",
		Drills: map[string]Drill{
			DrillTricepsLeft.ID:            DrillTricepsLeft,
			DrillTricepsRight.ID:           DrillTricepsRight,
			DrillShoulderCrossLeft.ID:      DrillShoulderCrossLeft,
			DrillShoulderCrossRight.ID:     DrillShoulderCrossRight,
			DrillShoulderCrossRightBack.ID: DrillShoulderCrossRightBack,
			DrillShoulderCrossLeftBack.ID:  DrillShoulderCrossLeftBack,
			DrillObliqueLeft.ID:            DrillObliqueLeft,
			DrillObliqueRight.ID:           DrillObliqueRight,
		},
	}
)
