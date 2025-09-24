package hardcover

type BookGenre struct {
	Tag string `json:"tag"`
}

type BookImage struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type BookSeries struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
	Slug string `json:"slug"`
}

type BookFeaturedSeries struct {
	Series   BookSeries
	Position float32
}

type BookAuthor struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
	Slug string `json:"slug"`
}

type BookContributor struct {
	Author       BookAuthor `json:"author"`
	Contribution string     `json:"contribution"`
}
