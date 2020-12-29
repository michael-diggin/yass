package api

import (
	"context"

	"github.com/michael-diggin/yass/models"
)

// GrpcClient is the interface needed to communicate with the cache server
type GrpcClient interface {
	SetValue(context.Context, *models.Pair, models.Replica) error
	GetValue(context.Context, string, models.Replica) (*models.Pair, error)
	DelValue(context.Context, string, models.Replica) error
	SetFollowerValue(context.Context, string, interface{}) error
	GetFollowerValue(context.Context, string) (interface{}, error)
	DelFollowerValue(context.Context, string) error
	Close() error
}
