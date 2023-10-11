package repository

import (
	"fmt"

	"github.com/lib/pq"
	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres/schema"
)

type ArticlesRepository struct {
	db service
}

func NewArticlesRepository(db service) ArticlesRepository {
	return ArticlesRepository{db: db}
}

func (r ArticlesRepository) Publish(n schema.NewArticle) (*schema.Article, error) {

	sql := `
		with inserted_article as (
		  insert into articles (author_id, slug, title, description, body)
		  values ($1, $2, $3, $4, $5)
		      on conflict (slug) do nothing
		  returning id,
		            slug,
		            title,
		            description,
		            body,
		            created_at,
		            updated_at,
		            author_id
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
		select ia.slug,
		       ia.title,
		       ia.description,
		       ia.body,
		       array_agg(at.name) filter (where at.name is not null) tag_list,
		       ia.created_at,
		       ia.updated_at,
		       json_build_object(
		         'username',  u.username,
		         'bio',       u.bio,
		         'image',     u.image_url,
		         'following', f.following_id is not null
		       ) author
		  from inserted_article ia
		  join users u
		    on ia.author_id = u.id
		  left join inserted_article_tags iat
		    on ia.id = iat.article_id
		  left join applied_tags at
		    on iat.tag_id = at.id
		  left join followers f
		    on ia.author_id = f.following_id and u.id = f.follower_id
		 group by ia.slug,
		          ia.title,
		          ia.description,
		          ia.body,
		          ia.created_at,
		          ia.updated_at,
		          u.username,
		          u.bio,
		          u.image_url,
		          f.following_id
	`

	rows, err := r.db.QueryRows(sql, n.AuthorID, n.Slug, n.Title, n.Description, n.Body, pq.Array(n.TagList))
	if err != nil {
		return nil, fmt.Errorf(`inserting new article: %w`, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	a := schema.Article{}
	if err := rows.Scan(&a.Slug, &a.Title, &a.Description, &a.Body, pq.Array(&a.TagList), &a.CreatedAt, &a.UpdatedAt, &a.Author); err != nil {
		return nil, fmt.Errorf(`scanning published article: %w`, err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`iterating article results: %w`, err)
	}

	return &a, nil
}

func (r ArticlesRepository) PartialUpdate(p schema.ArticlePatch) (*schema.Article, error) {

	sql := `
		with updated_article as (
		  update articles
		     set slug        = coalesce($3, slug),
		         title       = coalesce($4, title),
		         description = coalesce($5, description),
		         body        = coalesce($6, body)
		   where id        = $1
		     and author_id = $2
		     and case when $3::text is null
		              then true
		              else not exists (select 1 from articles where slug = $3)
		         end
		  returning id,
		            slug,
		            title,
		            description,
		            body,
		            created_at,
		            updated_at,
		            author_id
		)
		select ua.slug,
		       ua.title,
		       ua.description,
		       ua.body,
		       array_agg(t.name) filter (where t.name is not null) tag_list,
		       ua.created_at,
		       ua.updated_at,
		       json_build_object(
		         'username',  u.username,
		         'bio',       u.bio,
		         'image',     u.image_url,
		         'following', f.following_id is not null
		       ) author
		  from updated_article ua
		  join users u
		    on ua.author_id = u.id
		  left join article_tags at
		    on ua.id = at.article_id
		  left join tags t
		    on at.tag_id = t.id
		  left join followers f
		    on ua.author_id = f.following_id and u.id = f.follower_id
		 group by ua.slug,
		          ua.title,
		          ua.description,
		          ua.body,
		          ua.created_at,
		          ua.updated_at,
		          u.username,
		          u.bio,
		          u.image_url,
		          f.following_id
	`

	rows, err := r.db.QueryRows(sql, p.ID, p.AuthorID, p.Slug, p.Title, p.Description, p.Body)
	if err != nil {
		return nil, fmt.Errorf(`updating article: %w`, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	a := schema.Article{}
	if err := rows.Scan(&a.Slug, &a.Title, &a.Description, &a.Body, pq.Array(&a.TagList), &a.CreatedAt, &a.UpdatedAt, &a.Author); err != nil {
		return nil, fmt.Errorf(`scanning updated article: %w`, err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`iterating article results: %w`, err)
	}

	return &a, nil
}
