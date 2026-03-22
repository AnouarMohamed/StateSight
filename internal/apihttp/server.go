package apihttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/AnouarMohamed/StateSight/internal/jobs"
	"github.com/AnouarMohamed/StateSight/internal/render"
	"github.com/AnouarMohamed/StateSight/internal/storage"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

type Store interface {
	Ping(ctx context.Context) error
	GetOverview(ctx context.Context) (model.Overview, error)
	ListApplications(ctx context.Context) ([]model.Application, error)
	CreateApplication(ctx context.Context, params storage.CreateApplicationParams) (model.Application, error)
	GetApplicationByID(ctx context.Context, id string) (model.Application, error)
	ListIncidentsByApplication(ctx context.Context, applicationID string) ([]model.DriftIncident, error)
	CreateJob(ctx context.Context, params storage.CreateJobParams) (model.Job, error)
	MarkJobFailed(ctx context.Context, id, message string) error
	GetIncidentDetails(ctx context.Context, id string) (model.IncidentDetails, error)
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
}

func NewServer(store Store, queue JobQueue, logger *slog.Logger, webhookSecret string) *Server {
	return &Server{
		store:         store,
		queue:         queue,
		logger:        logger,
		webhookSecret: webhookSecret,
		metrics:       &metrics{},
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
		v1.Get("/incidents/{id}", s.handleGetIncident)
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
	overview, err := s.store.GetOverview(r.Context())
	if err != nil {
		s.logger.Error("overview query failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "overview_query_failed", "failed to load overview", s.responseMeta(r))
		return
	}
	writeSuccess(w, http.StatusOK, overview, s.responseMeta(r))
}

func (s *Server) handleListApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := s.store.ListApplications(r.Context())
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

	incidents, err := s.store.ListIncidentsByApplication(r.Context(), app.ID)
	if err != nil {
		s.logger.Error("list application incidents failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "application_incidents_query_failed", "failed to load application incidents", s.responseMeta(r))
		return
	}

	writeSuccess(w, http.StatusOK, applicationDetailsResponse{
		Application: app,
		Incidents:   incidents,
	}, s.responseMeta(r))
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

	writeSuccess(w, http.StatusOK, details, s.responseMeta(r))
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

func (s *Server) responseMeta(r *http.Request) map[string]any {
	return map[string]any{
		"request_id": requestIDFromContext(r.Context()),
	}
}
