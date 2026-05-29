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

type CreateSuppressedFindingParams struct {
	ApplicationID     string
	DesiredSnapshotID string
	LiveSnapshotID    string
	IgnoreRuleID      string
	IgnoreRuleName    string
	IgnoreRuleReason  string
	Title             string
	Category          string
	Severity          string
	ResourceRef       string
	FieldPath         string
	DesiredValue      string
	LiveValue         string
	DifferenceType    string
}

type CreateIgnoreRuleParams struct {
	WorkspaceID     string
	ApplicationID   string
	ResourceRef     string
	Name            string
	MatchExpression string
	Reason          string
	CreatedBy       string
}

type UpdateIgnoreRuleParams struct {
	ResourceRef     string
	Name            string
	MatchExpression string
	Reason          string
}

// UpsertIncidentParams provides the context needed to either create a new drift incident
// or update an existing open one for the same resource and field.
type UpsertIncidentParams struct {
	ApplicationID     string
	DesiredSnapshotID string
	LiveSnapshotID    string
	Finding           DriftFinding
	RecommendedAction string
}

// DriftFinding represents a single point of drift detected during analysis.
type DriftFinding struct {
	Title          string
	Category       string
	Severity       string
	Confidence     float64
	ResourceRef    string
	FieldPath      string
	DesiredValue   string
	LiveValue      string
	DifferenceType string
}

type UpsertGitHubEventParams struct {
	EventType  string
	DeliveryID string
	Action     string
	Repository string
	Payload    string
}
