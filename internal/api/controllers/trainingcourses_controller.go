package controllers

import (
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
)

type TrainingCoursesController struct {
	service svc.Service[mdl.TrainingBadges]
}

func NewTrainingCoursesController(service svc.Service[mdl.TrainingBadges]) Controller[mdl.TrainingBadges] {
	return NewController[mdl.TrainingBadges](nil, service)
}
