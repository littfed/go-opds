package handler

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/littfed/go-opds/convert"
	"github.com/littfed/go-opds/opds1"
	"github.com/littfed/go-opds/opds2"
)

// OPDS version constants.
const (
	Version12 = "1.2"
	Version20 = "2.0"
)

// ErrNotFound is returned when a requested catalog feed or entry cannot be found.
var ErrNotFound = errors.New("handler: not found")

// ErrUnsupportedType is returned when a feed or entry value cannot be converted to the requested OPDS representation.
var ErrUnsupportedType = errors.New("handler: unsupported type")

// FeedFunc is a function that returns an OPDS feed or catalog for the given path.
type FeedFunc func(ctx context.Context, path string) (any, error)

// EntryFunc is a function that returns an OPDS entry or publication for the given ID.
type EntryFunc func(ctx context.Context, id string) (any, error)

// FileFunc is a function that returns a file for the given ID.
type FileFunc func(ctx context.Context, id string) (io.ReadCloser, int64, string, error)

// CatalogProvider is an adapter interface for supplying catalog feeds dynamically.
type CatalogProvider interface {
	Feed(ctx context.Context, path string) (any, error)
}

// Config configures the OPDS HTTP handler.
type Config struct {
	// BaseURL is the base URL for the catalog (e.g., "http://localhost:8080").
	BaseURL string

	// DefaultVersion is the fallback OPDS version when no Accept header is present: "1.2" or "2.0". Default: "1.2".
	DefaultVersion string

	// FeedFunc returns a feed for the given path.
	FeedFunc FeedFunc

	// EntryFunc returns an entry for the given ID.
	EntryFunc EntryFunc

	// FileFunc returns a file for the given ID.
	FileFunc FileFunc

	// Provider is an optional dynamic catalog provider seam taking precedence over FeedFunc.
	Provider CatalogProvider

	// Logger is an optional structured logger. If nil, logs are discarded.
	Logger *slog.Logger

	// ExposeErrors enables returning internal backend error messages in HTTP 500 responses.
	// When false (default), a generic error message is returned to clients, and full details are logged server-side.
	ExposeErrors bool
}

// Handler serves OPDS catalogs over HTTP with automatic content negotiation and translation.
type Handler struct {
	config Config
	mux    *http.ServeMux
	logger *slog.Logger
}

// New creates a new OPDS HTTP handler.
func New(cfg Config) *Handler {
	cfg.DefaultVersion = cmp.Or(cfg.DefaultVersion, Version12)
	logger := cmp.Or(cfg.Logger, slog.New(slog.NewTextHandler(io.Discard, nil)))

	h := &Handler{
		config: cfg,
		mux:    http.NewServeMux(),
		logger: logger,
	}

	h.mux.HandleFunc("/", h.handleRoot)
	h.mux.HandleFunc("/entry/", h.handleEntry)
	h.mux.HandleFunc("/file/", h.handleFile)

	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	h.mux.ServeHTTP(w, r)
}

func isSafeID(id string) bool {
	if id == "" || len(id) > 512 {
		return false
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, "\x00\r\n\\") {
		return false
	}
	clean := path.Clean(id)
	return clean != ".." && !strings.HasPrefix(clean, "../") && !path.IsAbs(clean)
}

