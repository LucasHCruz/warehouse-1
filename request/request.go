package request

import (
	"context"
	"github.com/google/uuid"
)

//WithID returns context with contextIDKey
func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextIDKey, id)
}

//IDFromContext returns the contextIDKey
func IDFromContext(ctx context.Context) string {
	v := ctx.Value(contextIDKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

type contextIDType struct{}

var contextIDKey = &contextIDType{}

//GetRID returns the contextIDKey by generating or using the existing one
func GetRID(ctx context.Context) string {
	v := IDFromContext(ctx)
	if v == "" {
		v = uuid.New().String()
		WithID(ctx, v)
	}

	return v
}
