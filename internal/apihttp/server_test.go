package apihttp

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnouarMohamed/StateSight/internal/jobs"
	"github.com/AnouarMohamed/StateSight/internal/storage"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

type mockStore struct{}

func (m mockStore) Ping(context.Context) error                          { return nil }
func (m mockStore) GetOverview(context.Context) (model.Overview, error) { return model.Overview{}, nil }
func (m mockStore) GetOverviewByWorkspace(context.Context, string) (model.Overview, error) {
	return model.Overview{}, nil
}
func (m mockStore) ListApplications(context.Context) ([]model.Application, error) { return nil, nil }
func (m mockStore) ListApplicationsByWorkspace(context.Context, string) ([]model.Application, error) {
	return nil, nil
}
func (m mockStore) CreateApplication(context.Context, storage.CreateApplicationParams) (model.Application, error) {
	return model.Application{}, nil
}
func (m mockStore) GetApplicationByID(context.Context, string) (model.Application, error) {
	return model.Application{}, storage.ErrNotFound
}
func (m mockStore) ListIncidentsByApplication(context.Context, string) ([]model.DriftIncident, error) {
	return nil, nil
}
func (m mockStore) ListSuppressedFindingsByApplication(context.Context, string) ([]model.SuppressedFinding, error) {
	return nil, nil
}
func (m mockStore) CreateJob(context.Context, storage.CreateJobParams) (model.Job, error) {
	return model.Job{}, nil
}
func (m mockStore) MarkJobFailed(context.Context, string, string) error { return nil }
func (m mockStore) GetIncidentDetails(context.Context, string) (model.IncidentDetails, error) {
	return model.IncidentDetails{}, storage.ErrNotFound
}
func (m mockStore) GetIncidentTimeline(context.Context, string) ([]model.TimelineEvent, error) {
	return nil, storage.ErrNotFound
}
func (m mockStore) GetWorkspaceRole(context.Context, string, string) (string, error) {
	return "admin", nil
}

type mockQueue struct{}

func (q mockQueue) Enqueue(context.Context, jobs.Message) error { return nil }
func (q mockQueue) Ping(context.Context) error                  { return nil }

func TestHealthz(t *testing.T) {
	s := NewServer(mockStore{}, mockQueue{}, slog.Default(), "", false)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) == "" {
		t.Fatal("expected body to be non-empty")
	}
}

type applicationDetailsStore struct {
	mockStore
}

func (applicationDetailsStore) GetApplicationByID(context.Context, string) (model.Application, error) {
	return model.Application{ID: "application-1"}, nil
}

func (applicationDetailsStore) ListSuppressedFindingsByApplication(context.Context, string) ([]model.SuppressedFinding, error) {
	return []model.SuppressedFinding{{ID: "suppressed-1", FieldPath: "spec.replicas"}}, nil
}

func TestGetApplicationIncludesSuppressedFindings(t *testing.T) {
	s := NewServer(applicationDetailsStore{}, mockQueue{}, slog.Default(), "", false)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/applications/application-1", nil)
	rec := httptest.NewRecorder()

	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		Data applicationDetailsResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Data.Suppressions) != 1 || response.Data.Suppressions[0].ID != "suppressed-1" {
		t.Fatalf("expected suppression audit record in application response, got %#v", response.Data.Suppressions)
	}
}
