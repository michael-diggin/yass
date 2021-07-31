package kv

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrNotFound struct {
	Key string
}

// GRPCStatus implements the GRPC status interface
func (e ErrNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, fmt.Sprintf("no record found for key: %s", e.Key))
}

// Error implements the error interface
func (e ErrNotFound) Error() string {
	return e.GRPCStatus().Err().Error()
}
