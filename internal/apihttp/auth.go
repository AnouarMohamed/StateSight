package apihttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/AnouarMohamed/StateSight/internal/auth"
	"github.com/AnouarMohamed/StateSight/internal/storage"
)

func (s *Server) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, err := s.authenticator.Authenticate(r.Context(), r)
		if err != nil {
			s.writeBearerUnauthorized(w, r)
			return
		}

		userID, err := s.store.GetUserIDByIdentity(r.Context(), identity.Issuer, identity.Subject)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				writeError(w, http.StatusForbidden, "identity_not_provisioned", "authenticated identity is not provisioned", s.responseMeta(r))
				return
			}
			s.logger.Error("identity mapping lookup failed", "error", err.Error(), "request_id", requestIDFromContext(r.Context()))
			writeError(w, http.StatusInternalServerError, "authentication_failed", "failed to resolve authenticated identity", s.responseMeta(r))
			return
		}

		principal := auth.Principal{
			UserID:  userID,
			Issuer:  identity.Issuer,
			Subject: identity.Subject,
			Email:   identity.Email,
		}
		next.ServeHTTP(w, r.WithContext(auth.ContextWithPrincipal(r.Context(), principal)))
	})
}

func (s *Server) selectedWorkspaceID(w http.ResponseWriter, r *http.Request) (string, bool) {
	workspaceID := strings.TrimSpace(r.Header.Get("X-Workspace-ID"))
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_required", "X-Workspace-ID is required to select a workspace", s.responseMeta(r))
		return "", false
	}
	return workspaceID, true
}

func (s *Server) authorizeWorkspace(w http.ResponseWriter, r *http.Request, workspaceID, requiredRole string) bool {
	if s.authenticator == nil {
		return true
	}

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		s.writeBearerUnauthorized(w, r)
		return false
	}
	if strings.TrimSpace(workspaceID) == "" {
		writeError(w, http.StatusForbidden, "workspace_forbidden", "workspace access denied", s.responseMeta(r))
		return false
	}

	role, err := s.store.GetWorkspaceRole(r.Context(), principal.UserID, workspaceID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
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

func (s *Server) writeBearerUnauthorized(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="statesight"`)
	writeError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required", s.responseMeta(r))
}
