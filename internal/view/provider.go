package view

//go:generate go tool templ generate ./...

type ProviderData struct {
	Title string
	URL   string
}
