package auth

const (
	RoleViewer = "viewer"
	RoleEditor = "editor"
	RoleAdmin  = "admin"
)

func HasRequiredRole(actualRole, requiredRole string) bool {
	return roleRank(actualRole) >= roleRank(requiredRole)
}

func roleRank(role string) int {
	switch role {
	case RoleAdmin:
		return 3
	case RoleEditor:
		return 2
	case RoleViewer:
		return 1
	default:
		return 0
	}
}
