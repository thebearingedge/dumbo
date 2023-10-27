package schema

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type NewArticle struct {
	AuthorID    uint64
	Slug        string
	Title       string
	Description string
	Body        string
	TagList     []string
}

type ArticlePatch struct {
	ID          uint64
	AuthorID    uint64
	Title       sql.NullString
	Slug        sql.NullString
	Description sql.NullString
	Body        sql.NullString
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
	Author         Author
}

type ArticleList struct {
	Articles      []Article
	ArticlesCount int
}

type Author struct {
	Username  string         `json:"username"`
	Bio       sql.NullString `json:"bio"`
	Image     sql.NullString `json:"image"`
	Following bool           `json:"following"`
}

func (a *Author) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("converting author value to []byte")
	}
	return json.Unmarshal(b, &a)
}
