package main

import "coach/models"

var (
	Categories = map[string]models.Category{
		CategoryIdBack:     CategoryBack,
		CategoryIdShoulder: CategoryShoulder,
		CategoryIdNeck:     CategoryNeck,
		CategoryIdLeg:      CategoryLeg,
	}
)
