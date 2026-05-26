package auth

import "context"

type Identity struct {
	Issuer  string
	Subject string
	Email   string
}

type Principal struct {
	UserID  string
	Issuer  string
	Subject string
	Email   string
}

type principalContextKey struct{}

func ContextWithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}
