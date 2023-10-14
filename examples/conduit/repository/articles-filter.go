package repository

type ArticleFilter struct {
	Tags      []string
	Author    string
	Favorited string
	Limit     int
	Offset    int
}
