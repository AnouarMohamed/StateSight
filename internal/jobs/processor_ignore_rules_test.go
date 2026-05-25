package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/AnouarMohamed/StateSight/internal/k8scollect"
	"github.com/AnouarMohamed/StateSight/internal/sourceingest"
	"github.com/AnouarMohamed/StateSight/internal/storage"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func TestProcessAnalyzeSuppressesCandidateMatchingEffectiveIgnoreRule(t *testing.T) {
	store := &ignoreRuleAnalysisStore{
		app: model.Application{
			ID:                 "application-1",
			WorkspaceID:        "workspace-1",
			SourceDefinitionID: "source-1",
			ClusterID:          "cluster-1",
			Name:               "ledger-api",
			Namespace:          "payments",
		},
		rules: []model.IgnoreRule{{
			ID:              "rule-1",
			ApplicationID:   "application-1",
			ResourceRef:     "apps/v1/Deployment:payments/ledger-api",
			Name:            "Allow HPA replica variance",
			MatchExpression: "spec.replicas",
			Reason:          "replicas are managed outside Git",
			Active:          true,
		}},
	}

	processor := NewProcessor(store, slog.New(slog.NewTextHandler(io.Discard, nil)), ProcessorOptions{})
	processor.fetcher = staticDesiredStateFetcher{}
	processor.collector = staticLiveStateCollector{}

	if err := processor.processAnalyze(context.Background(), Message{ApplicationID: store.app.ID}); err != nil {
		t.Fatalf("process analysis: %v", err)
	}

	if store.rulesWorkspaceID != store.app.WorkspaceID || store.rulesApplicationID != store.app.ID {
		t.Fatalf("expected rule lookup for workspace/application %q/%q, got %q/%q",
			store.app.WorkspaceID, store.app.ID, store.rulesWorkspaceID, store.rulesApplicationID)
	}
	if store.rulesQueries != 1 {
		t.Fatalf("expected one rule query per analysis, got %d", store.rulesQueries)
	}
	if store.incidentsCreated != 0 {
		t.Fatalf("expected matching candidate to be suppressed, created %d incidents", store.incidentsCreated)
	}
	if len(store.suppressedFindings) != 1 {
		t.Fatalf("expected one suppression audit record, got %d", len(store.suppressedFindings))
	}
	suppressed := store.suppressedFindings[0]
	if suppressed.IgnoreRuleID != "rule-1" || suppressed.FieldPath != "spec.replicas" {
		t.Fatalf("unexpected suppression audit payload: %#v", suppressed)
	}
	if suppressed.DesiredSnapshotID != "desired-1" || suppressed.LiveSnapshotID != "live-1" {
		t.Fatalf("expected suppression record linked to analysis snapshots, got %#v", suppressed)
	}
	if store.desiredSnapshotsCreated != 1 || store.liveSnapshotsCreated != 1 {
		t.Fatal("expected analysis snapshots to be persisted before candidate suppression")
	}
}

func TestProcessAnalyzeSkipsIgnoreRulesWhenThereAreNoCandidates(t *testing.T) {
	store := &ignoreRuleAnalysisStore{
		app: model.Application{
			ID:                 "application-1",
			WorkspaceID:        "workspace-1",
			SourceDefinitionID: "source-1",
			ClusterID:          "cluster-1",
			Name:               "ledger-api",
			Namespace:          "payments",
		},
	}

	processor := NewProcessor(store, slog.New(slog.NewTextHandler(io.Discard, nil)), ProcessorOptions{})
	processor.fetcher = staticDesiredStateFetcher{}
	processor.collector = unchangedLiveStateCollector{}

	if err := processor.processAnalyze(context.Background(), Message{ApplicationID: store.app.ID}); err != nil {
		t.Fatalf("process analysis: %v", err)
	}

	if store.rulesQueries != 0 {
		t.Fatalf("expected no rule query without candidates, got %d", store.rulesQueries)
	}
	if store.incidentsCreated != 0 {
		t.Fatalf("expected no incidents without drift, created %d", store.incidentsCreated)
	}
	if len(store.suppressedFindings) != 0 {
		t.Fatalf("expected no suppressions without drift, got %d", len(store.suppressedFindings))
	}
}

