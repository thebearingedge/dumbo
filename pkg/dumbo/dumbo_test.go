package dumbo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo/internal/dumbotest"
)

func TestInsertingRecords(t *testing.T) {
	db := dumbotest.RequireDB(t)

	t.Run("inserting one", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		inserted := InsertOne(t, tx, "user", Record{
			"username": "gopher",
		})

		assert.Equal(t, int64(1), inserted["id"])
		assert.Equal(t, "gopher", inserted["username"])
	})

	t.Run("inserting many", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		inserted := InsertMany(t, tx, "user", []Record{
			{
				"username": "gopher",
			},
			{
				"username": "rustacean",
			},
		})

		gopher, rustacean := inserted[0], inserted[1]

		assert.Equal(t, int64(1), gopher["id"])
		assert.Equal(t, "gopher", gopher["username"])
		assert.Equal(t, int64(2), rustacean["id"])
		assert.Equal(t, "rustacean", rustacean["username"])
	})
}
