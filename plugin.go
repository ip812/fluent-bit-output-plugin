package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

// Type represents the type of OutputClient
type Type int

const (
	// NOOP type represents a no-operation client type
	NOOP Type = iota
	// STDOUT type represents a standard output client type
	STDOUT
	// OTLPGRPC type represents an OTLP gRPC client type
	OTLPGRPC
	// OTLPHTTP type represents an OTLP HTTP client type
	OTLPHTTP
	// UNKNOWN type represents an unknown client type
	UNKNOWN
)

// ClientTypeFromString converts a string representation of client type to Type. It returns NOOP for unknown types.
func ClientTypeFromString(clientType string) Type {
	switch strings.ToUpper(clientType) {
	case "NOOP":
		return NOOP
	case "STDOUT":
		return STDOUT
	case "OTLPGRPC", "OTLP_GRPC":
		return OTLPGRPC
	case "OTLPHTTP", "OTLP_HTTP":
		return OTLPHTTP
	default:
		return NOOP
	}
}

// String returns the string representation of the client Type
func (t Type) String() string {
	switch t {
	case NOOP:
		return "noop"
	case STDOUT:
		return "stdout"
	case OTLPGRPC:
		return "otlp_grpc"
	case OTLPHTTP:
		return "otlp_http"
	case UNKNOWN:
		return "unknown"
	default:
		return ""
	}
}

type OutputEntry struct {
	Timestamp time.Time
	Record    map[string]any
}

type OutputPlugin interface {
	SendRecord(log OutputEntry) error
	Close()
}

type Plugin struct {
	id     string
	ctx    context.Context
	cancel context.CancelFunc
	logger logr.Logger
	cfg    *Config
	client OutputClient
}

func NewPlugin(id string, logger logr.Logger, cfg *Config) (OutputPlugin, error) {
	ctx, cancel := context.WithCancel(context.Background())

	clientType := ClientTypeFromString(cfg.PluginConfig.ClientType)
	var ncf NewClientFunc
	switch clientType {
	// case OTLPGRPC:
	// 	newClientFunc = NewOTLPGRPCClient
	// case OTLPHTTP:
	// 	newClientFunc = NewOTLPHTTPClient
	case STDOUT:
		ncf = NewStdoutClient
	case NOOP:
		ncf = NewNoopClient
	default:
		cancel()
		return nil, fmt.Errorf("unknown client type: %v", clientType)
	}

	// Create a single context for the entire plugin lifecycle
	client, err := ncf(ctx, *cfg, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("can't create client")
	}

	return &Plugin{
		id:     id,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
		cfg:    cfg,
		client: client,
	}, nil
}

func (p *Plugin) SendRecord(log OutputEntry) error {
	record := log.Record
	if len(record) == 0 {
		p.logger.Info("no record left after removing keys")
		return nil
	}

	return p.client.Handle(log)
}

func (p *Plugin) Close() {
	// Cancel the plugin context first to signal all operations to stop
	p.cancel()
	p.client.StopWait()
	p.logger.Info("logging plugin stopped")
}
