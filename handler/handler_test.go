package handler_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/littfed/go-opds/handler"
	"github.com/littfed/go-opds/opds1"
	"github.com/littfed/go-opds/opds2"
)

func TestHandlerOPDS1Default(t *testing.T) {
	feed := opds1.NewFeedBuilder("urn:test:root", "Test Library", "2026-01-01T00:00:00Z").MustBuild()

	h := handler.New(handler.Config{
		DefaultVersion: "1.2",
		FeedFunc: func(ctx context.Context, path string) (any, error) {
			if ctx == nil {
				return nil, errors.New("nil context")
			}
			return feed, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/atom+xml") {
		t.Errorf("expected atom+xml content type, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Test Library") {
		t.Errorf("body missing feed title: %s", body)
	}
}

func TestHandlerContentNegotiation1To2(t *testing.T) {
	// Provider supplies OPDS 1.2 Feed
	feed := opds1.NewFeedBuilder("urn:test:root", "Test Library", "2026-01-01T00:00:00Z").
		AddEntry(opds1.NewEntryBuilder("urn:book:1", "My Book", "2026-01-01T00:00:00Z").
			AddLink(opds1.NewLinkBuilder().Rel(opds1.RelAcquisition).Href("/book.epub").Build()).
			MustBuild()).
		MustBuild()

	h := handler.New(handler.Config{
		DefaultVersion: "1.2",
		FeedFunc: func(ctx context.Context, path string) (any, error) {
			return feed, nil
		},
	})

	// Client requests OPDS 2.0 JSON
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/opds-catalog+json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/opds-catalog+json") {
		t.Errorf("expected opds-catalog+json content type, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "My Book") || !strings.Contains(body, "publications") {
		t.Errorf("expected auto-translated OPDS 2.0 JSON with publications: %s", body)
	}
}

func TestHandlerContentNegotiation2To1(t *testing.T) {
	// Provider supplies OPDS 2.0 Catalog
	catalog := opds2.NewCatalogBuilderWithID("urn:test:2", "Modern Library").
		AddPublication(opds2.NewPublicationBuilderWithID("urn:book:json", "JSON Book").
			AddAcquisitionLink("/book.epub", opds2.MIMEEPUB).
			MustBuild()).
		MustBuild()

	h := handler.New(handler.Config{
		DefaultVersion: "2.0",
		FeedFunc: func(ctx context.Context, path string) (any, error) {
			return catalog, nil
		},
	})

	// Client requests OPDS 1.2 XML
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/atom+xml")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/atom+xml") {
		t.Errorf("expected atom+xml content type, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "JSON Book") || !strings.Contains(body, "<feed") {
		t.Errorf("expected auto-translated OPDS 1.2 XML feed: %s", body)
	}
}

func TestHandlerEntryEndpoint(t *testing.T) {
	entry := opds1.NewEntryBuilder("urn:book:1", "Single Book", "2026-01-01T00:00:00Z").
		AddLink(opds1.NewLinkBuilder().Rel(opds1.RelAcquisition).Href("/book.epub").Build()).
		MustBuild()

	h := handler.New(handler.Config{
		EntryFunc: func(ctx context.Context, id string) (any, error) {
			if id == "book-1" {
				return entry, nil
			}
			return nil, handler.ErrNotFound
		},
	})

	t.Run("xml", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/entry/book-1", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "application/atom+xml") {
			t.Errorf("expected atom+xml, got %s", rec.Header().Get("Content-Type"))
		}
	})

	t.Run("json_negotiation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/entry/book-1", nil)
		req.Header.Set("Accept", "application/opds-publication+json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "application/opds-publication+json") {
			t.Errorf("expected publication+json, got %s", rec.Header().Get("Content-Type"))
		}
		if !strings.Contains(rec.Body.String(), "Single Book") {
			t.Errorf("expected translated publication in JSON body: %s", rec.Body.String())
		}
	})

	t.Run("not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/entry/unknown", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for unknown entry, got %d", rec.Code)
		}
	})
}

