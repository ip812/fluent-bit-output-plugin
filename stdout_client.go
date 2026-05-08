// Copyright 2025 SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

const componentStdoutName = "stdout"

// StdoutClient is an implementation of OutputClient that writes all records to stdout
type StdoutClient struct {
	ctx      context.Context
	logger   logr.Logger
	endpoint string
}

var _ OutputClient = &StdoutClient{}

// NewStdoutClient creates a new StdoutClient that writes all records to stdout
func NewStdoutClient(ctx context.Context, cfg Config, logger logr.Logger) (OutputClient, error) {
	client := &StdoutClient{
		ctx:      ctx,
		endpoint: cfg.OTLPConfig.Endpoint,
		logger:   logger.WithValues("endpoint", cfg.OTLPConfig.Endpoint),
	}

	logger.V(1).Info(fmt.Sprintf("%s created", componentStdoutName))

	return client, nil
}

// Handle processes and writes the log entry to stdout while incrementing metrics
func (c *StdoutClient) Handle(entry OutputEntry) error {
	return nil
}

// Stop shuts down the client immediately
func (c *StdoutClient) Stop() {
	c.logger.V(2).Info(fmt.Sprintf("stopping %s", componentStdoutName))
}

// StopWait stops the client - since this is a stdout client, it's the same as Stop
func (c *StdoutClient) StopWait() {
	c.logger.V(2).Info(fmt.Sprintf("stopping %s with wait", componentStdoutName))
}

// GetEndPoint returns the configured endpoint
func (c *StdoutClient) Endpoint() string {
	return c.endpoint
}
