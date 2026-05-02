package main

import (
	"C"
	"fmt"
	"unsafe"

	"log/slog"
	"time"

	"github.com/fluent/fluent-bit-go/output"
)

// FLBPluginRegister registers your plugin with Fluent Bit (name, description, callbacks)
//
//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "ip812", "output plugin")
}

// FLBPluginInit is called for each plugin instance
//
//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	return output.FLB_OK
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

		slog.Info("[%d] %s: [%s, {", count, C.GoString(tag), timestamp.String())
		handle(record, timestamp, tag)
		count++

	}
	return output.FLB_OK
}

func handle(record map[interface{}]interface{}, timestamp time.Time, tag *C.char) {
	for k, v := range record {
		fmt.Printf("\"%s\": %v, ", k, v)
	}
	fmt.Printf("}\n")
}

// FLBPluginExit is called when the plugin is being destroyed (global cleanup)
//
//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {}
