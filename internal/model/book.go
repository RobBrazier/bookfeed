package model

import "time"

type Collection struct {
	Name    string
	Slug    string
	Created time.Time
	Books   []Book
}

func NewCollection(name, slug string, books []Book) Collection {
	return Collection{
		Name:    name,
		Slug:    slug,
		Created: time.Now(),
		Books:   books,
	}
}

type Book struct {
	Id          int
	Slug        string
	Link        string
	Compilation bool
	Title       string
	ReleaseDate time.Time
	Headline    string
	Description string
	Genres      []string
	Authors     []string
	Image       Image
	Series      Series
}

type Image struct {
	Width  int
	Height int
	Url    string
}

type Series struct {
	Title    string
	Position float32
}

type UserInterests struct {
	Authors []string
	Series  []string
}
