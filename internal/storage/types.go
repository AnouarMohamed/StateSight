package storage

type CreateApplicationParams struct {
	WorkspaceID        string
	ClusterID          string
	SourceDefinitionID string
	Name               string
	Namespace          string
}

type CreateJobParams struct {
	JobType       string
	ApplicationID *string
	PayloadJSON   string
}

type CreateDesiredSnapshotParams struct {
	ApplicationID string
	Revision      string
	SummaryJSON   string
}

type CreateLiveSnapshotParams struct {
	ApplicationID string
	SummaryJSON   string
}

type CreateIncidentParams struct {
	ApplicationID     string
	DesiredSnapshotID string
	LiveSnapshotID    string
	Title             string
	Category          string
	Severity          string
	Confidence        float64
	RecommendedAction string
	Status            string
}

type CreateDriftFieldParams struct {
	IncidentID     string
	ResourceRef    string
	FieldPath      string
	DesiredValue   string
	LiveValue      string
	DifferenceType string
}

type CreateEvidenceRecordParams struct {
	IncidentID string
	Source     string
	Detail     string
	Actor      string
	Confidence float64
	Metadata   string
}

type UpsertGitHubEventParams struct {
	EventType  string
	DeliveryID string
	Action     string
	Repository string
	Payload    string
}
