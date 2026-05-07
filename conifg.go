package main

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/go-viper/mapstructure/v2"
)

type PluginConfig struct {
	LogLevel string `mapstructure:"LogLevel"`
}

type OTLPConfig struct {
	Endpoint string `mapstructure:"Endpoint"`
}

type Config struct {
	PluginConfig PluginConfig `mapstructure:",squash"`
	OTLPConfig   OTLPConfig   `mapstructure:",squash"`
}

func (conf *Config) Dump() {
	logger.V(1).Info("[flb-go] =====   Plugin Config   =====")

	logger.V(1).Info("[flb-go]", "LogLevel", conf.PluginConfig.LogLevel)
	logger.V(1).Info("")
	logger.V(1).Info("[flb-go] =====   OTLP Config   =====")
	// OTLP general configuration
	logger.V(1).Info("[flb-go]", "Endpoint", fmt.Sprintf("%+v", conf.OTLPConfig.Endpoint))
}

func defaultConfig() *Config {
	return &Config{
		PluginConfig: PluginConfig{
			LogLevel: "info",
		},
		OTLPConfig: OTLPConfig{
			Endpoint: "localhost:4317",
		},
	}
}

// This is necessary because there is no direct C interface to retrieve the complete plugin configuration at once.
//
// When adding new configuration options to the plugin, the corresponding keys must be
// added to the configKeys slice below to ensure they are properly extracted.
func NewConfig(ctx unsafe.Pointer) (*Config, error) {
	rawCfg := rawConfig(ctx)
	// We intentionally call strings.ToLower twice.
	// Once https://github.com/fluent/fluent-bit/issues/11776 is resolved,
	// we will retrieve the configuration directly from the Fluent Bit C API at
	// once without calling the need to lower the characters.
	normCfg := normalizeConfigMapKeys(rawCfg)
	sanitizeConfigMap(normCfg)
	cfg, err := decodeConfig(normCfg)
	if err != nil {
		return nil, err
	}

	return cfg, err
}

func rawConfig(ctx unsafe.Pointer) map[string]any {
	raw := make(map[string]string)

	// Define all possible configuration keys based on the structs and documentation
	configKeys := []string{
		// General config
		"LogLevel", "logLevel", "log_level",

		// Common OTLP configs
		"Endpoint", "endpoint",
	}

	for _, key := range configKeys {
		if value := output.FLBPluginConfigKey(ctx, key); value != "" {
			raw[strings.ToLower(strings.ReplaceAll(key, "_", ""))] = value
		}
	}

	interfaceMap := make(map[string]any)
	for k, v := range raw {
		interfaceMap[k] = v
	}

	return interfaceMap
}

// normalizeConfigMapKeys converts all keys in the configuration map to lowercase
// This ensures case-insensitive configuration key matching throughout the codebase
func normalizeConfigMapKeys(configMap map[string]any) map[string]any {
	normalized := make(map[string]any, len(configMap))

	for key, value := range configMap {
		lowerKey := strings.ToLower(key)

		// Recursively normalize nested maps
		switch v := value.(type) {
		case map[string]any:
			normalized[lowerKey] = normalizeConfigMapKeys(v)
		default:
			normalized[lowerKey] = value
		}
	}

	return normalized
}

// sanitizeConfigMap recursively sanitizes all string values in the configuration map
func sanitizeConfigMap(configMap map[string]any) {
	for key, value := range configMap {
		//nolint:revive // enforce-switch-style: default-case is omitted on purpose
		switch v := value.(type) {
		case string:
			// Remove leading and trailing whitespace first
			v = strings.TrimSpace(v)
			configMap[key] = v

			// Remove surrounding double quotes
			if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
				configMap[key] = v[1 : len(v)-1]
			}

			// Remove surrounding single quotes
			if len(v) >= 2 && v[0] == '\'' && v[len(v)-1] == '\'' {
				configMap[key] = v[1 : len(v)-1]
			}
		case map[string]any:
			sanitizeConfigMap(v)
		}
	}
}

func decodeConfig(configMap map[string]any) (*Config, error) {
	config := defaultConfig()

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToBoolHookFunc(),
			mapstructure.StringToIntHookFunc(),
		),
		WeaklyTypedInput: true,
		Result:           config,
		TagName:          "mapstructure",
		// Ignore fields that need custom processing
		IgnoreUntaggedFields: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create mapstructure decoder: %w", err)
	}

	// Decode the configuration
	if err = decoder.Decode(configMap); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	// Apply custom processing for complex fields that can't be handled by mapstructure
	if err = postProcessConfig(config, configMap); err != nil {
		return nil, fmt.Errorf("failed to post-process config: %w", err)
	}

	return config, nil

}

// postProcessConfig handles complex field processing that can't be done with simple mapping
func postProcessConfig(config *Config, configMap map[string]any) error {
	processors := []func(*Config, map[string]any) error{
		processOTLPConfig,
		processLogLevel,
	}

	for _, processor := range processors {
		if err := processor(config, configMap); err != nil {
			return err
		}
	}

	return nil
}

func processOTLPConfig(config *Config, configMap map[string]any) error {
	if endpoint, ok := configMap["endpoint"].(string); ok && endpoint != "" {
		config.OTLPConfig.Endpoint = endpoint
	}

	return nil
}

func processLogLevel(config *Config, configMap map[string]any) error {
	if logLevel, ok := configMap["loglevel"].(string); ok && logLevel != "" {
		config.PluginConfig.LogLevel = logLevel
	}

	return nil
}