func TestHandlerErrorHandling(t *testing.T) {
	t.Run("default safe error message", func(t *testing.T) {
		h := handler.New(handler.Config{
			FeedFunc: func(ctx context.Context, path string) (any, error) {
				return nil, errors.New("database connection failed")
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500 Internal Server Error, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "internal server error") {
			t.Errorf("expected generic error message in default mode, got %s", rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "database connection failed") {
			t.Errorf("internal database details leaked in response body: %s", rec.Body.String())
		}
	})

	t.Run("opt-in ExposeErrors mode", func(t *testing.T) {
		h := handler.New(handler.Config{
			FeedFunc: func(ctx context.Context, path string) (any, error) {
				return nil, errors.New("database connection failed")
			},
			ExposeErrors: true,
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500 Internal Server Error, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "database connection failed") {
			t.Errorf("expected detailed error message in debug mode, got %s", rec.Body.String())
		}
	})
}

func TestHandlerFileServing(t *testing.T) {
	content := "dummy epub binary data"
	h := handler.New(handler.Config{
		FileFunc: func(ctx context.Context, id string) (io.ReadCloser, int64, string, error) {
			if id == "test.epub" {
				return io.NopCloser(strings.NewReader(content)), int64(len(content)), "application/epub+zip", nil
			}
			return nil, 0, "", handler.ErrNotFound
		},
	})

	t.Run("serve_file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/file/test.epub", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		if rec.Header().Get("Content-Type") != "application/epub+zip" {
			t.Errorf("unexpected content type: %s", rec.Header().Get("Content-Type"))
		}
		if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("missing X-Content-Type-Options: nosniff header, got: %s", rec.Header().Get("X-Content-Type-Options"))
		}
		if rec.Body.String() != content {
			t.Errorf("body = %q, want %q", rec.Body.String(), content)
		}
	})

	t.Run("file_not_found", func(t *testing.T) {
		reqNF := httptest.NewRequest(http.MethodGet, "/file/unknown.epub", nil)
		recNF := httptest.NewRecorder()
		h.ServeHTTP(recNF, reqNF)
		if recNF.Code != http.StatusNotFound {
			t.Errorf("expected 404 for missing file, got %d", recNF.Code)
		}
	})

	t.Run("path_traversal_rejected", func(t *testing.T) {
		traversals := []string{
			"/file/..%2F..%2Fetc%2Fpasswd",
			"/file/%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			"/file/book/../../../secret.txt",
			"/entry/book/../../root",
			"/file/..\\windows\\win.ini",
		}
		for _, target := range traversals {
			t.Run(target, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, target, nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)
				if rec.Code != http.StatusNotFound && rec.Code != http.StatusTemporaryRedirect && rec.Code != http.StatusMovedPermanently {
					t.Errorf("path traversal %q not blocked, got %d", target, rec.Code)
				}
			})
		}
	})
}

func TestHandlerFileUnexpectedError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	h := handler.New(handler.Config{
		FileFunc: func(ctx context.Context, id string) (io.ReadCloser, int64, string, error) {
			return nil, 0, "", errors.New("disk read error")
		},
		Logger:       logger,
		ExposeErrors: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/file/corrupt.epub", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "disk read error") {
		t.Errorf("expected error message in response body, got %s", rec.Body.String())
	}
	if !strings.Contains(logBuf.String(), "file provider failed") {
		t.Errorf("expected error logged, got: %s", logBuf.String())
	}
}

func TestHandlerUnsupportedType(t *testing.T) {
	h := handler.New(handler.Config{
		FeedFunc: func(ctx context.Context, path string) (any, error) {
			return "a string is not a valid feed", nil
		},
		ExposeErrors: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for unsupported type, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "handler: unsupported type") {
		t.Errorf("expected ErrUnsupportedType in response, got %s", rec.Body.String())
	}
}

type mockProvider struct {
	feed *opds1.Feed
}

func (m *mockProvider) Feed(ctx context.Context, path string) (any, error) {
	return m.feed, nil
}

func TestHandlerCatalogProvider(t *testing.T) {
	feed := opds1.NewFeedBuilder("urn:test:prov", "Provider Library", "2026-01-01T00:00:00Z").MustBuild()
	h := handler.New(handler.Config{
		Provider: &mockProvider{feed: feed},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Provider Library") {
		t.Errorf("expected feed title in body: %s", rec.Body.String())
	}
}

func TestHandlerSafety(t *testing.T) {
	t.Run("typed nil feed pointer", func(t *testing.T) {
		h := handler.New(handler.Config{
			FeedFunc: func(ctx context.Context, path string) (any, error) {
				var nilFeed *opds1.Feed
				return nilFeed, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for typed nil feed, got %d", rec.Code)
		}
	})

	t.Run("typed nil entry pointer", func(t *testing.T) {
		h := handler.New(handler.Config{
			EntryFunc: func(ctx context.Context, id string) (any, error) {
				var nilEntry *opds1.Entry
				return nilEntry, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/entry/book-1", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for typed nil entry, got %d", rec.Code)
		}
	})

	t.Run("nil file reader does not panic", func(t *testing.T) {
		h := handler.New(handler.Config{
			FileFunc: func(ctx context.Context, id string) (io.ReadCloser, int64, string, error) {
				return nil, 0, "application/epub+zip", nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/file/book.epub", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for nil file reader, got %d", rec.Code)
		}
	})
}

func TestHandlerConcurrentTraffic(t *testing.T) {
	t.Parallel()
	feed := opds1.NewFeedBuilder("urn:test:root", "Concurrent Feed", "2026-01-01T00:00:00Z").MustBuild()
	h := handler.New(handler.Config{
		FeedFunc: func(ctx context.Context, path string) (any, error) {
			return feed, nil
		},
		EntryFunc: func(ctx context.Context, id string) (any, error) {
			return &opds1.Entry{ID: id, Title: "Book " + id}, nil
		},
	})

	var wg sync.WaitGroup
	const workers = 50
	for i := range workers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("worker %d: expected 200 OK, got %d", id, rec.Code)
			}
		}(i)
	}
	wg.Wait()
}
