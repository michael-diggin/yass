package api

import (
	"context"

	"github.com/michael-diggin/yass/models"
)

// GrpcClient is the interface needed to communicate with the cache server
type GrpcClient interface {
	SetValue(context.Context, *models.Pair) error
	GetValue(context.Context, string) (*models.Pair, error)
	DelValue(context.Context, string) error
	SetFollowerValue(context.Context, string, interface{}) error
	GetFollowerValue(context.Context, string) (interface{}, error)
	DelFollowerValue(context.Context, string) error
	Close() error
}
