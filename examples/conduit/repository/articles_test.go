package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo"
	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres/schema"
	"github.com/thebearingedge/dumbo/examples/conduit/internal/conduittest"
)

func TestPublishArticle(t *testing.T) {
	db := conduittest.RequireDB(t)
	conduittest.RequireTruncate(t, db, "article_tags")
	conduittest.RequireTruncate(t, db, "tags")
	conduittest.RequireTruncate(t, db, "articles")

	t.Run("saves the article", func(t *testing.T) {
		tx := conduittest.RequireBegin(t, db)
		user := conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{})

		articles := NewArticlesRepository(tx)

		published, err := articles.Publish(schema.NewArticle{
			AuthorID:    uint64(user["id"].(int64)),
			Slug:        "postgres-is-the-best",
			Title:       "Postgres is the best",
			Description: "it's obvious",
			Body:        "blah",
			TagList:     []string{"sql", "databases"},
		})

		assert.NoError(t, err)
		assert.Equal(t, "postgres-is-the-best", published.Slug)
		assert.Equal(t, "Postgres is the best", published.Title)
		assert.Equal(t, "it's obvious", published.Description)
		assert.Equal(t, "blah", published.Body)
		assert.Equal(t, []string{"sql", "databases"}, published.TagList)
		assert.False(t, published.Favorited)
		assert.Equal(t, uint(0), published.FavoritesCount)
		assert.NotEmpty(t, published.CreatedAt)
		assert.NotEmpty(t, published.UpdatedAt)
		assert.Equal(t, schema.Author{Username: user["username"].(string)}, published.Author)
	})

	t.Run("does not save duplicate articles", func(t *testing.T) {
		tx := conduittest.RequireBegin(t, db)
		user := conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{})

		articles := NewArticlesRepository(tx)

		_, _ = articles.Publish(schema.NewArticle{
			AuthorID:    uint64(user["id"].(int64)),
			Slug:        "postgres-is-the-best",
			Title:       "Postgres is the best",
			Description: "it's obvious",
			Body:        "blah",
			TagList:     []string{"sql", "databases"},
		})

		published, err := articles.Publish(schema.NewArticle{
			AuthorID:    uint64(user["id"].(int64)),
			Slug:        "postgres-is-the-best",
			Title:       "Postgres is the best",
			Description: "it's obvious",
			Body:        "blah",
			TagList:     []string{"sql", "databases"},
		})

		assert.NoError(t, err)
		assert.Nil(t, published)
	})
}
