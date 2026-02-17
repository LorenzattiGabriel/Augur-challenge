package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_SetAndGet(t *testing.T) {
	c, err := New(Config{MaxSizeMB: 10})
	require.NoError(t, err)

	c.Set("test-key", "test-value", time.Minute)

	time.Sleep(10 * time.Millisecond)

	value, found := c.Get("test-key")

	assert.True(t, found)
	assert.Equal(t, "test-value", value)
}

func TestCache_GetMiss(t *testing.T) {
	c, err := New(Config{MaxSizeMB: 10})
	require.NoError(t, err)

	_, found := c.Get("non-existent")

	assert.False(t, found)
}

func TestCache_Delete(t *testing.T) {
	c, err := New(Config{MaxSizeMB: 10})
	require.NoError(t, err)

	c.Set("delete-key", "value", time.Minute)
	time.Sleep(10 * time.Millisecond)

	c.Delete("delete-key")
	time.Sleep(10 * time.Millisecond)

	_, found := c.Get("delete-key")
	assert.False(t, found)
}

func TestCache_Clear(t *testing.T) {
	c, err := New(Config{MaxSizeMB: 10})
	require.NoError(t, err)

	c.Set("key1", "value1", time.Minute)
	c.Set("key2", "value2", time.Minute)
	time.Sleep(10 * time.Millisecond)

	c.Clear()
	time.Sleep(10 * time.Millisecond)

	_, found1 := c.Get("key1")
	_, found2 := c.Get("key2")

	assert.False(t, found1)
	assert.False(t, found2)
}

func TestGenerateKey(t *testing.T) {
	params1 := map[string]interface{}{"id": "123", "type": "ip"}
	params2 := map[string]interface{}{"id": "123", "type": "ip"}
	params3 := map[string]interface{}{"id": "456", "type": "ip"}

	key1 := GenerateKey("indicator", params1)
	key2 := GenerateKey("indicator", params2)
	key3 := GenerateKey("indicator", params3)

	assert.Equal(t, key1, key2)
	assert.NotEqual(t, key1, key3)
	assert.Contains(t, key1, "indicator:")
}

func TestTTLConstants(t *testing.T) {
	assert.Equal(t, 2*time.Minute, TTLIndicatorDetail)
	assert.Equal(t, 30*time.Second, TTLIndicatorSearch)
	assert.Equal(t, 1*time.Minute, TTLCampaignTimeline)
	assert.Equal(t, 5*time.Minute, TTLDashboardSummary)
}
