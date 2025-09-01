package hardcover

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/go-retryablehttp"
)

type authTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", t.key)
	req.Header.Set("User-Agent", fmt.Sprintf("hardcover-rss/%s (https://github.com/RobBrazier/hardcover-rss)", "0.0.0"))
	return t.wrapped.RoundTrip(req)
}

func GetClient(token string) graphql.Client {
	url := "https://api.hardcover.app/v1/graphql"
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{
		Transport: &authTransport{
			key:     token,
			wrapped: http.DefaultTransport,
		},
	}
	retryClient.Logger = slog.Default()
	httpClient := retryClient.StandardClient()
	return graphql.NewClient(url, httpClient)
}
