package services

import (
	"github.com/ko1N/zeebe-video-service/internal/environment"
)

// ServiceContext - context for a services
type ServiceContext struct {
	Environment environment.Environment
	Tracker     *Tracker
}

func NewServiceContext(env environment.Environment, tracker *Tracker) *ServiceContext {
	return &ServiceContext{
		Environment: env,
		Tracker:     tracker,
	}
}