func TestProcessAnalyzePersistsCapturedProvenanceEvidence(t *testing.T) {
	store := &ignoreRuleAnalysisStore{
		app: model.Application{
			ID:                 "application-1",
			WorkspaceID:        "workspace-1",
			SourceDefinitionID: "source-1",
			ClusterID:          "cluster-1",
			Name:               "ledger-api",
			Namespace:          "payments",
		},
		source: model.SourceDefinition{
			RepoURL:       "https://github.com/example/platform-config",
			DefaultBranch: "main",
			Path:          "clusters/prod",
		},
		cluster: model.Cluster{ID: "cluster-1", Name: "prod-eu"},
	}

	processor := NewProcessor(store, slog.New(slog.NewTextHandler(io.Discard, nil)), ProcessorOptions{})
	processor.fetcher = staticDesiredStateFetcher{}
	processor.collector = staticLiveStateCollector{}

	if err := processor.processAnalyze(context.Background(), Message{ApplicationID: store.app.ID}); err != nil {
		t.Fatalf("process analysis: %v", err)
	}
	if store.incidentsCreated != 1 {
		t.Fatalf("expected one incident for unsuppressed drift, got %d", store.incidentsCreated)
	}
	if len(store.evidenceRecords) != 2 {
		t.Fatalf("expected Git and live evidence records, got %#v", store.evidenceRecords)
	}
	if store.evidenceRecords[0].Source != "git" || store.evidenceRecords[0].Actor != "not-attributed" {
		t.Fatalf("unexpected Git evidence: %#v", store.evidenceRecords[0])
	}
	var gitMetadata map[string]any
	if err := json.Unmarshal([]byte(store.evidenceRecords[0].Metadata), &gitMetadata); err != nil {
		t.Fatalf("decode Git metadata: %v", err)
	}
	if gitMetadata["revision"] != "revision-1" || gitMetadata["repository"] != store.source.RepoURL {
		t.Fatalf("expected persisted Git provenance, got %#v", gitMetadata)
	}
	if store.evidenceRecords[1].Source != "kubectl" || store.evidenceRecords[1].Actor != "not-attributed" {
		t.Fatalf("unexpected live evidence: %#v", store.evidenceRecords[1])
	}
}

func TestProcessAnalyzeFailsWhenSuppressionCannotBeAudited(t *testing.T) {
	store := &ignoreRuleAnalysisStore{
		app: model.Application{
			ID:                 "application-1",
			WorkspaceID:        "workspace-1",
			SourceDefinitionID: "source-1",
			ClusterID:          "cluster-1",
			Name:               "ledger-api",
			Namespace:          "payments",
		},
		rules: []model.IgnoreRule{{
			ID:              "rule-1",
			MatchExpression: "spec.replicas",
			Active:          true,
		}},
		suppressionErr: errors.New("storage unavailable"),
	}

	processor := NewProcessor(store, slog.New(slog.NewTextHandler(io.Discard, nil)), ProcessorOptions{})
	processor.fetcher = staticDesiredStateFetcher{}
	processor.collector = staticLiveStateCollector{}

	err := processor.processAnalyze(context.Background(), Message{ApplicationID: store.app.ID})
	if err == nil {
		t.Fatal("expected analysis to fail when suppression audit cannot be persisted")
	}
	if store.incidentsCreated != 0 {
		t.Fatalf("expected failed audit not to create incident, created %d", store.incidentsCreated)
	}
}

type ignoreRuleAnalysisStore struct {
	app                     model.Application
	source                  model.SourceDefinition
	cluster                 model.Cluster
	rules                   []model.IgnoreRule
	rulesWorkspaceID        string
	rulesApplicationID      string
	rulesQueries            int
	desiredSnapshotsCreated int
	liveSnapshotsCreated    int
	incidentsCreated        int
	evidenceRecords         []storage.CreateEvidenceRecordParams
	suppressedFindings      []storage.CreateSuppressedFindingParams
	suppressionErr          error
}

