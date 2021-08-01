package api

import (
	"fmt"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// ErrOffsetOutOfRange represents an error found when the offset
// is too large
type ErrOffsetOutOfRange struct {
	Offset uint64
}

// GRPCStatus implements the GRPC status interface
func (e ErrOffsetOutOfRange) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, fmt.Sprintf("offset out of range: %d", e.Offset))
}

// Error implements the error interface
func (e ErrOffsetOutOfRange) Error() string {
	return e.GRPCStatus().Err().Error()
}

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
