package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionNilPool(t *testing.T) {
	conn := &Connection{}

	_, err := conn.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")

	_, err = conn.Exec(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")

	_, err = conn.BeginTx(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")

	stats := conn.Stats()
	assert.Nil(t, stats)

	pool := conn.Pool()
	assert.Nil(t, pool)
}
