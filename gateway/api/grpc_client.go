package api

import (
	"context"

	"github.com/michael-diggin/yass/models"
)

// GrpcClient is the interface needed to communicate with the cache server
type GrpcClient interface {
	Check(context.Context) (bool, error)
	SetValue(context.Context, *models.Pair, int) error
	GetValue(context.Context, string, int) (*models.Pair, error)
	DelValue(context.Context, string, int) error

	BatchSend(context.Context, int, int, string, uint32, uint32) error
	BatchDelete(context.Context, int, uint32, uint32) error

	Close() error
}
