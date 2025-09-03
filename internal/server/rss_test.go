package server

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/feeds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBuilder is a mock implementation of the Builder interface
type MockBuilder struct {
	mock.Mock
}

func (m *MockBuilder) GetRecentReleases(ctx context.Context) (feeds.Feed, error) {
	args := m.Called(ctx)
	return args.Get(0).(feeds.Feed), args.Error(1)
}

func (m *MockBuilder) GetAuthorReleases(ctx context.Context, author string) (feeds.Feed, error) {
	args := m.Called(ctx, author)
	return args.Get(0).(feeds.Feed), args.Error(1)
}

func (m *MockBuilder) GetSeriesReleases(ctx context.Context, series string) (feeds.Feed, error) {
	args := m.Called(ctx, series)
	return args.Get(0).(feeds.Feed), args.Error(1)
}

func TestRecentHandler(t *testing.T) {
	mockBuilder := new(MockBuilder)
	s := &Server{
		builder: mockBuilder,
	}

	// Create a mock feed
	mockFeed := createMockFeed("Test Feed", "Test Item")

	mockBuilder.On("GetRecentReleases", mock.Anything).Return(mockFeed, nil)

	req := httptest.NewRequest("GET", "/recent", nil)
	w := httptest.NewRecorder()
	s.RecentHandler(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/rss+xml; charset=utf-8")

	// Test with format parameter
	req = httptest.NewRequest("GET", "/recent/atom", nil)
	req.SetPathValue("format", "atom")
	w = httptest.NewRecorder()
	s.RecentHandler(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/atom+xml; charset=utf-8")
}

func TestAuthorHandler(t *testing.T) {
	mockBuilder := new(MockBuilder)
	s := &Server{
		builder: mockBuilder,
	}

	// Create a mock feed
	mockFeed := createMockFeed("Author Feed", "Author Item")

	mockBuilder.On("GetAuthorReleases", mock.Anything, "test-author").Return(mockFeed, nil)

	req := httptest.NewRequest("GET", "/author/test-author", nil)
	req.SetPathValue("author", "test-author")
	w := httptest.NewRecorder()
	s.AuthorHandler(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/rss+xml; charset=utf-8")

	// Test with format parameter
	req = httptest.NewRequest("GET", "/author/test-author/atom", nil)
	req.SetPathValue("author", "test-author")
	req.SetPathValue("format", "atom")
	w = httptest.NewRecorder()
	s.AuthorHandler(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/atom+xml; charset=utf-8")
}

func TestSeriesHandler(t *testing.T) {
	mockBuilder := new(MockBuilder)
	s := &Server{
		builder: mockBuilder,
	}

	// Create a mock feed
	mockFeed := createMockFeed("Series Feed", "Series Item")

	mockBuilder.On("GetSeriesReleases", mock.Anything, "test-series").Return(mockFeed, nil)

	req := httptest.NewRequest("GET", "/series/test-series", nil)
	req.SetPathValue("series", "test-series")
	w := httptest.NewRecorder()
	s.SeriesHandler(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/rss+xml; charset=utf-8")

	// Test with format parameter
	req = httptest.NewRequest("GET", "/series/test-series/json", nil)
	req.SetPathValue("series", "test-series")
	req.SetPathValue("format", "json")
	w = httptest.NewRecorder()
	s.SeriesHandler(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json; charset=utf-8")
}

func createMockFeed(title, itemTitle string) feeds.Feed {
	return feeds.Feed{
		Title:   title,
		Created: time.Now(),
		Items: []*feeds.Item{
			{
				Title: itemTitle,
			},
		},
	}
}
