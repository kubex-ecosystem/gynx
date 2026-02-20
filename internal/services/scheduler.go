package services

import "github.com/kubex-ecosystem/kbx/types"

type SchedulerService struct {
	cfg *types.SrvConfig
	// Add fields for managing scheduled tasks, e.g., a task queue, worker pool, etc.
}

func NewSchedulerService(cfg *types.SrvConfig) *SchedulerService {
	return &SchedulerService{
		cfg: cfg,
	}
}
