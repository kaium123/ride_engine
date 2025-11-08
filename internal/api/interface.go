// Package api provides the API server for the application.
package api

import "context"

// Server is the interface for the server
type Server interface {
	Name() string
	Run() error
	Shutdown(ctx context.Context) error
}
