package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/AnouarMohamed/StateSight/internal/diff"
	"github.com/AnouarMohamed/StateSight/internal/evidence"
	"github.com/AnouarMohamed/StateSight/internal/ignorerules"
	"github.com/AnouarMohamed/StateSight/internal/incidents"
	"github.com/AnouarMohamed/StateSight/internal/k8scollect"
	"github.com/AnouarMohamed/StateSight/internal/normalize"
	"github.com/AnouarMohamed/StateSight/internal/render"
	"github.com/AnouarMohamed/StateSight/internal/scoring"
	"github.com/AnouarMohamed/StateSight/internal/sourceingest"
	"github.com/AnouarMohamed/StateSight/internal/storage"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

type Store interface {
	MarkJobProcessing(ctx context.Context, id string) error
	MarkJobCompleted(ctx context.Context, id string) error
	MarkJobFailed(ctx context.Context, id, message string) error
	GetApplicationByID(ctx context.Context, id string) (model.Application, error)
	GetSourceDefinitionByID(ctx context.Context, id string) (model.SourceDefinition, error)
	CreateDesiredSnapshot(ctx context.Context, params storage.CreateDesiredSnapshotParams) (model.DesiredSnapshot, error)
	CreateLiveSnapshot(ctx context.Context, params storage.CreateLiveSnapshotParams) (model.LiveSnapshot, error)
	CreateIncident(ctx context.Context, params storage.CreateIncidentParams) (model.DriftIncident, error)
	CreateDriftField(ctx context.Context, params storage.CreateDriftFieldParams) (model.DriftField, error)
	CreateEvidenceRecord(ctx context.Context, params storage.CreateEvidenceRecordParams) (model.EvidenceRecord, error)
	InsertGitHubEvent(ctx context.Context, params storage.UpsertGitHubEventParams) (model.GitHubEvent, error)
}

type Processor struct {
	store           Store
	fetcher         sourceingest.Fetcher
	collector       k8scollect.Collector
	normalizer      normalize.Normalizer
	diffEngine      diff.Engine
	grouper         incidents.Grouper
	attributor      evidence.Attributor
	recommendation  scoring.Recommendation
	ignoreEvaluator ignorerules.Evaluator
	logger          *slog.Logger
}

func NewProcessor(store Store, logger *slog.Logger) *Processor {
	return &Processor{
		store:           store,
		fetcher:         sourceingest.NewGitFetcher("git", ".statesight/git-cache"),
		collector:       k8scollect.MockCollector{},
		normalizer:      normalize.PassThroughNormalizer{},
		diffEngine:      diff.SeededEngine{},
		grouper:         incidents.SimpleGrouper{},
		attributor:      evidence.MockAttributor{},
		recommendation:  scoring.RuleBasedRecommendation{},
		ignoreEvaluator: ignorerules.NoopEvaluator{},
		logger:          logger,
	}
}

func (p *Processor) HandleMessage(ctx context.Context, msg Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("validate message: %w", err)
	}
	if msg.JobID == "" {
		return fmt.Errorf("job id is required")
	}

	if err := p.store.MarkJobProcessing(ctx, msg.JobID); err != nil {
		return fmt.Errorf("mark job processing: %w", err)
	}

	var err error
	switch msg.JobType {
	case model.JobTypeAnalyzeApplication:
		err = p.processAnalyze(ctx, msg)
	case model.JobTypeIngestGitHubEvent:
		err = p.processGitHubEvent(ctx, msg)
	default:
		err = fmt.Errorf("unsupported job type: %s", msg.JobType)
	}

	if err != nil {
		p.logger.Error("job processing failed", "job_id", msg.JobID, "job_type", msg.JobType, "error", err.Error())
		if markErr := p.store.MarkJobFailed(ctx, msg.JobID, err.Error()); markErr != nil {
			return fmt.Errorf("mark failed job: %w", markErr)
		}
		return nil
	}

	if err := p.store.MarkJobCompleted(ctx, msg.JobID); err != nil {
		return fmt.Errorf("mark job completed: %w", err)
	}

	p.logger.Info("job completed", "job_id", msg.JobID, "job_type", msg.JobType)
	return nil
}

