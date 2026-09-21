# Changelog

All notable changes to `go-opds` will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-09-22

### Added
- **OPDS 1.2 Support (`opds1`)**:
  - Full Atom/XML feed data models: `Feed`, `Entry`, `Link`, `Author`, `Category`, `Text`.
  - Fluent invariant-enforcing builders: `FeedBuilder`, `EntryBuilder`, `LinkBuilder`.
  - XML serializers with `ToXML()` and `io.WriterTo` (`WriteTo(io.Writer)`).
  - XML parsers with stream protection: `Parse([]byte)` and `ParseReader(io.Reader)`.
- **OPDS 2.0 Support (`opds2`)**:
  - Full JSON-based catalog data models: `Catalog`, `Publication`, `Navigation`, `Group`, `Link`, `Author`.
  - Fluent builders: `CatalogBuilder`, `PublicationBuilder`, `NavigationBuilder`, `LinkBuilder`.
  - JSON serializers: `ToJSON()`, `ToJSONCompact()`, and `io.WriterTo` (`WriteTo(io.Writer)`).
  - JSON parsers with stream protection: `Parse([]byte)` and `ParseReader(io.Reader)`.
- **Symmetric Translator (`convert`)**:
  - Bidirectional loss-free mapping between OPDS 1.2 Feeds/Entries and OPDS 2.0 Catalogs/Publications.
  - `Catalog1To2` and `Catalog2To1`.
  - `EntryToPublication`, `PublicationToEntry`, `EntryToNavigation`, `NavigationToEntry`.
- **Content-Negotiating HTTP Handler (`handler`)**:
  - Standard `http.Handler` implementation with transparent version negotiation via `Accept` headers.
  - Endpoints for root and sub-catalogs (`/`), entries (`/entry/{id}`), and book files (`/file/{id}`).
  - Built-in defense-in-depth: path traversal prevention (`isSafeID`), MIME confusion mitigation (`X-Content-Type-Options: nosniff`), and input stream bounds (`DefaultMaxReadLimit`).
  - Customizable logging via Go 1.21+ `log/slog`.
- **Examples**:
  - `examples/basic`: Ready-to-run web server with sample library and content negotiation.
  - `examples/client`: CLI browser and crawler for exploring remote OPDS catalogs.
- **CI & Quality**:
  - Multi-version Go matrix (1.22, 1.23, 1.24) and multi-OS (Linux, Windows) CI workflow.
  - Over 40 strict linters configured in `.golangci.yml`.
  - Continuous security vulnerability scanning with `govulncheck`.
  - Fuzz tests and performance benchmarks.
