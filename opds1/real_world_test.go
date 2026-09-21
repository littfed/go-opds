package opds1_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/littfed/go-opds/opds1"
)

//go:embed testdata/gutenberg_root.xml
var gutenbergRootXML []byte

//go:embed testdata/gutenberg_pride_and_prejudice.xml
var gutenbergPrideXML []byte

func TestRealWorldGutenbergRoot(t *testing.T) {
	t.Parallel()

	feed, err := opds1.Parse(gutenbergRootXML)
	if err != nil {
		t.Fatalf("failed to parse real Gutenberg root OPDS: %v", err)
	}

	if feed.Title != "Project Gutenberg" {
		t.Errorf("Title = %q, want 'Project Gutenberg'", feed.Title)
	}
	if feed.Subtitle != "Free eBooks since 1971." {
		t.Errorf("Subtitle = %q", feed.Subtitle)
	}
	if feed.Author == nil || feed.Author.Name != "Project Gutenberg" {
		t.Errorf("Author = %+v, want 'Project Gutenberg'", feed.Author)
	}
	if len(feed.Entries) != 3 {
		t.Fatalf("Entries count = %d, want 3", len(feed.Entries))
	}

	expectedTitles := []string{"Popular", "Latest", "Random"}
	for i, expected := range expectedTitles {
		if feed.Entries[i].Title != expected {
			t.Errorf("entry[%d].Title = %q, want %q", i, feed.Entries[i].Title, expected)
		}
		if len(feed.Entries[i].Links) == 0 {
			t.Errorf("entry[%d] has no links", i)
		}
	}

	// Verify serialization roundtrip on real feed
	outXML, err := feed.ToXML()
	if err != nil {
		t.Fatalf("failed to serialize real feed to XML: %v", err)
	}
	if !strings.Contains(string(outXML), "Project Gutenberg") {
		t.Error("serialized XML missing feed title")
	}

	reparsed, err := opds1.Parse(outXML)
	if err != nil {
		t.Fatalf("failed to re-parse serialized real feed: %v", err)
	}
	if reparsed.Title != feed.Title {
		t.Errorf("reparsed Title = %q, want %q", reparsed.Title, feed.Title)
	}
}

func TestRealWorldGutenbergPublication(t *testing.T) {
	t.Parallel()

	feed, err := opds1.Parse(gutenbergPrideXML)
	if err != nil {
		t.Fatalf("failed to parse real Gutenberg publication OPDS: %v", err)
	}

	if !strings.Contains(feed.Title, "Pride and Prejudice") {
		t.Errorf("feed.Title = %q, expected to contain 'Pride and Prejudice'", feed.Title)
	}
	if len(feed.Entries) == 0 {
		t.Fatal("feed has no entries")
	}

	firstEntry := feed.Entries[0]
	if firstEntry.Title != "Pride and Prejudice" {
		t.Errorf("firstEntry.Title = %q, want 'Pride and Prejudice'", firstEntry.Title)
	}
	if firstEntry.Author == nil || !strings.Contains(firstEntry.Author.Name, "Austen") {
		t.Errorf("firstEntry.Author = %+v, expected Jane Austen", firstEntry.Author)
	}
	if firstEntry.Language() != "en" {
		t.Errorf("Language() = %q, want 'en'", firstEntry.Language())
	}
	if firstEntry.DCTermsLanguage != "en" {
		t.Errorf("DCTermsLanguage = %q, want 'en'", firstEntry.DCTermsLanguage)
	}

	// Verify acquisition links present in real feed
	var hasEPUB, hasKindle bool
	for _, link := range firstEntry.Links {
		if link.Rel == opds1.RelAcquisition {
			if link.Type == opds1.MIMEEPUB {
				hasEPUB = true
			}
			if link.Type == opds1.MIMEMOBI {
				hasKindle = true
			}
		}
	}

	if !hasEPUB {
		t.Error("real Gutenberg entry missing EPUB acquisition link")
	}
	if !hasKindle {
		t.Error("real Gutenberg entry missing Kindle/MOBI acquisition link")
	}
}
