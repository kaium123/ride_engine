package api

import "context"

// Server is the interface for the server
type Server interface {
	Name() string
	Run() error
	Shutdown(ctx context.Context) error
}
