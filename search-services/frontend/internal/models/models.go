package models

type Comic struct {
	ID      int64  `json:"id"`
	URL     string `json:"url"`
	Title   string
	Image   string
	PageURL string
}

type PageData struct {
	Title    string
	Comics   []Comic
	Error    string
	IsAdmin  bool
	Username string
	Stats    *Stats
	Status   string
	Query    string
	Limit    string
}

type Stats struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}
