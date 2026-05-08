package main

import (
	"C"
	"fmt"
	"unsafe"

	"time"

	"strings"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/util/uuid"
)
import "sync"

var (
	logger  logr.Logger
	plugins sync.Map // map[string]OutputPlugin
)

// FLBPluginInit is called for each plugin instance
//
//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	// Check if the plugin was already initialized
	pluginID := output.FLBPluginGetContext(ctx)
	_, okPlugin := plugins.Load(pluginID.(string))
	if pluginID != nil && okPlugin {
		logger.Info("[flb-go]", "output plugin is already present")
		return output.FLB_OK
	}

	// Load the plugin's configuration
	cfg, err := NewConfig(ctx)
	if err != nil {
		logger.Info("[flb-go] failed to launch", "error", err)
		return output.FLB_ERROR
	}
	logger = NewLogger(cfg.PluginConfig.LogLevel)
	cfg.Dump()

	// Initialize the new plugin
	id, _, _ := strings.Cut(string(uuid.NewUUID()), "-")
	outputPlugin, err := NewPlugin(id, NewLogger(cfg.PluginConfig.LogLevel), cfg)
	if err != nil {
		logger.Error(err, "[flb-go] error creating output plugin", "id", id)
		return output.FLB_ERROR
	}
	output.FLBPluginSetContext(ctx, id)
	plugins.Store(id, outputPlugin)

	pluginsLen := 0
	plugins.Range(func(_, _ any) bool {
		pluginsLen++
		return true
	})
	logger.Info(
		"[flb-go] output plugin initialized",
		"id", id,
		"count", pluginsLen,
	)

	return output.FLB_OK
}

// FLBPluginRegister registers your plugin with Fluent Bit (name, description, callbacks)
//
//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "ip812", "output plugin")
}

// FLBPluginFlushCtx is called when the plugin is invoked to flush data
//
//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
	dec := output.NewDecoder(data, int(length))

	for {
		ret, ts, record := output.GetRecord(dec)
		if ret != 0 {
			break
		}

		count := 0
		var timestamp time.Time
		switch t := ts.(type) {
		case output.FLBTime:
			timestamp = ts.(output.FLBTime).Time
		case uint64:
			timestamp = time.Unix(int64(t), 0)
		default:
			fmt.Println("time provided invalid, defaulting to now.")
			timestamp = time.Now()
		}

		handle(record, timestamp, tag)
		count++

	}
	return output.FLB_OK
}

func handle(record map[interface{}]interface{}, timestamp time.Time, tag *C.char) {
	for k, v := range record {
		logger.Info(
			fmt.Sprintf("%s: %s", k, v),
		)
	}
}

// FLBPluginExit is called when the plugin is being destroyed (global cleanup)
//
//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {}