func (p *Processor) processAnalyze(ctx context.Context, msg Message) error {
	if msg.ApplicationID == "" {
		return fmt.Errorf("application_id is required for analyze jobs")
	}

	app, err := p.store.GetApplicationByID(ctx, msg.ApplicationID)
	if err != nil {
		return fmt.Errorf("load application: %w", err)
	}

	source, err := p.store.GetSourceDefinitionByID(ctx, app.SourceDefinitionID)
	if err != nil {
		return fmt.Errorf("load source definition: %w", err)
	}

	desiredState, err := p.fetcher.FetchDesired(ctx, app, source)
	if err != nil {
		return fmt.Errorf("fetch desired state: %w", err)
	}
	liveState, err := p.collector.CollectLiveState(ctx, app)
	if err != nil {
		return fmt.Errorf("collect live state: %w", err)
	}

	normalizedDesired := p.normalizer.Normalize(desiredState.Summary)
	normalizedLive := p.normalizer.Normalize(liveState.Summary)

	desiredJSON, err := render.JSON(normalizedDesired.Body)
	if err != nil {
		return fmt.Errorf("render desired snapshot: %w", err)
	}
	liveJSON, err := render.JSON(normalizedLive.Body)
	if err != nil {
		return fmt.Errorf("render live snapshot: %w", err)
	}

	desiredSnapshot, err := p.store.CreateDesiredSnapshot(ctx, storage.CreateDesiredSnapshotParams{
		ApplicationID: app.ID,
		Revision:      desiredState.Revision,
		SummaryJSON:   desiredJSON,
	})
	if err != nil {
		return fmt.Errorf("create desired snapshot: %w", err)
	}

	liveSnapshot, err := p.store.CreateLiveSnapshot(ctx, storage.CreateLiveSnapshotParams{
		ApplicationID: app.ID,
		SummaryJSON:   liveJSON,
	})
	if err != nil {
		return fmt.Errorf("create live snapshot: %w", err)
	}

	findings, err := p.diffEngine.Compare(ctx, app, normalizedDesired, normalizedLive)
	if err != nil {
		return fmt.Errorf("run diff engine: %w", err)
	}

	candidates, err := p.grouper.Group(ctx, findings)
	if err != nil {
		return fmt.Errorf("group incidents: %w", err)
	}

	for _, candidate := range candidates {
		ignored, reason, err := p.ignoreEvaluator.ShouldIgnore(ctx, app.ID, candidate.FieldPath)
		if err != nil {
			return fmt.Errorf("evaluate ignore rule: %w", err)
		}
		if ignored {
			p.logger.Info("candidate ignored by rule", "application_id", app.ID, "field_path", candidate.FieldPath, "reason", reason)
			continue
		}

		action, scoredConfidence, err := p.recommendation.Recommend(ctx, candidate)
		if err != nil {
			return fmt.Errorf("recommend action: %w", err)
		}

		incident, err := p.store.CreateIncident(ctx, storage.CreateIncidentParams{
			ApplicationID:     app.ID,
			DesiredSnapshotID: desiredSnapshot.ID,
			LiveSnapshotID:    liveSnapshot.ID,
			Title:             candidate.Title,
			Category:          candidate.Category,
			Severity:          candidate.Severity,
			Confidence:        scoredConfidence,
			RecommendedAction: action,
			Status:            "open",
		})
		if err != nil {
			return fmt.Errorf("create incident: %w", err)
		}

		_, err = p.store.CreateDriftField(ctx, storage.CreateDriftFieldParams{
			IncidentID:     incident.ID,
			ResourceRef:    candidate.ResourceRef,
			FieldPath:      candidate.FieldPath,
			DesiredValue:   candidate.DesiredValue,
			LiveValue:      candidate.LiveValue,
			DifferenceType: candidate.DifferenceType,
		})
		if err != nil {
			return fmt.Errorf("create drift field: %w", err)
		}

		attributions, err := p.attributor.BuildAttributions(ctx, app, candidate)
		if err != nil {
			return fmt.Errorf("build attribution: %w", err)
		}

		for _, attribution := range attributions {
			metadataJSON, err := render.JSON(attribution.Metadata)
			if err != nil {
				return fmt.Errorf("marshal attribution metadata: %w", err)
			}
			_, err = p.store.CreateEvidenceRecord(ctx, storage.CreateEvidenceRecordParams{
				IncidentID: incident.ID,
				Source:     attribution.Source,
				Detail:     attribution.Detail,
				Actor:      attribution.Actor,
				Confidence: attribution.Confidence,
				Metadata:   metadataJSON,
			})
			if err != nil {
				return fmt.Errorf("create evidence record: %w", err)
			}
		}
	}

	return nil
}

func (p *Processor) processGitHubEvent(ctx context.Context, msg Message) error {
	payloadJSON, err := render.JSON(msg.Payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	eventType := valueOrDefault(msg.Headers["x-github-event"], "unknown")
	deliveryID := valueOrDefault(msg.Headers["x-github-delivery"], msg.JobID)
	action := "unknown"
	repository := "unknown"

	if v, ok := msg.Payload["action"].(string); ok && v != "" {
		action = v
	}
	if repoRaw, ok := msg.Payload["repository"].(map[string]any); ok {
		if fullName, ok := repoRaw["full_name"].(string); ok && fullName != "" {
			repository = fullName
		}
	}

	_, err = p.store.InsertGitHubEvent(ctx, storage.UpsertGitHubEventParams{
		EventType:  eventType,
		DeliveryID: deliveryID,
		Action:     action,
		Repository: repository,
		Payload:    payloadJSON,
	})
	if err != nil {
		return fmt.Errorf("insert github event: %w", err)
	}
	return nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func PayloadFromRawJSON(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
