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

func TestFindArticleBySlug(t *testing.T) {
	db := conduittest.RequireDB(t)

	tx := conduittest.RequireBegin(t, db)
	user := conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{})
	article := conduittest.Seeder.SeedOne(t, tx, "articles", dumbo.Record{
		"author_id": user["id"],
		"slug":      "postgres-rules",
		"title":     "Postgres Rules",
	})

	articles := NewArticlesRepository(tx)

	t.Run("finds existing articles", func(t *testing.T) {
		found, err := articles.FindBySlug(article["slug"].(string))
		assert.NoError(t, err)
		assert.Equal(t, "postgres-rules", found.Slug)
		assert.Equal(t, "Postgres Rules", found.Title)
	})

	t.Run("does not find non-existent articles", func(t *testing.T) {
		found, err := articles.FindBySlug("postgres-is-mid")
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestListArticlesReverseChronological(t *testing.T) {
	db := conduittest.RequireDB(t)

	tx := conduittest.RequireBegin(t, db)
	users := conduittest.Seeder.SeedMany(t, tx, "users", []dumbo.Record{
		{"username": "irwin"},
		{"username": "billy"},
	})
	seed := conduittest.Seeder.SeedMany(t, tx, "articles", []dumbo.Record{
		{
			"author_id": users[0]["id"],
			"slug":      "postgres-rules",
			"title":     "Postgres Rules",
		},
		{
			"author_id": users[1]["id"],
			"slug":      "postgres-sucks",
			"title":     "Postgres Sucks",
		},
		{
			"author_id": users[0]["id"],
			"slug":      "postgres-ok",
			"title":     "Postgres OK",
		},
	})
	tags := conduittest.Seeder.SeedMany(t, tx, "tags", []dumbo.Record{
		{"name": "good"},
		{"name": "bad"},
	})
	conduittest.Seeder.SeedMany(t, tx, "article_tags", []dumbo.Record{
		{"article_id": seed[0]["id"], "tag_id": tags[0]["id"]},
		{"article_id": seed[1]["id"], "tag_id": tags[1]["id"]},
		{"article_id": seed[2]["id"], "tag_id": tags[0]["id"]},
	})
	conduittest.Seeder.SeedMany(t, tx, "favorites", []dumbo.Record{
		{"article_id": seed[1]["id"], "user_id": users[1]["id"]},
		{"article_id": seed[2]["id"], "user_id": users[1]["id"]},
	})

	articles := NewArticlesRepository(tx)

	t.Run("filters by tag", func(t *testing.T) {
		good, err := articles.List(ArticleFilter{
			Tags: []string{"good"},
		})

		assert.NoError(t, err)
		assert.Len(t, good.Articles, 2)
		assert.Equal(t, "postgres-ok", good.Articles[0].Slug)
		assert.Equal(t, "postgres-rules", good.Articles[1].Slug)

		bad, err := articles.List(ArticleFilter{
			Tags: []string{"bad"},
		})

		assert.NoError(t, err)
		assert.Len(t, bad.Articles, 1)
		assert.Equal(t, "postgres-sucks", bad.Articles[0].Slug)

		all, err := articles.List(ArticleFilter{})

		assert.NoError(t, err)
		assert.Len(t, all.Articles, 3)
		assert.Equal(t, "postgres-ok", all.Articles[0].Slug)
		assert.Equal(t, "postgres-sucks", all.Articles[1].Slug)
		assert.Equal(t, "postgres-rules", all.Articles[2].Slug)

		alsoAll, err := articles.List(ArticleFilter{
			Tags: []string{"good", "bad"},
		})

		assert.NoError(t, err)
		assert.Len(t, alsoAll.Articles, 3)
		assert.Equal(t, "postgres-ok", alsoAll.Articles[0].Slug)
		assert.Equal(t, "postgres-sucks", alsoAll.Articles[1].Slug)
		assert.Equal(t, "postgres-rules", alsoAll.Articles[2].Slug)
	})

	t.Run("filters by author username", func(t *testing.T) {
		byIrwin, err := articles.List(ArticleFilter{
			Author: "irwin",
		})

		assert.NoError(t, err)
		assert.Len(t, byIrwin.Articles, 2)
		assert.Equal(t, "irwin", byIrwin.Articles[0].Author.Username)
		assert.Equal(t, "irwin", byIrwin.Articles[1].Author.Username)
	})

	t.Run("filters by user favorite", func(t *testing.T) {
		billyFavs, err := articles.List(ArticleFilter{
			Favorited: "billy",
		})

		assert.NoError(t, err)
		assert.Len(t, billyFavs.Articles, 2)
		assert.Equal(t, "postgres-ok", billyFavs.Articles[0].Slug)
		assert.Equal(t, "postgres-sucks", billyFavs.Articles[1].Slug)
	})

	t.Run("skips by offset", func(t *testing.T) {
		skipFirst, err := articles.List(ArticleFilter{
			Offset: 1,
		})

		assert.NoError(t, err)
		assert.Len(t, skipFirst.Articles, 2)
		assert.Equal(t, "postgres-sucks", skipFirst.Articles[0].Slug)
		assert.Equal(t, "postgres-rules", skipFirst.Articles[1].Slug)
	})
}

func TestDeleteArticleBySlug(t *testing.T) {
	db := conduittest.RequireDB(t)
	tx := conduittest.RequireBegin(t, db)

	user := conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{})
	published := conduittest.Seeder.SeedOne(t, tx, "articles", dumbo.Record{
		"author_id": user["id"],
	})

	t.Run("deletes the target article", func(t *testing.T) {
		sp := conduittest.RequireSavepoint(t, tx)
		articles := NewArticlesRepository(sp)

		deleted, err := articles.DeleteBySlug(user["id"].(int64), published["slug"].(string))
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)

		remaining := conduittest.Seeder.FetchMany(t, sp, `
			select *
			  from articles
			 where slug = $1
		`, published["slug"])

		assert.Len(t, remaining, 0)
	})

	t.Run("only deletes the target article", func(t *testing.T) {
		sp := conduittest.RequireSavepoint(t, tx)
		articles := NewArticlesRepository(sp)

		deleted, err := articles.DeleteBySlug(
			user["id"].(int64),
			published["slug"].(string)+"-foo",
		)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), deleted)

		article := conduittest.Seeder.FetchOne(t, sp, `
			select *
			  from articles
			 where id = $1
		`, published["id"])

		assert.Equal(t, published["slug"], article["slug"])
	})

	t.Run("only deletes the article owned by user", func(t *testing.T) {
		sp := conduittest.RequireSavepoint(t, tx)
		articles := NewArticlesRepository(sp)

		deleted, err := articles.DeleteBySlug(
			int64(2),
			published["slug"].(string)+"-foo",
		)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), deleted)

		article := conduittest.Seeder.FetchOne(t, sp, `
			select *
			  from articles
			 where id = $1
		`, published["id"])

		assert.Equal(t, published["slug"], article["slug"])
	})
}
