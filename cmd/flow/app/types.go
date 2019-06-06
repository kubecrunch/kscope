package app

import (
	"github.com/kubecrunch/kscope/api/v1alpha1"
)

// FlowConfiguration
type FlowConfiguration struct {
	Stages []v1alpha1.KscopeStage `json:"stages"`
}
