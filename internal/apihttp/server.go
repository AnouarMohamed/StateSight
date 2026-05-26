package apihttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/AnouarMohamed/StateSight/internal/auth"
	"github.com/AnouarMohamed/StateSight/internal/jobs"
	"github.com/AnouarMohamed/StateSight/internal/render"
	"github.com/AnouarMohamed/StateSight/internal/storage"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

type Store interface {
	Ping(ctx context.Context) error
	GetOverview(ctx context.Context) (model.Overview, error)
	GetOverviewByWorkspace(ctx context.Context, workspaceID string) (model.Overview, error)
	ListApplications(ctx context.Context) ([]model.Application, error)
	ListApplicationsByWorkspace(ctx context.Context, workspaceID string) ([]model.Application, error)
	CreateApplication(ctx context.Context, params storage.CreateApplicationParams) (model.Application, error)
	GetApplicationByID(ctx context.Context, id string) (model.Application, error)
	ListIncidentsByApplication(ctx context.Context, applicationID string) ([]model.DriftIncident, error)
	ListSuppressedFindingsByApplication(ctx context.Context, applicationID string) ([]model.SuppressedFinding, error)
	ListIgnoreRulesForApplication(ctx context.Context, workspaceID, applicationID string) ([]model.IgnoreRule, error)
	CreateIgnoreRule(ctx context.Context, params storage.CreateIgnoreRuleParams) (model.IgnoreRule, error)
	UpdateIgnoreRuleForApplication(ctx context.Context, id, applicationID string, params storage.UpdateIgnoreRuleParams) (model.IgnoreRule, error)
	SetIgnoreRuleActiveForApplication(ctx context.Context, id, applicationID string, active bool) (model.IgnoreRule, error)
	DeleteIgnoreRuleForApplication(ctx context.Context, id, applicationID string) error
	CreateJob(ctx context.Context, params storage.CreateJobParams) (model.Job, error)
	MarkJobFailed(ctx context.Context, id, message string) error
	GetIncidentDetails(ctx context.Context, id string) (model.IncidentDetails, error)
	GetIncidentTimeline(ctx context.Context, incidentID string) ([]model.TimelineEvent, error)
	GetWorkspaceRole(ctx context.Context, userID, workspaceID string) (string, error)
}

type JobQueue interface {
	Enqueue(ctx context.Context, msg jobs.Message) error
	Ping(ctx context.Context) error
}

type Server struct {
	store         Store
	queue         JobQueue
	logger        *slog.Logger
	webhookSecret string
	metrics       *metrics
	authRequired  bool
}

func NewServer(store Store, queue JobQueue, logger *slog.Logger, webhookSecret string, authRequired bool) *Server {
	return &Server{
		store:         store,
		queue:         queue,
		logger:        logger,
		webhookSecret: webhookSecret,
		metrics:       &metrics{},
		authRequired:  authRequired,
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(requestIDMiddleware)
	r.Use(s.metrics.requestCounterMiddleware)
	r.Use(loggingMiddleware(s.logger))

	r.Get("/healthz", s.handleHealthz)
	r.Get("/readyz", s.handleReadyz)
	r.Get("/metrics", s.metrics.metricsHandler)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/overview", s.handleOverview)
		v1.Get("/applications", s.handleListApplications)
		v1.Post("/applications", s.handleCreateApplication)
		v1.Get("/applications/{id}", s.handleGetApplication)
		v1.Post("/applications/{id}/analyze", s.handleAnalyzeApplication)
		v1.Post("/applications/{id}/ignore-rules", s.handleCreateIgnoreRule)
		v1.Put("/applications/{id}/ignore-rules/{ruleID}", s.handleUpdateIgnoreRule)
		v1.Patch("/applications/{id}/ignore-rules/{ruleID}", s.handleSetIgnoreRuleActive)
		v1.Delete("/applications/{id}/ignore-rules/{ruleID}", s.handleDeleteIgnoreRule)
		v1.Get("/incidents/{id}", s.handleGetIncident)
		v1.Get("/incidents/{id}/timeline", s.handleGetIncidentTimeline)
		v1.Post("/github/webhook", s.handleGitHubWebhook)
	})

	return r
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"}, s.responseMeta(r))
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := s.store.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database_unavailable", "database is not ready", s.responseMeta(r))
		return
	}
	if err := s.queue.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "redis queue is not ready", s.responseMeta(r))
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ready"}, s.responseMeta(r))
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	var (
		overview model.Overview
		err      error
	)

	if s.authRequired {
		principal, authErr := auth.PrincipalFromRequest(r)
		if authErr != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing required authentication headers", s.responseMeta(r))
			return
		}
		if !s.authorizeWorkspace(w, r, principal.WorkspaceID, auth.RoleViewer) {
			return
		}
		overview, err = s.store.GetOverviewByWorkspace(r.Context(), principal.WorkspaceID)
	} else {
		overview, err = s.store.GetOverview(r.Context())
	}

	if err != nil {
		s.logger.Error("overview query failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "overview_query_failed", "failed to load overview", s.responseMeta(r))
		return
	}
	writeSuccess(w, http.StatusOK, overview, s.responseMeta(r))
}

