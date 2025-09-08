package hardcover

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/go-retryablehttp"
)

type authTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	version := getVersion()
	req.Header.Set("Authorization", t.key)
	req.Header.Set("User-Agent", fmt.Sprintf("bookfeed/%s (https://github.com/RobBrazier/bookfeed)", version))
	return t.wrapped.RoundTrip(req)
}

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				if len(setting.Value) >= 7 {
					return setting.Value[:7]
				}
				return setting.Value
			}
		}
	}
	return "unknown"
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
