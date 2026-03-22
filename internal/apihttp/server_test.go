package apihttp

import (
	"context"
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