func (s *Server) handleListApplications(w http.ResponseWriter, r *http.Request) {
	var (
		apps []model.Application
		err  error
	)

	if s.authRequired {
		principal, authErr := auth.PrincipalFromRequest(r)
		if authErr != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing required authentication headers", s.responseMeta(r))
			return
		}
		if !s.authorizeWorkspace(w, r, principal.WorkspaceID, auth.RoleViewer) {
			return
		}
		apps, err = s.store.ListApplicationsByWorkspace(r.Context(), principal.WorkspaceID)
	} else {
		apps, err = s.store.ListApplications(r.Context())
	}

	if err != nil {
		s.logger.Error("list applications failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "applications_query_failed", "failed to load applications", s.responseMeta(r))
		return
	}
	writeSuccess(w, http.StatusOK, apps, s.responseMeta(r))
}

func (s *Server) handleCreateApplication(w http.ResponseWriter, r *http.Request) {
	var req createApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON", s.responseMeta(r))
		return
	}

	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Namespace) == "" ||
		strings.TrimSpace(req.WorkspaceID) == "" || strings.TrimSpace(req.ClusterID) == "" || strings.TrimSpace(req.SourceDefinitionID) == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", "workspace_id, cluster_id, source_definition_id, name, and namespace are required", s.responseMeta(r))
		return
	}
	if !s.authorizeWorkspace(w, r, req.WorkspaceID, auth.RoleEditor) {
		return
	}

	app, err := s.store.CreateApplication(r.Context(), storage.CreateApplicationParams{
		WorkspaceID:        req.WorkspaceID,
		ClusterID:          req.ClusterID,
		SourceDefinitionID: req.SourceDefinitionID,
		Name:               req.Name,
		Namespace:          req.Namespace,
	})
	if err != nil {
		s.logger.Error("create application failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_create_failed", "failed to create application", s.responseMeta(r))
		return
	}
	writeSuccess(w, http.StatusCreated, app, s.responseMeta(r))
}

func (s *Server) handleGetApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	app, err := s.store.GetApplicationByID(r.Context(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "application_not_found", "application was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("get application failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_query_failed", "failed to load application", s.responseMeta(r))
		return
	}
	if !s.authorizeWorkspace(w, r, app.WorkspaceID, auth.RoleViewer) {
		return
	}

	incidents, err := s.store.ListIncidentsByApplication(r.Context(), app.ID)
	if err != nil {
		s.logger.Error("list application incidents failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_incidents_query_failed", "failed to load application incidents", s.responseMeta(r))
		return
	}

	suppressions, err := s.store.ListSuppressedFindingsByApplication(r.Context(), app.ID)
	if err != nil {
		s.logger.Error("list application suppressed findings failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_suppressions_query_failed", "failed to load application suppressions", s.responseMeta(r))
		return
	}

	ignoreRules, err := s.store.ListIgnoreRulesForApplication(r.Context(), app.WorkspaceID, app.ID)
	if err != nil {
		s.logger.Error("list application ignore rules failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_ignore_rules_query_failed", "failed to load application ignore rules", s.responseMeta(r))
		return
	}

	writeSuccess(w, http.StatusOK, applicationDetailsResponse{
		Application:  app,
		Incidents:    incidents,
		Suppressions: suppressions,
		IgnoreRules:  ignoreRules,
	}, s.responseMeta(r))
}

