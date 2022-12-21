/*
Package ctxutil defines common contexts for use across MCMP services.
*/
package ctxutil

import (
	"context"

	"github.com/go-openapi/strfmt"
)

var (
	principalKey = &contextKey{"Principal"}
	tokenKey     = &contextKey{"Token"}
	requestIDKey = &contextKey{"RequestID"}
	tenantIDKey  = &contextKey{"TenantID"}
	sgIDKey      = &contextKey{"ServiceGroupID"}
	clientIDKey  = &contextKey{"ClientID"}
	accountIDKey = &contextKey{"AccountID"}
)

// WithPrincipal adds the provided principal to the context and returns the updated context.
func WithPrincipal(ctx context.Context, principal string) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}

// Principal retrieves the principal from the current context and returns the value.
func Principal(ctx context.Context) string {
	if v := ctx.Value(principalKey); v != nil {
		return v.(string)
	}

	return ""
}

// WithToken adds the provided token to the context and returns the updated context.
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// Token retrieves the token from the current context and returns the value.
func Token(ctx context.Context) string {
	if v := ctx.Value(tokenKey); v != nil {
		return v.(string)
	}

	return ""
}

// WithRequestID adds the provided request ID to the context and returns the updated context.
func WithRequestID(ctx context.Context, requestID strfmt.UUID) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestID retrieves the request ID from the current context and returns the value.
func RequestID(ctx context.Context) strfmt.UUID {
	if v := ctx.Value(requestIDKey); v != nil {
		return v.(strfmt.UUID)
	}

	return strfmt.UUID("")
}

// WithTenantID adds the provided tenant ID to the context and returns the updated context.
func WithTenantID(ctx context.Context, tenantID strfmt.UUID) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// TenantID retrieves the tenant ID from the current context and returns the value.
func TenantID(ctx context.Context) strfmt.UUID {
	if v := ctx.Value(tenantIDKey); v != nil {
		return v.(strfmt.UUID)
	}

	return strfmt.UUID("")
}

// WithServiceGroupID adds the provided servicegroup ID to the context and returns the updated context.
func WithServiceGroupID(ctx context.Context, sgID strfmt.UUID) context.Context {
	return context.WithValue(ctx, sgIDKey, sgID)
}

// ServiceGroupID retrieves the servicegroup ID from the current context and returns the value.
func ServiceGroupID(ctx context.Context) strfmt.UUID {
	if v := ctx.Value(sgIDKey); v != nil {
		return v.(strfmt.UUID)
	}

	return strfmt.UUID("")
}

// WithAccountID adds the provided account ID to the context and returns the updated context.
func WithAccountID(ctx context.Context, acctID strfmt.UUID) context.Context {
	return context.WithValue(ctx, accountIDKey, acctID)
}

// AccountID retrieves the account ID from the current context and returns the value.
func AccountID(ctx context.Context) strfmt.UUID {
	if v := ctx.Value(accountIDKey); v != nil {
		return v.(strfmt.UUID)
	}

	return strfmt.UUID("")
}

// WithClientID adds the provided clientID to context and returns updated content.
func WithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDKey, clientID)
}

// ClientID returns the authz stored in this context.
func ClientID(ctx context.Context) string {
	if v := ctx.Value(clientIDKey); v != nil {
		return v.(string)
	}

	return ""
}

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "ctxutil context value " + k.name
}
