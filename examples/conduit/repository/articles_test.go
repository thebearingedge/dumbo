package repository

import (
	"database/sql"
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

func TestUpdateArticle(t *testing.T) {
	db := conduittest.RequireDB(t)
	conduittest.RequireTruncate(t, db, "article_tags")
	conduittest.RequireTruncate(t, db, "tags")
	conduittest.RequireTruncate(t, db, "articles")

	t.Run("skips non-existent article", func(t *testing.T) {
		tx := conduittest.RequireBegin(t, db)
		user := conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{})

		articles := NewArticlesRepository(tx)
		patched, err := articles.PartialUpdate(schema.ArticlePatch{
			ID:          uint64(1),
			AuthorID:    uint64(user["id"].(int64)),
			Slug:        sql.NullString{String: "postgres-is-the-best", Valid: true},
			Title:       sql.NullString{String: "Postgres is the Best", Valid: true},
			Description: sql.NullString{String: "it's obvious", Valid: true},
			Body:        sql.NullString{String: "blah", Valid: true},
		})

		assert.NoError(t, err)
		assert.Nil(t, patched)
	})

	t.Run("updates the article", func(t *testing.T) {
		tx := conduittest.RequireBegin(t, db)
		user := conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{})
		published := conduittest.Seeder.SeedOne(t, tx, "articles", dumbo.Record{
			"author_id": user["id"],
		})

		articles := NewArticlesRepository(tx)
		patched, err := articles.PartialUpdate(schema.ArticlePatch{
			ID:       uint64(published["id"].(int64)),
			AuthorID: uint64(user["id"].(int64)),
			Slug:     sql.NullString{String: "postgres-is-just-ok", Valid: true},
			Title:    sql.NullString{String: "Postgres is Just OK", Valid: true},
		})

		assert.NoError(t, err)
		assert.Equal(t, patched.Author.Username, user["username"])
		assert.Equal(t, patched.Slug, "postgres-is-just-ok")
		assert.Equal(t, patched.Title, "Postgres is Just OK")
		assert.Equal(t, patched.Description, published["description"])
		assert.Equal(t, patched.Body, published["body"])
	})

	t.Run("skips duplicate slugs", func(t *testing.T) {
		tx := conduittest.RequireBegin(t, db)
		user := conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{})
		published := conduittest.Seeder.SeedOne(t, tx, "articles", dumbo.Record{
			"slug":      "postgres-is-just-ok",
			"author_id": user["id"],
		})

		articles := NewArticlesRepository(tx)
		patched, err := articles.PartialUpdate(schema.ArticlePatch{
			ID:       uint64(published["id"].(int64)),
			AuthorID: uint64(user["id"].(int64)),
			Slug:     sql.NullString{String: "postgres-is-just-ok", Valid: true},
		})

		assert.NoError(t, err)
		assert.Nil(t, patched)
	})
}