func (s *Server) handleCreateIgnoreRule(w http.ResponseWriter, r *http.Request) {
	app, err := s.store.GetApplicationByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "application_not_found", "application was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("lookup application for ignore rule failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_query_failed", "failed to load application", s.responseMeta(r))
		return
	}
	if !s.authorizeWorkspace(w, r, app.WorkspaceID, auth.RoleEditor) {
		return
	}

	var req ignoreRuleDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON", s.responseMeta(r))
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.MatchExpression = strings.TrimSpace(req.MatchExpression)
	req.ResourceRef = strings.TrimSpace(req.ResourceRef)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Name == "" || req.MatchExpression == "" || req.Reason == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", "name, match_expression, and reason are required", s.responseMeta(r))
		return
	}

	createdBy := "api"
	if s.authRequired {
		principal, _ := auth.PrincipalFromRequest(r)
		createdBy = principal.UserID
	}
	rule, err := s.store.CreateIgnoreRule(r.Context(), storage.CreateIgnoreRuleParams{
		WorkspaceID:     app.WorkspaceID,
		ApplicationID:   app.ID,
		ResourceRef:     req.ResourceRef,
		Name:            req.Name,
		MatchExpression: req.MatchExpression,
		Reason:          req.Reason,
		CreatedBy:       createdBy,
	})
	if err != nil {
		if errors.Is(err, storage.ErrConflict) {
			writeError(w, http.StatusConflict, "ignore_rule_name_conflict", "an ignore rule with this name already exists for this application", s.responseMeta(r))
			return
		}
		s.logger.Error("create application ignore rule failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "ignore_rule_create_failed", "failed to create ignore rule", s.responseMeta(r))
		return
	}
	writeSuccess(w, http.StatusCreated, rule, s.responseMeta(r))
}

func (s *Server) handleUpdateIgnoreRule(w http.ResponseWriter, r *http.Request) {
	app, err := s.store.GetApplicationByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "application_not_found", "application was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("lookup application for ignore rule edit failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_query_failed", "failed to load application", s.responseMeta(r))
		return
	}
	if !s.authorizeWorkspace(w, r, app.WorkspaceID, auth.RoleEditor) {
		return
	}

	var req ignoreRuleDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON", s.responseMeta(r))
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.MatchExpression = strings.TrimSpace(req.MatchExpression)
	req.ResourceRef = strings.TrimSpace(req.ResourceRef)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Name == "" || req.MatchExpression == "" || req.Reason == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", "name, match_expression, and reason are required", s.responseMeta(r))
		return
	}

	rule, err := s.store.UpdateIgnoreRuleForApplication(r.Context(), chi.URLParam(r, "ruleID"), app.ID, storage.UpdateIgnoreRuleParams{
		ResourceRef:     req.ResourceRef,
		Name:            req.Name,
		MatchExpression: req.MatchExpression,
		Reason:          req.Reason,
	})
	if err != nil {
		if errors.Is(err, storage.ErrConflict) {
			writeError(w, http.StatusConflict, "ignore_rule_name_conflict", "an ignore rule with this name already exists for this application", s.responseMeta(r))
			return
		}
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ignore_rule_not_found", "application-owned ignore rule was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("edit application ignore rule failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "ignore_rule_update_failed", "failed to update ignore rule", s.responseMeta(r))
		return
	}
	writeSuccess(w, http.StatusOK, rule, s.responseMeta(r))
}

func (s *Server) handleSetIgnoreRuleActive(w http.ResponseWriter, r *http.Request) {
	app, err := s.store.GetApplicationByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "application_not_found", "application was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("lookup application for ignore rule update failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_query_failed", "failed to load application", s.responseMeta(r))
		return
	}
	if !s.authorizeWorkspace(w, r, app.WorkspaceID, auth.RoleEditor) {
		return
	}

	var req updateIgnoreRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON", s.responseMeta(r))
		return
	}
	if req.Active == nil {
		writeError(w, http.StatusBadRequest, "missing_fields", "active is required", s.responseMeta(r))
		return
	}

	rule, err := s.store.SetIgnoreRuleActiveForApplication(r.Context(), chi.URLParam(r, "ruleID"), app.ID, *req.Active)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ignore_rule_not_found", "application-owned ignore rule was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("update application ignore rule failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "ignore_rule_update_failed", "failed to update ignore rule", s.responseMeta(r))
		return
	}
	writeSuccess(w, http.StatusOK, rule, s.responseMeta(r))
}

