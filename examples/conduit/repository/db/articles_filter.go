package repository

type ArticlesFilter struct {
	Tags      []string
	Author    string
	Favorited string
	Slug      string
	Limit     int
	Offset    int
}