func (*ignoreRuleAnalysisStore) MarkJobProcessing(context.Context, string) error { return nil }
func (*ignoreRuleAnalysisStore) MarkJobCompleted(context.Context, string) error  { return nil }
func (*ignoreRuleAnalysisStore) MarkJobFailed(context.Context, string, string) error {
	return nil
}

func (s *ignoreRuleAnalysisStore) GetApplicationByID(context.Context, string) (model.Application, error) {
	return s.app, nil
}

func (s *ignoreRuleAnalysisStore) GetSourceDefinitionByID(context.Context, string) (model.SourceDefinition, error) {
	return s.source, nil
}

func (s *ignoreRuleAnalysisStore) GetClusterByID(context.Context, string) (model.Cluster, error) {
	return s.cluster, nil
}

func (s *ignoreRuleAnalysisStore) ListActiveIgnoreRulesForAnalysis(_ context.Context, workspaceID, applicationID string) ([]model.IgnoreRule, error) {
	s.rulesQueries++
	s.rulesWorkspaceID = workspaceID
	s.rulesApplicationID = applicationID
	return s.rules, nil
}

func (s *ignoreRuleAnalysisStore) CreateDesiredSnapshot(context.Context, storage.CreateDesiredSnapshotParams) (model.DesiredSnapshot, error) {
	s.desiredSnapshotsCreated++
	return model.DesiredSnapshot{ID: "desired-1"}, nil
}

func (s *ignoreRuleAnalysisStore) CreateLiveSnapshot(context.Context, storage.CreateLiveSnapshotParams) (model.LiveSnapshot, error) {
	s.liveSnapshotsCreated++
	return model.LiveSnapshot{ID: "live-1"}, nil
}

func (s *ignoreRuleAnalysisStore) CreateIncident(context.Context, storage.CreateIncidentParams) (model.DriftIncident, error) {
	s.incidentsCreated++
	return model.DriftIncident{ID: "incident-1"}, nil
}

func (*ignoreRuleAnalysisStore) CreateDriftField(context.Context, storage.CreateDriftFieldParams) (model.DriftField, error) {
	return model.DriftField{}, nil
}

func (s *ignoreRuleAnalysisStore) CreateEvidenceRecord(_ context.Context, params storage.CreateEvidenceRecordParams) (model.EvidenceRecord, error) {
	s.evidenceRecords = append(s.evidenceRecords, params)
	return model.EvidenceRecord{}, nil
}

func (s *ignoreRuleAnalysisStore) CreateSuppressedFinding(_ context.Context, params storage.CreateSuppressedFindingParams) (model.SuppressedFinding, error) {
	if s.suppressionErr != nil {
		return model.SuppressedFinding{}, s.suppressionErr
	}
	s.suppressedFindings = append(s.suppressedFindings, params)
	return model.SuppressedFinding{ID: "suppressed-1"}, nil
}

func (*ignoreRuleAnalysisStore) InsertGitHubEvent(context.Context, storage.UpsertGitHubEventParams) (model.GitHubEvent, error) {
	return model.GitHubEvent{}, nil
}

type staticDesiredStateFetcher struct{}

func (staticDesiredStateFetcher) FetchDesired(context.Context, model.Application, model.SourceDefinition) (sourceingest.DesiredState, error) {
	return sourceingest.DesiredState{
		Revision:  "revision-1",
		Resources: []map[string]any{deploymentWithReplicas(3)},
	}, nil
}

type staticLiveStateCollector struct{}

func (staticLiveStateCollector) CollectLiveState(context.Context, model.Cluster, model.Application) (k8scollect.LiveState, error) {
	return k8scollect.LiveState{
		Resources: []map[string]any{deploymentWithReplicas(2)},
		Summary:   map[string]any{"source": "kubectl"},
	}, nil
}

type unchangedLiveStateCollector struct{}

func (unchangedLiveStateCollector) CollectLiveState(context.Context, model.Cluster, model.Application) (k8scollect.LiveState, error) {
	return k8scollect.LiveState{Resources: []map[string]any{deploymentWithReplicas(3)}}, nil
}

func deploymentWithReplicas(replicas int) map[string]any {
	return map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			"name":      "ledger-api",
			"namespace": "payments",
		},
		"spec": map[string]any{"replicas": replicas},
	}
}
