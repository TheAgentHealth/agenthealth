// Package gateway implements read-only Agentgateway HTTP health signals.
package gateway

import (
	"context"
	httpadapter "github.com/TheAgentHealth/agenthealth/adapters/http"
	"github.com/TheAgentHealth/agenthealth/core"
)

type Adapter struct{}

func (Adapter) Metadata() core.Metadata {
	m := (httpadapter.Adapter{}).Metadata()
	m.Name = "agentgateway"
	m.Version = "0.6.0"
	m.TargetTypes = []string{"gateway"}
	return m
}
func (Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	return (httpadapter.Adapter{ReadinessGET: true}).Check(ctx, r, dimension)
}
