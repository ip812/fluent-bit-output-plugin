package main

import (
	"context"

	"github.com/go-logr/logr"
)

// OutputClient represents an instance which sends logs to OTLP backend
type OutputClient interface {
	// Handle processes logs and then sends them to OTLP backend
	Handle(log OutputEntry) error
	// Stop shut down the client immediately without waiting to send the saved logs
	Stop()
	// StopWait stops the client of receiving new logs and waits all saved logs to be sent until shutting down
	StopWait()
	// Endpoint returns the target logging backend endpoint
	Endpoint() string
}

// NewClientFunc is a function type for creating new OutputClient instances
type NewClientFunc func(ctx context.Context, cfg Config, logger logr.Logger) (OutputClient, error)
