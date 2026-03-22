package auth

import (
	"errors"
	"net/http"
	"strings"
)

var (
	ErrMissingUserID      = errors.New("missing X-User-ID header")
	ErrMissingWorkspaceID = errors.New("missing X-Workspace-ID header")
)

type Principal struct {
	UserID      string
	WorkspaceID string
	Email       string
}

func PrincipalFromRequest(r *http.Request) (Principal, error) {
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if userID == "" {
		return Principal{}, ErrMissingUserID
	}
	workspaceID := strings.TrimSpace(r.Header.Get("X-Workspace-ID"))
	if workspaceID == "" {
		return Principal{}, ErrMissingWorkspaceID
	}

	return Principal{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Email:       strings.TrimSpace(r.Header.Get("X-User-Email")),
	}, nil
}
