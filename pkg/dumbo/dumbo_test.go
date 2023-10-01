package dumbo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo/internal/dumbotest"
)

func TestInsertOne(t *testing.T) {
	db := dumbotest.RequireDB(t)
	tx := dumbotest.RequireBegin(t, db)
	inserted := InsertOne(t, tx, "user", map[string]any{
		"username": "gopher",
	})
	assert.Equal(t, int64(1), inserted["id"])
	assert.Equal(t, "gopher", inserted["username"])
}
