package api

import (
	"context"

	"github.com/michael-diggin/yass/models"
)

// GrpcClient is the interface needed to communicate with the cache server
type GrpcClient interface {
	Check(context.Context) (bool, error)
	SetValue(context.Context, *models.Pair, models.Replica) error
	GetValue(context.Context, string, models.Replica) (*models.Pair, error)
	DelValue(context.Context, string, models.Replica) error
	Close() error
}