func (h *Handler) writeInternalError(w http.ResponseWriter, err error, fallback string) {
	if h.config.ExposeErrors {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(w, fallback, http.StatusInternalServerError)
}

func (h *Handler) negotiateVersion(r *http.Request) string {
	accept := r.Header.Get("Accept")
	switch {
	case strings.Contains(accept, "application/opds-catalog+json"),
		strings.Contains(accept, "application/opds-publication+json"):
		return Version20
	case strings.Contains(accept, "application/atom+xml"):
		return Version12
	case strings.Contains(accept, "application/json"):
		return Version20
	case strings.Contains(accept, "application/xml"), strings.Contains(accept, "text/xml"):
		return Version12
	default:
		return h.config.DefaultVersion
	}
}

func (h *Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var feed any
	var err error

	switch {
	case h.config.Provider != nil:
		feed, err = h.config.Provider.Feed(ctx, r.URL.Path)
	case h.config.FeedFunc != nil:
		feed, err = h.config.FeedFunc(ctx, r.URL.Path)
	default:
		h.logger.Error("no feed provider configured", slog.String("path", r.URL.Path))
		http.Error(w, "no feed provider configured", http.StatusInternalServerError)
		return
	}

	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("feed provider failed", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
		h.writeInternalError(w, err, "internal server error")
		return
	}

	if feed == nil {
		http.NotFound(w, r)
		return
	}

	h.serveCatalog(w, r, feed)
}

//nolint:dupl // symmetric version negotiation and serialization for catalogs and entries
func (h *Handler) serveCatalog(w http.ResponseWriter, r *http.Request, feed any) {
	var (
		data        []byte
		contentType string
		err         error
	)
	if h.negotiateVersion(r) == Version20 {
		var cat *opds2.Catalog
		if cat, err = toCatalog2(feed); err == nil {
			contentType = "application/opds-catalog+json; charset=utf-8"
			data, err = cat.ToJSON()
		}
	} else {
		var f *opds1.Feed
		if f, err = toFeed1(feed); err == nil {
			contentType = "application/atom+xml;profile=opds-catalog;kind=navigation; charset=utf-8"
			data, err = f.ToXML()
		}
	}
	h.serveContent(w, r, contentType, data, err, "catalog")
}

func (h *Handler) handleEntry(w http.ResponseWriter, r *http.Request) {
	if h.config.EntryFunc == nil {
		h.logger.Error("no entry function configured", slog.String("path", r.URL.Path))
		http.Error(w, "no entry function configured", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	id, ok := strings.CutPrefix(r.URL.Path, "/entry/")
	if !ok || !isSafeID(id) {
		http.NotFound(w, r)
		return
	}
	item, err := h.config.EntryFunc(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("entry provider failed", slog.String("path", r.URL.Path), slog.String("id", id), slog.String("error", err.Error()))
		h.writeInternalError(w, err, "internal server error")
		return
	}

	if item == nil {
		http.NotFound(w, r)
		return
	}

	h.serveEntry(w, r, item)
}

//nolint:dupl // symmetric version negotiation and serialization for catalogs and entries
func (h *Handler) serveEntry(w http.ResponseWriter, r *http.Request, item any) {
	var (
		data        []byte
		contentType string
		err         error
	)
	if h.negotiateVersion(r) == Version20 {
		var pub *opds2.Publication
		if pub, err = toPublication2(item); err == nil {
			contentType = "application/opds-publication+json; charset=utf-8"
			data, err = pub.ToJSON()
		}
	} else {
		var entry *opds1.Entry
		if entry, err = toEntry1(item); err == nil {
			contentType = "application/atom+xml;type=entry;profile=opds-catalog; charset=utf-8"
			data, err = entry.ToXML()
		}
	}
	h.serveContent(w, r, contentType, data, err, "entry")
}

func (h *Handler) serveContent(w http.ResponseWriter, r *http.Request, contentType string, data []byte, err error, entity string) {
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("failed preparing "+entity+" response", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
		h.writeInternalError(w, err, "internal server error")
		return
	}
	w.Header().Set("Content-Type", contentType)
	if _, writeErr := w.Write(data); writeErr != nil {
		h.logger.Error("failed writing "+entity+" response", slog.String("path", r.URL.Path), slog.String("error", writeErr.Error()))
	}
}

func (h *Handler) handleFile(w http.ResponseWriter, r *http.Request) {
	if h.config.FileFunc == nil {
		h.logger.Error("no file function configured", slog.String("path", r.URL.Path))
		http.Error(w, "no file function configured", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	id, ok := strings.CutPrefix(r.URL.Path, "/file/")
	if !ok || !isSafeID(id) {
		http.NotFound(w, r)
		return
	}
	rc, size, contentType, err := h.config.FileFunc(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("file provider failed", slog.String("path", r.URL.Path), slog.String("id", id), slog.String("error", err.Error()))
		h.writeInternalError(w, err, "internal server error")
		return
	}
	if rc == nil {
		http.NotFound(w, r)
		return
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			h.logger.Warn("failed closing file reader", slog.String("path", r.URL.Path), slog.String("id", id), slog.String("error", closeErr.Error()))
		}
	}()

	w.Header().Set("Content-Type", contentType)
	var src io.Reader = rc
	if size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		src = io.LimitReader(rc, size)
	}
	if _, err := io.Copy(w, src); err != nil {
		h.logger.Error("streaming file failed", slog.String("path", r.URL.Path), slog.String("id", id), slog.String("error", err.Error()))
	}
}

func toCatalog2(val any) (*opds2.Catalog, error) {
	switch v := val.(type) {
	case *opds2.Catalog:
		if v == nil {
			return nil, fmt.Errorf("%w: catalog is nil", ErrNotFound)
		}
		return v, nil
	case opds2.Catalog:
		return &v, nil
	case *opds1.Feed:
		if v == nil {
			return nil, fmt.Errorf("%w: feed is nil", ErrNotFound)
		}
		return convert.Catalog1To2(v), nil
	case opds1.Feed:
		return convert.Catalog1To2(&v), nil
	default:
		return nil, fmt.Errorf("%w for OPDS 2.0 Catalog: %T", ErrUnsupportedType, val)
	}
}

func toFeed1(val any) (*opds1.Feed, error) {
	switch v := val.(type) {
	case *opds1.Feed:
		if v == nil {
			return nil, fmt.Errorf("%w: feed is nil", ErrNotFound)
		}
		return v, nil
	case opds1.Feed:
		return &v, nil
	case *opds2.Catalog:
		if v == nil {
			return nil, fmt.Errorf("%w: catalog is nil", ErrNotFound)
		}
		return convert.Catalog2To1(v), nil
	case opds2.Catalog:
		return convert.Catalog2To1(&v), nil
	default:
		return nil, fmt.Errorf("%w for OPDS 1.2 Feed: %T", ErrUnsupportedType, val)
	}
}

//nolint:dupl // inverse symmetric type conversion between OPDS 1.2 Entry and OPDS 2.0 Publication
func toPublication2(val any) (*opds2.Publication, error) {
	switch v := val.(type) {
	case *opds2.Publication:
		if v == nil {
			return nil, fmt.Errorf("%w: publication is nil", ErrNotFound)
		}
		return v, nil
	case opds2.Publication:
		return &v, nil
	case *opds1.Entry:
		if v == nil {
			return nil, fmt.Errorf("%w: entry is nil", ErrNotFound)
		}
		p := convert.EntryToPublication(*v)
		return &p, nil
	case opds1.Entry:
		p := convert.EntryToPublication(v)
		return &p, nil
	default:
		return nil, fmt.Errorf("%w for OPDS 2.0 Publication: %T", ErrUnsupportedType, val)
	}
}

//nolint:dupl // inverse symmetric type conversion between OPDS 1.2 Entry and OPDS 2.0 Publication
func toEntry1(val any) (*opds1.Entry, error) {
	switch v := val.(type) {
	case *opds1.Entry:
		if v == nil {
			return nil, fmt.Errorf("%w: entry is nil", ErrNotFound)
		}
		return v, nil
	case opds1.Entry:
		return &v, nil
	case *opds2.Publication:
		if v == nil {
			return nil, fmt.Errorf("%w: publication is nil", ErrNotFound)
		}
		e := convert.PublicationToEntry(*v)
		return &e, nil
	case opds2.Publication:
		e := convert.PublicationToEntry(v)
		return &e, nil
	default:
		return nil, fmt.Errorf("%w for OPDS 1.2 Entry: %T", ErrUnsupportedType, val)
	}
}
