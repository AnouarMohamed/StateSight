package apihttp

import (
	"bytes"
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
func (m mockStore) ListIgnoreRulesForApplication(context.Context, string, string) ([]model.IgnoreRule, error) {
	return nil, nil
}
func (m mockStore) CreateIgnoreRule(context.Context, storage.CreateIgnoreRuleParams) (model.IgnoreRule, error) {
	return model.IgnoreRule{}, nil
}
func (m mockStore) UpdateIgnoreRuleForApplication(context.Context, string, string, storage.UpdateIgnoreRuleParams) (model.IgnoreRule, error) {
	return model.IgnoreRule{}, nil
}
func (m mockStore) SetIgnoreRuleActiveForApplication(context.Context, string, string, bool) (model.IgnoreRule, error) {
	return model.IgnoreRule{}, nil
}
func (m mockStore) DeleteIgnoreRuleForApplication(context.Context, string, string) error { return nil }
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

func (applicationDetailsStore) ListIgnoreRulesForApplication(context.Context, string, string) ([]model.IgnoreRule, error) {
	return []model.IgnoreRule{{ID: "rule-1", MatchExpression: "spec.replicas"}}, nil
}

func TestGetApplicationIncludesSuppressionsAndIgnoreRules(t *testing.T) {
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
	if len(response.Data.IgnoreRules) != 1 || response.Data.IgnoreRules[0].ID != "rule-1" {
		t.Fatalf("expected ignore rule in application response, got %#v", response.Data.IgnoreRules)
	}
}

type ignoreRuleMutationStore struct {
	mockStore
	created storage.CreateIgnoreRuleParams
	updated storage.UpdateIgnoreRuleParams
	ruleID  string
	appID   string
	active  bool
	deleted bool
}

func (ignoreRuleMutationStore) GetApplicationByID(context.Context, string) (model.Application, error) {
	return model.Application{ID: "application-1", WorkspaceID: "workspace-1"}, nil
}

func (s *ignoreRuleMutationStore) CreateIgnoreRule(_ context.Context, params storage.CreateIgnoreRuleParams) (model.IgnoreRule, error) {
	s.created = params
	return model.IgnoreRule{ID: "rule-1", ApplicationID: params.ApplicationID, Active: true}, nil
}

func (s *ignoreRuleMutationStore) SetIgnoreRuleActiveForApplication(_ context.Context, ruleID, applicationID string, active bool) (model.IgnoreRule, error) {
	s.ruleID = ruleID
	s.appID = applicationID
	s.active = active
	return model.IgnoreRule{ID: ruleID, ApplicationID: applicationID, Active: active}, nil
}

func (s *ignoreRuleMutationStore) UpdateIgnoreRuleForApplication(_ context.Context, ruleID, applicationID string, params storage.UpdateIgnoreRuleParams) (model.IgnoreRule, error) {
	s.ruleID = ruleID
	s.appID = applicationID
	s.updated = params
	return model.IgnoreRule{ID: ruleID, ApplicationID: applicationID, Name: params.Name}, nil
}

func (s *ignoreRuleMutationStore) DeleteIgnoreRuleForApplication(_ context.Context, ruleID, applicationID string) error {
	s.ruleID = ruleID
	s.appID = applicationID
	s.deleted = true
	return nil
}

func TestCreateIgnoreRuleScopesRuleToApplicationAndActor(t *testing.T) {
	store := &ignoreRuleMutationStore{}
	s := NewServer(store, mockQueue{}, slog.Default(), "", true)
	body := bytes.NewBufferString(`{"name":"  HPA replicas ","match_expression":" spec.replicas ","resource_ref":" apps/v1/Deployment:payments/ledger-api ","reason":" autoscaler "}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/applications/application-1/ignore-rules", body)
	req.Header.Set("X-User-ID", "user-1")
	req.Header.Set("X-Workspace-ID", "workspace-1")
	rec := httptest.NewRecorder()

	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if store.created.ApplicationID != "application-1" || store.created.WorkspaceID != "workspace-1" {
		t.Fatalf("expected application-scoped rule, got %#v", store.created)
	}
	if store.created.CreatedBy != "user-1" || store.created.MatchExpression != "spec.replicas" ||
		store.created.ResourceRef != "apps/v1/Deployment:payments/ledger-api" {
		t.Fatalf("unexpected create parameters: %#v", store.created)
	}
}

func TestSetIgnoreRuleActiveTargetsApplicationOwnedRule(t *testing.T) {
	store := &ignoreRuleMutationStore{}
	s := NewServer(store, mockQueue{}, slog.Default(), "", false)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/applications/application-1/ignore-rules/rule-1", bytes.NewBufferString(`{"active":false}`))
	rec := httptest.NewRecorder()

	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if store.ruleID != "rule-1" || store.appID != "application-1" || store.active {
		t.Fatalf("expected application-owned rule to be disabled, got id=%q app=%q active=%t", store.ruleID, store.appID, store.active)
	}
}

func TestUpdateIgnoreRuleTargetsApplicationOwnedRuleAndTrimsFields(t *testing.T) {
	store := &ignoreRuleMutationStore{}
	s := NewServer(store, mockQueue{}, slog.Default(), "", false)
	body := bytes.NewBufferString(`{"name":"  Keep HPA scale  ","match_expression":" spec.replicas ","resource_ref":" apps/v1/Deployment:payments/ledger-api ","reason":" controller owned "}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/applications/application-1/ignore-rules/rule-1", body)
	rec := httptest.NewRecorder()

	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if store.ruleID != "rule-1" || store.appID != "application-1" {
		t.Fatalf("expected application-owned update target, got id=%q app=%q", store.ruleID, store.appID)
	}
	if store.updated.Name != "Keep HPA scale" || store.updated.MatchExpression != "spec.replicas" ||
		store.updated.ResourceRef != "apps/v1/Deployment:payments/ledger-api" || store.updated.Reason != "controller owned" {
		t.Fatalf("unexpected edit parameters: %#v", store.updated)
	}
}

func TestDeleteIgnoreRuleTargetsApplicationOwnedRule(t *testing.T) {
	store := &ignoreRuleMutationStore{}
	s := NewServer(store, mockQueue{}, slog.Default(), "", false)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/applications/application-1/ignore-rules/rule-1", nil)
	rec := httptest.NewRecorder()

	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !store.deleted || store.ruleID != "rule-1" || store.appID != "application-1" {
		t.Fatalf("expected application-owned rule deletion, got id=%q app=%q deleted=%t", store.ruleID, store.appID, store.deleted)
	}
}
