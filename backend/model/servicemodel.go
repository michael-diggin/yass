package model

import (
	"context"
)

// Service defines the interface for getting and setting cache key/values
type Service interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, string) (string, error)
	Delete(context.Context, string) error
}
