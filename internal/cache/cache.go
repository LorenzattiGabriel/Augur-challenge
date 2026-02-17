package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/dgraph-io/ristretto"
)

type Cache struct {
	cache *ristretto.Cache
}

type Config struct {
	MaxSizeMB int64
}

func New(cfg Config) (*Cache, error) {
	maxCost := cfg.MaxSizeMB * 1024 * 1024
	numCounters := maxCost / 100

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &Cache{cache: cache}, nil
}

func (c *Cache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	cost := int64(1)
	c.cache.SetWithTTL(key, value, cost, ttl)
}

func (c *Cache) Delete(key string) {
	c.cache.Del(key)
}

func (c *Cache) Clear() {
	c.cache.Clear()
}

func GenerateKey(prefix string, params interface{}) string {
	data, _ := json.Marshal(params)
	hash := sha256.Sum256(data)
	return prefix + ":" + hex.EncodeToString(hash[:8])
}

const (
	TTLIndicatorDetail  = 2 * time.Minute
	TTLIndicatorSearch  = 30 * time.Second
	TTLCampaignTimeline = 1 * time.Minute
	TTLDashboardSummary = 5 * time.Minute
)
