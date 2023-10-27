package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/db/schema"
	"github.com/thebearingedge/dumbo/examples/conduit/repository/service"
)

type ArticlesRepository struct {
	service service.DB
}

func NewArticlesRepository(service service.DB) *ArticlesRepository {
	return &ArticlesRepository{service: service}
}

func (r *ArticlesRepository) Publish(n schema.NewArticle) (*schema.Article, error) {

	query := `
		with inserted_article as (
		  insert into articles (author_id, slug, title, description, body)
		  values ($1, $2, $3, $4, $5)
		      on conflict (slug) do nothing
		  returning id,
		            slug
		), upserted_tags as (
		  insert into tags (name)
		  select unnest($6::text[])
		  on conflict (name) do nothing
		  returning id,
		            name
		), applied_tags as (
		  select id,
		         name
		    from tags
		   where name = any($6::text[])
		   union
		  select id,
		         name
		    from upserted_tags
		   order by id
		), inserted_article_tags as (
		  insert into article_tags (article_id, tag_id)
		  select (select id from inserted_article),
		         id
		    from applied_tags
		   where exists (select id from inserted_article)
		  returning article_id,
		            tag_id
		)
		select slug
		  from inserted_article
	`

	row := r.service.QueryRow(query, n.AuthorID, n.Slug, n.Title, n.Description, n.Body, pq.Array(n.TagList))

	var slug string
	if err := row.Scan(&slug); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf(`scanning published article: %w`, err)
	}

	return r.FindBySlug(slug)
}

func (r *ArticlesRepository) PartialUpdate(p schema.ArticlePatch) (*schema.Article, error) {

	query := `
		update articles
		   set slug        = coalesce($3, slug),
		       title       = coalesce($4, title),
		       description = coalesce($5, description),
		       body        = coalesce($6, body)
		 where id        = $1
		   and author_id = $2
		   and (
		         $3::text is null or not exists (
		           select 1
		             from articles
		            where id   != $1
		              and slug  = $3::text
		         )
		   )
		returning slug
	`

	row := r.service.QueryRow(query, p.ID, p.AuthorID, p.Slug, p.Title, p.Description, p.Body)

	var slug string
	if err := row.Scan(&slug); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf(`scanning published article: %w`, err)
	}

	return r.FindBySlug(slug)
}

func (r *ArticlesRepository) FindBySlug(slug string) (*schema.Article, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, nil
	}

	list, err := r.List(ArticlesFilter{
		Slug:  slug,
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}

	if list.ArticlesCount == 0 {
		return nil, nil
	}

	return &list.Articles[0], nil
}

func (r *ArticlesRepository) List(f ArticlesFilter) (*schema.ArticleList, error) {

	query := `
		select a.slug,
		       a.title,
		       a.description,
		       a.body,
		       array_agg(t.name) filter (where t.name is not null) tag_list,
		       a.created_at,
		       a.updated_at,
		       json_build_object(
		         'username',  u.username,
		         'bio',       u.bio,
		         'image',     u.image_url,
		         'following', fo.following_id is not null
		       ) author
		  from articles a
		  join users u
		    on a.author_id = u.id
		  left join article_tags at
		    on a.id = at.article_id
		  left join tags t
		    on at.tag_id = t.id
		  left join followers fo
		    on a.author_id = fo.following_id and u.id = fo.follower_id
		  left join favorites fa
		    on a.id = fa.article_id
		  left join users fu
		    on fa.user_id = fu.id
		 where ($1::text[] is null or t.name = any($1::text[]))
		   and ($2 = '' or u.username = $2)
		   and ($3 = '' or fu.username = $3)
		   and ($4 = '' or a.slug = $4)
		 group by a.id,
		          a.slug,
		          a.title,
		          a.description,
		          a.body,
		          a.created_at,
		          a.updated_at,
		          u.username,
		          u.bio,
		          u.image_url,
		          fo.following_id
		 order by a.id desc
		 limit case when $5 = 0 then 20 else $5 end
		 offset $6
	`

	rows, err := r.service.Query(
		query,
		pq.Array(f.Tags),
		f.Author,
		f.Favorited,
		f.Slug,
		f.Limit,
		f.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf(`selecting articles: %w`, err)
	}
	defer rows.Close()

	articles := make([]schema.Article, 0)
	for rows.Next() {
		a := schema.Article{}

		err := rows.Scan(
			&a.Slug,
			&a.Title,
			&a.Description,
			&a.Body,
			pq.Array(&a.TagList),
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.Author,
		)
		if err != nil {
			return nil, fmt.Errorf(`scanning article: %w`, err)
		}

		articles = append(articles, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`iterating article results: %w`, err)
	}

	list := schema.ArticleList{Articles: articles, ArticlesCount: len(articles)}

	return &list, nil
}

func (r *ArticlesRepository) DeleteBySlug(authorID uint64, slug string) (int64, error) {
	if strings.TrimSpace(slug) == "" {
		return 0, nil
	}

	sql := `
		delete
		  from articles
		 where slug      = $1
		   and author_id = $2
	`

	result, err := r.service.Exec(sql, slug, authorID)
	if err != nil {
		return 0, fmt.Errorf("deleting article by slug %q: %w", slug, err)
	}

	return result.RowsAffected()
}
