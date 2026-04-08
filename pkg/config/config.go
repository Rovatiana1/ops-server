// Package config re-exports the root configs package so internal packages
// can import "ops-server/pkg/config" for a consistent import path.
package config

import "ops-server/configs"

// Aliases — keeps backward compatibility if import path changes.
type (
	Config              = configs.Config
	AppConfig           = configs.AppConfig
	DatabaseConfig      = configs.DatabaseConfig
	RedisConfig         = configs.RedisConfig
	KafkaConfig         = configs.KafkaConfig
	JWTConfig           = configs.JWTConfig
	RateLimitConfig     = configs.RateLimitConfig
	ObservabilityConfig = configs.ObservabilityConfig
)

// Load delegates to the root loader.
var Load = configs.Load
