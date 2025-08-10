package main

import "coach/models"

const (
	CategoryIdBack = "back"
)

var (
	DrillHipStretchedLeft = models.Drill{
		ID: "HipStretchedLeft",
		Name: map[string]string{
			"fr-FR": "Etirement Hanche Gauche",
			"en-EN": "Left Hip Stretch",
		},
		Description: map[string]string{
			"fr-FR": "Position de fente au sol, jambe gauche pliée sur le coté, jambe droite tendue, genou au sol, bassin basculé à l'avant",
			"en-EN": "Ground lunge position, left leg bent to the side, right leg extended, knee on the ground, pelvis tilted forward",
		},
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillHipStretchedRight = models.Drill{
		ID: "HipStretchedRight",
		Name: map[string]string{
			"fr-FR": "Etirement Hanche Droit",
			"en-EN": "Right Hip Stretch",
		},
		Description: map[string]string{
			"fr-FR": "Position de fente au sol, jambe droite pliée sur le coté, jambe gauche tendue, genou au sol, bassin basculé à l'avant",
			"en-EN": "Ground lunge position, right leg bent to the side, left leg extended, knee on the ground, pelvis tilted forward",
		},
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillLumbarRotationLeft = models.Drill{
		ID: "LumbarRotationLeft",
		Name: map[string]string{
			"fr-FR": "Rotation Lombaire Gauche",
			"en-EN": "Left Lumbar Rotation",
		},
		Description: map[string]string{
			"fr-FR": "Position au sol, jambe gauche pliée, basculée sur le coté droit, jambe droite tendue, épaules au sol",
			"en-EN": "Ground position, left leg bent, tilted to the right side, right leg extended, shoulders on the ground",
		},
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillLumbarRotationRight = models.Drill{
		ID: "LumbarRotationRight",
		Name: map[string]string{
			"fr-FR": "Rotation Lombaire Droit",
			"en-EN": "Right Lumbar Rotation",
		},
		Description: map[string]string{
			"fr-FR": "Position au sol, jambe droite pliée, basculée sur le coté gauche, jambe gauche tendue, épaules au sol",
			"en-EN": "Ground position, right leg bent, tilted to the left side, left leg extended, shoulders on the ground",
		},
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillLumbarStretched = models.Drill{
		ID: "LumbarStretched",
		Name: map[string]string{
			"fr-FR": "Etirement Lombaire",
			"en-EN": "Lumbar Stretch",
		},
		Description: map[string]string{
			"fr-FR": "Position au sol, genoux sur la poitrine, rouler",
			"en-EN": "Ground position, knees to chest, roll",
		},
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillGluteusLeft = models.Drill{
		ID: "GluteusLeft",
		Name: map[string]string{
			"fr-FR": "Fessier Gauche",
			"en-EN": "Left Gluteus",
		},
		Description: map[string]string{
			"fr-FR": "Position au sol, jambe gauche pliée, la basculer vers la poitrine, jambe droite tendue",
			"en-EN": "Ground position, left leg bent, bring it towards the chest, right leg extended",
		},
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillGluteusRight = models.Drill{
		ID: "GluteusRight",
		Name: map[string]string{
			"fr-FR": "Fessier Droit",
			"en-EN": "Right Gluteus",
		},
		Description: map[string]string{
			"fr-FR": "Position au sol, jambe droite pliée, la basculer vers la poitrine, jambe gauche tendue",
			"en-EN": "Ground position, right leg bent, bring it towards the chest, left leg extended",
		},
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillAbs = models.Drill{
		ID: "Abs",
		Name: map[string]string{
			"fr-FR": "Etirement abdominaux",
			"en-EN": "Abdominal Stretch",
		},
		Description: map[string]string{
			"fr-FR": "Position au allongé face au sol, bras tendu, tête vers le haut",
			"en-EN": "Lying face down, arms extended, head up",
		},
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	DrillUpperBack = models.Drill{
		ID: "UpperBack",
		Name: map[string]string{
			"fr-FR": "Etirement haut du dos",
			"en-EN": "Upper Back Stretch",
		},
		Description: map[string]string{
			"fr-FR": "Position au sol, dos arrondit, bras tendus à l'avant",
			"en-EN": "Ground position, back rounded, arms extended forward",
		},
		CategoryID:       CategoryIdBack,
		TargetRepetition: 3,
	}

	CategoryBack = models.Category{
		ID: CategoryIdBack,
		Name: map[string]string{
			"fr-FR": "Dos",
			"en-EN": "Back",
		},
		Description: map[string]string{
			"fr-FR": "Etirement des hanches et du dos, position au sol",
			"en-EN": "Stretching of the hips and back, ground position",
		},
		Drills: map[string]models.Drill{
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
