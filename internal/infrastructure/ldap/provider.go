package ldap

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ops-server/internal/domain/config_general/service"

	redisInfra "ops-server/internal/infrastructure/redis"
)

type ConfigProvider interface {
	Get(ctx context.Context) (*Config, error)
}

type configProvider struct {
	configSvc service.ConfigGeneralService
	cache     redisInfra.Cache // interface Redis
	ttl       time.Duration
}

func NewConfigProvider(svc service.ConfigGeneralService, cache redisInfra.Cache) ConfigProvider {
	return &configProvider{
		configSvc: svc,
		cache:     cache,
		ttl:       5 * time.Minute,
	}
}

func (p *configProvider) Get(ctx context.Context) (*Config, error) {
	const cacheKey = "config:ldap_entity:ldap"

	// 1. cache
	if val, err := p.cache.Get(ctx, cacheKey); err == nil {
		var cfg Config
		if err := json.Unmarshal([]byte(val), &cfg); err == nil {
			return &cfg, nil
		}
	}

	// 2. DB
	conf, err := p.configSvc.GetByEntityKey(ctx, "ldap_entity", "ldap")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(conf.Data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid ldap config: %w", err)
	}

	// 3. cache set
	_ = p.cache.Set(ctx, cacheKey, string(conf.Data), p.ttl)

	return &cfg, nil
}
