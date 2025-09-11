package hardcover

type BookGenre struct {
	Tag          string `json:"tag"`
	TagSlug      string `json:"tagSlug"`
	Category     string `json:"category"`
	CategorySlug string `json:"categorySlug"`
}

type BookImage struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type BookSeries struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type BookFeaturedSeries struct {
	Series   BookSeries
	Position float32
}
