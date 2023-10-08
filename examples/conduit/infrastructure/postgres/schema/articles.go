package schema

import "time"

type NewArticle struct {
	Title       string
	Description string
	Body        string
	TagList     []string
}

type ArticleUpdate struct {
	Title       string
	Description string
	Body        string
	TagList     []string
}

type Article struct {
	Slug           string
	Title          string
	Description    string
	Body           string
	TagList        []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Favorited      bool
	FavoritesCount uint
	Author         Profile
}

type ArticleList struct {
	Articles      []Article
	ArticlesCount uint
}
