package repository

type ArticlesFilter struct {
	Tags      []string
	Author    string
	Favorited string
	Limit     int
	Offset    int
}
