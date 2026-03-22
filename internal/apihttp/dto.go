package apihttp

import "github.com/AnouarMohamed/StateSight/pkg/model"

type createApplicationRequest struct {
	WorkspaceID        string `json:"workspace_id"`
	ClusterID          string `json:"cluster_id"`
	SourceDefinitionID string `json:"source_definition_id"`
	Name               string `json:"name"`
	Namespace          string `json:"namespace"`
}

type applicationDetailsResponse struct {
	Application model.Application     `json:"application"`
	Incidents   []model.DriftIncident `json:"incidents"`
}

type analyzeResponse struct {
	JobID         string `json:"job_id"`
	JobType       string `json:"job_type"`
	Status        string `json:"status"`
	ApplicationID string `json:"application_id"`
}
