package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoOpCache_Behavior(t *testing.T) {
	ctx := context.Background()
	c := NewNoOpCache()

	// Get should always return ErrCacheMiss
	var dest interface{}
	err := c.Get(ctx, "key", &dest)
	assert.ErrorIs(t, err, ErrCacheMiss)

	// Set should return nil
	err = c.Set(ctx, "key", "value", time.Minute)
	assert.NoError(t, err)

	// Delete should return nil
	err = c.Delete(ctx, "key")
	assert.NoError(t, err)

	// DeletePattern should return nil
	err = c.DeletePattern(ctx, "pattern*")
	assert.NoError(t, err)

	// Exists should return false, nil
	exists, err := c.Exists(ctx, "key")
	assert.NoError(t, err)
	assert.False(t, exists)

	// TTL should return 0, nil
	ttl, err := c.TTL(ctx, "key")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), ttl)

	// Close should return nil
	err = c.Close()
	assert.NoError(t, err)

	// MGet should return nil
	err = c.MGet(ctx, []string{"key1", "key2"}, &dest)
	assert.NoError(t, err)

	// MSet should return nil
	err = c.MSet(ctx, map[string]interface{}{"key": "value"}, time.Minute)
	assert.NoError(t, err)

	// Ping should return nil
	err = c.Ping(ctx)
	assert.NoError(t, err)
}
