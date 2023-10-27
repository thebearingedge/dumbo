package repository

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/db/schema"
	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/dbtest"
)

func TestPublishArticle(t *testing.T) {
	db := dbtest.RequireDB(t)
	tx := dbtest.RequireBegin(t, db)

	dbtest.RequireExec(t, tx, `
		truncate table "users" restart identity cascade;

		insert into "users" ("username", "email", "password")
		values ('gopher', 'gopher@google.com', 'this should be hashed')
	`)

	const userID = uint64(1)

	t.Run("saves the article", func(t *testing.T) {
		sp := dbtest.RequireSavepoint(t, tx)

		articles := NewArticlesRepository(sp)

		published, err := articles.Publish(schema.NewArticle{
			AuthorID:    userID,
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
		assert.Equal(t, schema.Author{Username: "gopher"}, published.Author)
	})

	t.Run("does not save duplicate articles", func(t *testing.T) {

		articles := NewArticlesRepository(tx)

		_, _ = articles.Publish(schema.NewArticle{
			AuthorID:    userID,
			Slug:        "postgres-is-the-best",
			Title:       "Postgres is the best",
			Description: "it's obvious",
			Body:        "blah",
			TagList:     []string{"sql", "databases"},
		})

		unpublished, err := articles.Publish(schema.NewArticle{
			AuthorID:    userID,
			Slug:        "postgres-is-the-best",
			Title:       "Postgres is the best",
			Description: "it's obvious",
			Body:        "blah",
			TagList:     []string{"sql", "databases"},
		})

		assert.NoError(t, err)
		assert.Nil(t, unpublished)
	})
}

func TestUpdateArticle(t *testing.T) {
	db := dbtest.RequireDB(t)
	tx := dbtest.RequireBegin(t, db)

	dbtest.RequireExec(t, tx, `
		truncate table "users" restart identity cascade;

		insert into "users" ("username", "email", "password")
		values ('gopher', 'gopher@google.com', 'this should be hashed');
	`)

	const userID = uint64(1)
	const articleID = uint64(1)

	t.Run("skips non-existent article", func(t *testing.T) {

		articles := NewArticlesRepository(tx)

		unpatched, err := articles.PartialUpdate(schema.ArticlePatch{
			ID:          articleID,
			AuthorID:    userID,
			Slug:        sql.NullString{String: "postgres-is-the-best", Valid: true},
			Title:       sql.NullString{String: "Postgres is the Best", Valid: true},
			Description: sql.NullString{String: "it's obvious", Valid: true},
			Body:        sql.NullString{String: "blah", Valid: true},
		})

		assert.NoError(t, err)
		assert.Nil(t, unpatched)
	})

	t.Run("updates the article", func(t *testing.T) {
		sp := dbtest.RequireSavepoint(t, tx)

		dbtest.RequireExec(t, sp, `
			truncate table "articles" restart identity cascade;

			insert into "articles" ("author_id", "slug", "title", "description", "body")
			values (1, 'postgres-is-the-best', 'Postgres is the Best', 'it''s obvious', 'blah');
		`)

		articles := NewArticlesRepository(sp)

		patched, err := articles.PartialUpdate(schema.ArticlePatch{
			ID:       articleID,
			AuthorID: userID,
			Slug:     sql.NullString{String: "postgres-is-just-ok", Valid: true},
			Title:    sql.NullString{String: "Postgres is Just OK", Valid: true},
		})

		assert.NoError(t, err)
		assert.NotNil(t, patched)
		assert.Equal(t, "gopher", patched.Author.Username)
		assert.Equal(t, "postgres-is-just-ok", patched.Slug)
		assert.Equal(t, "Postgres is Just OK", patched.Title)
		assert.Equal(t, "it's obvious", patched.Description)
		assert.Equal(t, "blah", patched.Body)
	})

	t.Run("skips duplicate slugs", func(t *testing.T) {
		t.Skip() // this is wrong
		sp := dbtest.RequireSavepoint(t, tx)

		dbtest.RequireExec(t, sp, `
			truncate table "articles" restart identity cascade;

			insert into "articles" ("author_id", "slug", "title", "description", "body")
			values (1, 'postgres-is-the-best', 'Postgres is the Best', 'it''s obvious', 'blah');
		`)

		const userID = uint64(1)
		const articleID = uint64(1)

		articles := NewArticlesRepository(sp)
		unpatched, err := articles.PartialUpdate(schema.ArticlePatch{
			ID:       articleID,
			AuthorID: userID,
			Slug:     sql.NullString{String: "postgres-is-the-best", Valid: true},
		})

		assert.NoError(t, err)
		assert.Nil(t, unpatched)
	})
}

func TestFindArticleBySlug(t *testing.T) {
	db := dbtest.RequireDB(t)
	tx := dbtest.RequireBegin(t, db)

	dbtest.RequireExec(t, tx, `
		truncate table "users" restart identity cascade;

		insert into "users" ("username", "email", "password")
		values ('gopher', 'gopher@google.com', 'this should be hashed');

		insert into "articles" ("author_id", "slug", "title", "description", "body")
		values (1, 'postgres-is-the-best', 'Postgres is the Best', 'it''s obvious', 'blah');
	`)

	articles := NewArticlesRepository(tx)

	t.Run("finds existing articles", func(t *testing.T) {
		found, err := articles.FindBySlug("postgres-is-the-best")

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "postgres-is-the-best", found.Slug)
		assert.Equal(t, "Postgres is the Best", found.Title)
	})

	t.Run("does not find non-existent articles", func(t *testing.T) {
		notFound, err := articles.FindBySlug("postgres-is-mid")

		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})
}

func TestListArticlesReverseChronological(t *testing.T) {
	db := dbtest.RequireDB(t)
	tx := dbtest.RequireBegin(t, db)

	dbtest.RequireScript(t, tx, "./seeds/articles.list.sql")

	articles := NewArticlesRepository(tx)

	t.Run("filters by tag", func(t *testing.T) {
		good, err := articles.List(ArticlesFilter{
			Tags: []string{"good"},
		})

		assert.NoError(t, err)
		assert.Len(t, good.Articles, 2)
		assert.Equal(t, "postgres-ok", good.Articles[0].Slug)
		assert.Equal(t, "postgres-rules", good.Articles[1].Slug)

		bad, err := articles.List(ArticlesFilter{
			Tags: []string{"bad"},
		})

		assert.NoError(t, err)
		assert.Len(t, bad.Articles, 1)
		assert.Equal(t, "postgres-sucks", bad.Articles[0].Slug)

		all, err := articles.List(ArticlesFilter{})

		assert.NoError(t, err)
		assert.Len(t, all.Articles, 3)
		assert.Equal(t, "postgres-ok", all.Articles[0].Slug)
		assert.Equal(t, "postgres-sucks", all.Articles[1].Slug)
		assert.Equal(t, "postgres-rules", all.Articles[2].Slug)

		alsoAll, err := articles.List(ArticlesFilter{
			Tags: []string{"good", "bad"},
		})

		assert.NoError(t, err)
		assert.Len(t, alsoAll.Articles, 3)
		assert.Equal(t, "postgres-ok", alsoAll.Articles[0].Slug)
		assert.Equal(t, "postgres-sucks", alsoAll.Articles[1].Slug)
		assert.Equal(t, "postgres-rules", alsoAll.Articles[2].Slug)
	})

	t.Run("filters by author username", func(t *testing.T) {
		byErwin, err := articles.List(ArticlesFilter{
			Author: "erwin",
		})

		assert.NoError(t, err)
		assert.Len(t, byErwin.Articles, 2)
		assert.Equal(t, "erwin", byErwin.Articles[0].Author.Username)
		assert.Equal(t, "erwin", byErwin.Articles[1].Author.Username)
	})

	t.Run("filters by user favorite", func(t *testing.T) {
		billyFavs, err := articles.List(ArticlesFilter{
			Favorited: "billy",
		})

		assert.NoError(t, err)
		assert.Len(t, billyFavs.Articles, 2)
		assert.Equal(t, "postgres-ok", billyFavs.Articles[0].Slug)
		assert.Equal(t, "postgres-sucks", billyFavs.Articles[1].Slug)
	})

	t.Run("skips by offset", func(t *testing.T) {
		skipFirst, err := articles.List(ArticlesFilter{
			Offset: 1,
		})

		assert.NoError(t, err)
		assert.Len(t, skipFirst.Articles, 2)
		assert.Equal(t, "postgres-sucks", skipFirst.Articles[0].Slug)
		assert.Equal(t, "postgres-rules", skipFirst.Articles[1].Slug)
	})
}

func TestDeleteArticleBySlug(t *testing.T) {
	db := dbtest.RequireDB(t)
	tx := dbtest.RequireBegin(t, db)

	dbtest.RequireExec(t, tx, `
		truncate table "users" restart identity cascade;

		insert into "users" ("username", "email", "password")
		values ('gopher', 'gopher@google.com', 'this should be hashed');

		insert into "articles" ("author_id", "slug", "title", "description", "body")
		values (1, 'postgres-is-the-best', 'Postgres is the Best', 'it''s obvious', 'blah');
	`)

	const userID = uint64(1)
	const articleID = uint64(1)

	t.Run("deletes the target article", func(t *testing.T) {
		sp := dbtest.RequireSavepoint(t, tx)

		articles := NewArticlesRepository(sp)

		deleted, err := articles.DeleteBySlug(userID, "postgres-is-the-best")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)

		undeleted := dbtest.RequireRows(t, sp, `select * from "articles"`)
		assert.Len(t, undeleted, 0)
	})

	t.Run("only deletes the target article", func(t *testing.T) {
		sp := dbtest.RequireSavepoint(t, tx)
		articles := NewArticlesRepository(sp)

		deleted, err := articles.DeleteBySlug(userID, "postgres-is-the-best-foo")
		assert.NoError(t, err)
		assert.Equal(t, int64(0), deleted)

		undeleted := dbtest.RequireRows(t, sp, `select * from "articles"`)
		assert.Len(t, undeleted, 1)
		assert.Equal(t, int64(1), undeleted[0]["id"].(int64))
	})

	t.Run("only deletes the article owned by user", func(t *testing.T) {
		sp := dbtest.RequireSavepoint(t, tx)
		articles := NewArticlesRepository(sp)

		deleted, err := articles.DeleteBySlug(userID+1, "postgres-is-the-best-foo")
		assert.NoError(t, err)
		assert.Equal(t, int64(0), deleted)

		undeleted := dbtest.RequireRows(t, sp, `select * from "articles"`)
		assert.Len(t, undeleted, 1)
		assert.Equal(t, int64(1), undeleted[0]["id"].(int64))
	})
}
