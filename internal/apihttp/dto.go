package apihttp

import "github.com/AnouarMohamed/StateSight/pkg/model"

type createApplicationRequest struct {
	WorkspaceID        string `json:"workspace_id"`
	ClusterID          string `json:"cluster_id"`
	SourceDefinitionID string `json:"source_definition_id"`
	Name               string `json:"name"`
	Namespace          string `json:"namespace"`
}

type createIgnoreRuleRequest struct {
	Name            string `json:"name"`
	MatchExpression string `json:"match_expression"`
	ResourceRef     string `json:"resource_ref"`
	Reason          string `json:"reason"`
}

type updateIgnoreRuleRequest struct {
	Active *bool `json:"active"`
}

type applicationDetailsResponse struct {
	Application  model.Application         `json:"application"`
	Incidents    []model.DriftIncident     `json:"incidents"`
	Suppressions []model.SuppressedFinding `json:"suppressions"`
	IgnoreRules  []model.IgnoreRule        `json:"ignore_rules"`
}

type analyzeResponse struct {
	JobID         string `json:"job_id"`
	JobType       string `json:"job_type"`
	Status        string `json:"status"`
	ApplicationID string `json:"application_id"`
}
