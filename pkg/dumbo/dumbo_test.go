package dumbo

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thebearingedge/dumbo/internal/dumbotest"
)

func TestConnectToPostgres(t *testing.T) {
	db := dumbotest.RequireDB(t)
	err := db.Ping()
	require.NoError(t, err)
}
