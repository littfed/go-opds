package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/littfed/go-opds/handler"
	"github.com/littfed/go-opds/opds1"
)

func ExampleNew() {
	feed := opds1.NewFeedBuilder("urn:catalog:root", "Library", "2024-01-01T00:00:00Z").MustBuild()

	h := handler.New(handler.Config{
		DefaultVersion: "1.2",
		FeedFunc: func(ctx context.Context, path string) (any, error) {
			return feed, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("X-Content-Type-Options"))
	// Output:
	// 200
	// nosniff
}
