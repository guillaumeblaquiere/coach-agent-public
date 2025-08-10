package main

import "coach/models"

const (
	CategoryIdShoulder = "shoulder"
)

var (
	DrillTricepsLeft = models.Drill{
		ID: "TricepsLeft",
		Name: map[string]string{
			"fr-FR": "Triceps gauche",
			"en-EN": "Left Triceps",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du triceps gauche, coude au dessus de la tete, l'autre main tire le coude",
			"en-EN": "Stretching the left triceps, elbow above the head, the other hand pulls the elbow",
		},
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillTricepsRight = models.Drill{
		ID: "TricepsRight",
		Name: map[string]string{
			"fr-FR": "Triceps droit",
			"en-EN": "Right Triceps",
		},
		Description: map[string]string{
			"fr-FR": "Etirement du triceps droit, coude au dessus de la tete, l'autre main tire le coude",
			"en-EN": "Stretching the right triceps, elbow above the head, the other hand pulls the elbow",
		},
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillShoulderCrossLeft = models.Drill{
		ID: "ShoulderCrossLeft",
		Name: map[string]string{
			"fr-FR": "Epaule croisée gauche",
			"en-EN": "Left Cross Shoulder",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'épaule gauche, bras gauche tendu devant soi allant vers la droite, l'autre main tire le bras",
			"en-EN": "Stretching the left shoulder, left arm extended in front of you going to the right, the other hand pulls the arm",
		},
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillShoulderCrossRight = models.Drill{
		ID: "ShoulderCrossRight",
		Name: map[string]string{
			"fr-FR": "Epaule croisée droite",
			"en-EN": "Right Cross Shoulder",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'épaule droite, bras droit tendu devant soi allant vers la gauche, l'autre main tire le bras",
			"en-EN": "Stretching the right shoulder, right arm extended in front of you going to the left, the other hand pulls the arm",
		},
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillShoulderCrossRightBack = models.Drill{
		ID: "ShoulderCrossRightBack",
		Name: map[string]string{
			"fr-FR": "Epaule croisée droite arrière",
			"en-EN": "Right Back Cross Shoulder",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'épaule droite, bras droit tendu appuyé sur un montant fixe (porte, poteau) vers l'arrière. Le corps effectue une rotation pour étirer l'épaule",
			"en-EN": "Stretching the right shoulder, right arm extended resting on a fixed upright (door, post) towards the back. The body rotates to stretch the shoulder",
		},
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillShoulderCrossLeftBack = models.Drill{
		ID: "ShoulderCrossLeftBack",
		Name: map[string]string{
			"fr-FR": "Epaule croisée gauche arrière",
			"en-EN": "Left Back Cross Shoulder",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'épaule gauche, bras gauche tendu appuyé sur un montant fixe (porte, poteau) vers l'arrière. Le corps effectue une rotation pour étirer l'épaule",
			"en-EN": "Stretching the left shoulder, left arm extended resting on a fixed upright (door, post) towards the back. The body rotates to stretch the shoulder",
		},
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillObliqueLeft = models.Drill{
		ID: "ObliqueLeft",
		Name: map[string]string{
			"fr-FR": "Oblique gauche",
			"en-EN": "Left Oblique",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'oblique gauche, bras gauche le long du corps, bras droit au dessus de la tete, penché vers la gauche",
			"en-EN": "Stretching the left oblique, left arm along the body, right arm above the head, leaning to the left",
		},
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	DrillObliqueRight = models.Drill{
		ID: "ObliqueRight",
		Name: map[string]string{
			"fr-FR": "Oblique droit",
			"en-EN": "Right Oblique",
		},
		Description: map[string]string{
			"fr-FR": "Etirement de l'oblique droit, bras droit le long du corps, bras gauche au dessus de la tete, penché vers la droite",
			"en-EN": "Stretching the right oblique, right arm along the body, left arm above the head, leaning to the right",
		},
		CategoryID:       CategoryIdShoulder,
		TargetRepetition: 3,
	}

	CategoryShoulder = models.Category{
		ID: CategoryIdShoulder,
		Name: map[string]string{
			"fr-FR": "Epaule",
			"en-EN": "Shoulder",
		},
		Description: map[string]string{
			"fr-FR": "Etirement des épaules et triceps",
			"en-EN": "Stretching of the shoulders and triceps",
		},
		Drills: map[string]models.Drill{
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
