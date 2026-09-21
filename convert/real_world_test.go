package convert_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/littfed/go-opds/convert"
	"github.com/littfed/go-opds/opds1"
	"github.com/littfed/go-opds/opds2"
)

//go:embed testdata/gutenberg_pride_and_prejudice.xml
var realGutenbergXML []byte

func TestRealWorldConvertGutenbergToOPDS2(t *testing.T) {
	t.Parallel()

	feed, err := opds1.Parse(realGutenbergXML)
	if err != nil {
		t.Fatalf("failed to parse real Gutenberg feed: %v", err)
	}

	catalog := convert.Catalog1To2(feed)
	if catalog == nil {
		t.Fatal("Catalog1To2 returned nil for real Gutenberg feed")
	}

	if catalog.Metadata.Title != feed.Title {
		t.Errorf("catalog.Title = %q, want %q", catalog.Metadata.Title, feed.Title)
	}

	t.Run("Publications", func(t *testing.T) {
		assertPublications(t, catalog)
	})

	t.Run("JSONRoundtrip", func(t *testing.T) {
		assertJSONRoundtrip(t, catalog)
	})

	t.Run("ConvertBackToOPDS1", func(t *testing.T) {
		assertConvertBackToOPDS1(t, catalog, feed)
	})
}

func assertPublications(t *testing.T, catalog *opds2.Catalog) {
	t.Helper()
	if len(catalog.Publications) == 0 {
		t.Fatal("expected at least 1 publication in converted catalog")
	}

	firstPub := catalog.Publications[0]
	if firstPub.Metadata.Title != "Pride and Prejudice" {
		t.Errorf("firstPub.Title = %q, want 'Pride and Prejudice'", firstPub.Metadata.Title)
	}
	if len(firstPub.Metadata.Author) == 0 || !strings.Contains(firstPub.Metadata.Author[0].Name, "Austen") {
		t.Errorf("firstPub.Author = %+v, expected Jane Austen", firstPub.Metadata.Author)
	}
	if firstPub.Metadata.Language != "en" {
		t.Errorf("firstPub.Language = %q, want 'en'", firstPub.Metadata.Language)
	}
}

func assertJSONRoundtrip(t *testing.T, catalog *opds2.Catalog) {
	t.Helper()
	jsonData, err := catalog.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize converted catalog to OPDS 2.0 JSON: %v", err)
	}
	if !strings.Contains(string(jsonData), "Pride and Prejudice") {
		t.Error("JSON missing book title")
	}

	reparsedCat, err := opds2.Parse(jsonData)
	if err != nil {
		t.Fatalf("failed to parse converted OPDS 2.0 JSON: %v", err)
	}
	if reparsedCat.Metadata.Title != catalog.Metadata.Title {
		t.Errorf("reparsed catalog title = %q, want %q", reparsedCat.Metadata.Title, catalog.Metadata.Title)
	}
}

func assertConvertBackToOPDS1(t *testing.T, catalog *opds2.Catalog, originalFeed *opds1.Feed) {
	t.Helper()
	backFeed := convert.Catalog2To1(catalog)
	if backFeed == nil {
		t.Fatal("Catalog2To1 returned nil for real converted catalog")
	}
	if backFeed.Title != originalFeed.Title {
		t.Errorf("backFeed.Title = %q, want %q", backFeed.Title, originalFeed.Title)
	}
	if len(backFeed.Entries) != len(catalog.Publications) {
		t.Errorf("backFeed.Entries len = %d, want %d", len(backFeed.Entries), len(catalog.Publications))
	}
}
