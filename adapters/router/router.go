// Package router implements read-only HTTP health signals for agent routers.
package router

import (
	"context"
	httpadapter "github.com/TheAgentHealth/agenthealth/adapters/http"
	"github.com/TheAgentHealth/agenthealth/core"
)

type Adapter struct{}

func (Adapter) Metadata() core.Metadata {
	m := (httpadapter.Adapter{}).Metadata()
	m.Name = "agent-router"
	m.Version = "0.7.0"
	m.TargetTypes = []string{"router"}
	return m
}
func (Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	return (httpadapter.Adapter{ReadinessGET: true}).Check(ctx, r, dimension)
}
