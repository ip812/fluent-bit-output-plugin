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
	var newClientFunc NewClientFunc
	switch clientType {
	// case OTLPGRPC:
	// 	newClientFunc = NewOTLPGRPCClient
	// case OTLPHTTP:
	// 	newClientFunc = NewOTLPHTTPClient
	case STDOUT:
		newClientFunc = NewStdoutClient
	case NOOP:
		newClientFunc = NewNoopClient
	default:
		return nil, fmt.Errorf("unknown client type: %v", clientType)
	}

	client, err := newClientFunc(ctx, *cfg, logger)
	if err != nil {
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

func getNewClientFunc(t Type) (NewClientFunc, error) {
	switch t {
	// case OTLPGRPC:
	// 	return NewOTLPGRPCClient, nil
	// case OTLPHTTP:
	// 	return NewOTLPHTTPClient, nil
	case STDOUT:
		return NewStdoutClient, nil
	case NOOP:
		return NewNoopClient, nil
	default:
		return nil, fmt.Errorf("unknown client type: %v", t)
	}
}

func (p *Plugin) SendRecord(log OutputEntry) error {
	return nil
}

func (p *Plugin) Close() {
}
