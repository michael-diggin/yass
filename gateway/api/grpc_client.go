package api

import "context"

// GrpcClient is the interface needed to communicate with the cache server
type GrpcClient interface {
	SetValue(context.Context, string, interface{}) error
	GetValue(context.Context, string) (interface{}, error)
	DelValue(context.Context, string) error
	SetFollowerValue(context.Context, string, interface{}) error
	GetFollowerValue(context.Context, string) (interface{}, error)
	DelFollowerValue(context.Context, string) error
	Close() error
}
