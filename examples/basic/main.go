package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/littfed/go-opds/handler"
	"github.com/littfed/go-opds/opds1"
)

func main() {
	// Create a sample OPDS 1.2 catalog
	feed := opds1.NewFeedBuilder(
		"urn:uuid:my-catalog-2024",
		"My Book Collection",
		time.Now().UTC().Format(time.RFC3339),
	).
		Author("Catalog Admin", "http://example.com").
		AddLink(opds1.NewLinkBuilder().
			Rel(opds1.RelSelf).
			Href("/").
			Type(opds1.TypeNavigation).
			Build()).
		AddLink(opds1.NewLinkBuilder().
			Rel(opds1.RelStart).
			Href("/").
			Type(opds1.TypeNavigation).
			Build()).
		AddEntry(opds1.NewEntryBuilder(
			"urn:uuid:category-fiction",
			"Fiction",
			time.Now().UTC().Format(time.RFC3339),
		).
			Summary("Browse fiction books").
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelSubsection).
				Href("/fiction").
				Type(opds1.TypeAcquisition).
				Build()).
			MustBuild()).
		AddEntry(opds1.NewEntryBuilder(
			"urn:uuid:category-nonfiction",
			"Non-Fiction",
			time.Now().UTC().Format(time.RFC3339),
		).
			Summary("Browse non-fiction books").
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelSubsection).
				Href("/nonfiction").
				Type(opds1.TypeAcquisition).
				Build()).
			MustBuild()).
		MustBuild()

	// Create a sample acquisition feed
	fictionFeed := opds1.NewFeedBuilder(
		"urn:uuid:fiction-feed",
		"Fiction",
		time.Now().UTC().Format(time.RFC3339),
	).
		AddLink(opds1.NewLinkBuilder().
			Rel(opds1.RelSelf).
			Href("/fiction").
			Type(opds1.TypeAcquisition).
			Build()).
		AddLink(opds1.NewLinkBuilder().
			Rel(opds1.RelStart).
			Href("/").
			Type(opds1.TypeNavigation).
			Build()).
		AddEntry(opds1.NewEntryBuilder(
			"urn:uuid:book-1",
			"The Great Adventure",
			time.Now().UTC().Format(time.RFC3339),
		).
			Author("John Smith", "http://example.com/authors/john-smith").
			Summary("An epic adventure story that spans centuries.").
			DCLanguage("en").
			DCIssued("2020-01-01").
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelImage).
				Href("/covers/book-1.jpg").
				Type("image/jpeg").
				Build()).
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelImageThumbnail).
				Href("/covers/book-1-thumb.jpg").
				Type("image/jpeg").
				Build()).
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelAcquisition).
				Href("/files/book-1.epub").
				Type(opds1.MIMEEPUB).
				Build()).
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelAcquisition).
				Href("/files/book-1.pdf").
				Type(opds1.MIMEPDF).
				Build()).
			MustBuild()).
		MustBuild()

	// Feed lookup function
	feeds := map[string]any{
		"/": feed,
		"/fiction": fictionFeed,
		"/nonfiction": opds1.NewFeedBuilder(
			"urn:uuid:nonfiction-feed",
			"Non-Fiction",
			time.Now().UTC().Format(time.RFC3339),
		).MustBuild(),
	}

	logger := slog.Default()
	h := handler.New(handler.Config{
		BaseURL:        "http://localhost:8080",
		DefaultVersion: "1.2",
		FeedFunc: func(ctx context.Context, path string) (any, error) {
			path = strings.TrimSuffix(path, "/")
			if path == "" {
				path = "/"
			}
			return feeds[path], nil
		},
		Logger: logger,
	})

	logger.Info("OPDS server starting", slog.String("url", "http://localhost:8080"))
	logger.Info("Add this URL to your OPDS reader: http://localhost:8080")

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server stopped unexpectedly", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
