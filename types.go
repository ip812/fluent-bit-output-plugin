package main

import "time"

type OutputEntry struct {
	Timestamp time.Time
	Record    map[string]any
}

type OutputPlugin interface {
	SendRecord(log OutputEntry) error
	Close()
}
