package main

import (
	"sync"
	"time"
)

// Device represents a browser client connected to LAN Clipboard.
// It is the foundation for future PC-to-phone transfers.
type Device struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Connected bool      `json:"connected"`
	LastSeen  time.Time `json:"last_seen"`
}

// DeviceManager keeps track of connected browser devices.
// The current version prepares the data model; WebSocket registration will be
// integrated in the next change.
type DeviceManager struct {
	mu      sync.RWMutex
	devices map[string]*Device
}

func newDeviceManager() *DeviceManager {
	return &DeviceManager{devices: make(map[string]*Device)}
}