func (s *Server) handleDeleteIgnoreRule(w http.ResponseWriter, r *http.Request) {
	app, err := s.store.GetApplicationByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "application_not_found", "application was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("lookup application for ignore rule deletion failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_query_failed", "failed to load application", s.responseMeta(r))
		return
	}
	if !s.authorizeWorkspace(w, r, app.WorkspaceID, auth.RoleEditor) {
		return
	}

	ruleID := chi.URLParam(r, "ruleID")
	if err := s.store.DeleteIgnoreRuleForApplication(r.Context(), ruleID, app.ID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ignore_rule_not_found", "application-owned ignore rule was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("delete application ignore rule failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "ignore_rule_delete_failed", "failed to delete ignore rule", s.responseMeta(r))
		return
	}
	writeSuccess(w, http.StatusOK, map[string]string{"id": ruleID, "status": "deleted"}, s.responseMeta(r))
}

func (s *Server) handleAnalyzeApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	app, err := s.store.GetApplicationByID(r.Context(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "application_not_found", "application was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("lookup application for analyze failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_query_failed", "failed to load application", s.responseMeta(r))
		return
	}
	if !s.authorizeWorkspace(w, r, app.WorkspaceID, auth.RoleEditor) {
		return
	}

	payloadJSON, err := render.JSON(map[string]any{
		"trigger":        "api",
		"requested_at":   time.Now().UTC().Format(time.RFC3339),
		"request_id":     requestIDFromContext(r.Context()),
		"application_id": app.ID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "job_payload_build_failed", "failed to prepare job payload", s.responseMeta(r))
		return
	}

	job, err := s.store.CreateJob(r.Context(), storage.CreateJobParams{
		JobType:       model.JobTypeAnalyzeApplication,
		ApplicationID: &app.ID,
		PayloadJSON:   payloadJSON,
	})
	if err != nil {
		s.logger.Error("create analyze job failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "job_create_failed", "failed to enqueue analysis", s.responseMeta(r))
		return
	}

	queueMessage := jobs.Message{
		JobID:         job.ID,
		JobType:       model.JobTypeAnalyzeApplication,
		ApplicationID: app.ID,
		Payload: map[string]any{
			"request_id": requestIDFromContext(r.Context()),
		},
		EnqueuedAt: time.Now().UTC(),
	}
	if err := s.queue.Enqueue(r.Context(), queueMessage); err != nil {
		_ = s.store.MarkJobFailed(r.Context(), job.ID, fmt.Sprintf("queue enqueue failed: %v", err))
		s.logger.Error("enqueue analyze job failed", "error", err.Error(), "job_id", job.ID, "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "queue_enqueue_failed", "failed to enqueue job", s.responseMeta(r))
		return
	}

	s.metrics.markJobEnqueued()
	writeSuccess(w, http.StatusAccepted, analyzeResponse{
		JobID:         job.ID,
		JobType:       job.JobType,
		Status:        job.Status,
		ApplicationID: app.ID,
	}, s.responseMeta(r))
}

func (s *Server) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	details, err := s.store.GetIncidentDetails(r.Context(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "incident_not_found", "incident was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("get incident failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "incident_query_failed", "failed to load incident", s.responseMeta(r))
		return
	}
	app, err := s.store.GetApplicationByID(r.Context(), details.Incident.ApplicationID)
	if err != nil {
		s.logger.Error("incident application lookup failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "incident_application_query_failed", "failed to authorize incident workspace", s.responseMeta(r))
		return
	}
	if !s.authorizeWorkspace(w, r, app.WorkspaceID, auth.RoleViewer) {
		return
	}

	writeSuccess(w, http.StatusOK, details, s.responseMeta(r))
}

func (s *Server) handleGetIncidentTimeline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	details, err := s.store.GetIncidentDetails(r.Context(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "incident_not_found", "incident was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("get incident for timeline failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "incident_query_failed", "failed to load incident", s.responseMeta(r))
		return
	}
	app, err := s.store.GetApplicationByID(r.Context(), details.Incident.ApplicationID)
	if err != nil {
		s.logger.Error("incident application lookup failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "incident_application_query_failed", "failed to authorize incident workspace", s.responseMeta(r))
		return
	}
	if !s.authorizeWorkspace(w, r, app.WorkspaceID, auth.RoleViewer) {
		return
	}

	timeline, err := s.store.GetIncidentTimeline(r.Context(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "incident_not_found", "incident was not found", s.responseMeta(r))
			return
		}
		s.logger.Error("get incident timeline failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "incident_timeline_query_failed", "failed to load incident timeline", s.responseMeta(r))
		return
	}
	writeSuccess(w, http.StatusOK, timeline, s.responseMeta(r))
}

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "body_read_failed", "could not read request body", s.responseMeta(r))
		return
	}

	signatureHeader := r.Header.Get("X-Hub-Signature-256")
	if !isValidGitHubSignature(s.webhookSecret, body, signatureHeader) {
		writeError(w, http.StatusUnauthorized, "invalid_signature", "github webhook signature validation failed", s.responseMeta(r))
		return
	}

	payload, err := jobs.PayloadFromRawJSON(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "webhook payload must be valid JSON", s.responseMeta(r))
		return
	}

	rawJSON := string(body)
	job, err := s.store.CreateJob(r.Context(), storage.CreateJobParams{
		JobType:       model.JobTypeIngestGitHubEvent,
		ApplicationID: nil,
		PayloadJSON:   rawJSON,
	})
	if err != nil {
		s.logger.Error("create github event job failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "job_create_failed", "failed to queue github webhook event", s.responseMeta(r))
		return
	}

	message := jobs.Message{
		JobID:      job.ID,
		JobType:    model.JobTypeIngestGitHubEvent,
		Payload:    payload,
		EnqueuedAt: time.Now().UTC(),
		Headers: map[string]string{
			"x-github-event":    r.Header.Get("X-GitHub-Event"),
			"x-github-delivery": r.Header.Get("X-GitHub-Delivery"),
		},
	}
	if err := s.queue.Enqueue(r.Context(), message); err != nil {
		_ = s.store.MarkJobFailed(r.Context(), job.ID, fmt.Sprintf("queue enqueue failed: %v", err))
		s.logger.Error("enqueue github event job failed", "error", err.Error(), "job_id", job.ID, "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "queue_enqueue_failed", "failed to enqueue webhook event", s.responseMeta(r))
		return
	}

	s.metrics.markJobEnqueued()
	writeSuccess(w, http.StatusAccepted, map[string]any{
		"job_id":   job.ID,
		"job_type": job.JobType,
		"status":   job.Status,
	}, s.responseMeta(r))
}

func (s *Server) authorizeWorkspace(w http.ResponseWriter, r *http.Request, workspaceID, requiredRole string) bool {
	if !s.authRequired {
		return true
	}

	principal, err := auth.PrincipalFromRequest(r)
	if err != nil {
		code := "unauthorized"
		message := "missing required authentication headers"
		if errors.Is(err, auth.ErrMissingWorkspaceID) {
			message = "missing X-Workspace-ID header"
		}
		writeError(w, http.StatusUnauthorized, code, message, s.responseMeta(r))
		return false
	}

	if strings.TrimSpace(workspaceID) == "" || workspaceID != principal.WorkspaceID {
		writeError(w, http.StatusForbidden, "workspace_forbidden", "workspace access denied", s.responseMeta(r))
		return false
	}

	role, err := s.store.GetWorkspaceRole(r.Context(), principal.UserID, workspaceID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusForbidden, "workspace_forbidden", "workspace membership not found", s.responseMeta(r))
			return false
		}
		s.logger.Error("workspace role lookup failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "workspace_auth_failed", "failed to evaluate workspace role", s.responseMeta(r))
		return false
	}

	if !auth.HasRequiredRole(role, requiredRole) {
		writeError(w, http.StatusForbidden, "insufficient_role", "role does not permit this action", s.responseMeta(r))
		return false
	}

	return true
}

func (s *Server) responseMeta(r *http.Request) map[string]any {
	return map[string]any{
		"request_id": requestIDFromContext(r.Context()),
	}
}
