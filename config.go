package main

import "time"

type Config struct {
	APIKey            string
	Endpoint          string
	Reconnect         bool
	HeartbeatInterval time.Duration
}

func DefaultConfig() Config {
	return Config{
		Endpoint:          "ws://localhost:8080/ws",
		Reconnect:         true,
		HeartbeatInterval: 10 * time.Second,
	}
}
