package opds1

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

const testFeedXML = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>urn:uuid:2853dacf-ed79-42f5-8e8a-a7bb3d1ae6a2</id>
  <title>OPDS Catalog Root Example</title>
  <updated>2010-01-10T10:03:10Z</updated>
  <author>
    <name>Spec Writer</name>
    <uri>http://opds-spec.org</uri>
  </author>
  <link rel="self" href="/opds-catalogs/root.xml" type="application/atom+xml;profile=opds-catalog;kind=navigation"/>
  <link rel="start" href="/opds-catalogs/root.xml" type="application/atom+xml;profile=opds-catalog;kind=navigation"/>
  <entry>
    <title>Popular Publications</title>
    <link rel="http://opds-spec.org/sort/popular" href="/opds-catalogs/popular.xml" type="application/atom+xml;profile=opds-catalog;kind=acquisition"/>
    <updated>2010-01-10T10:01:01Z</updated>
    <id>urn:uuid:d49e8018-a0e0-499e-9423-7c175fa0c56e</id>
    <content type="text">Popular publications from this catalog based on downloads.</content>
  </entry>
</feed>`

func TestParse(t *testing.T) {
	t.Parallel()
	feed, err := Parse([]byte(testFeedXML))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if feed.ID != "urn:uuid:2853dacf-ed79-42f5-8e8a-a7bb3d1ae6a2" {
		t.Errorf("expected ID urn:uuid:2853dacf-ed79-42f5-8e8a-a7bb3d1ae6a2, got %s", feed.ID)
	}
	if feed.Title != "OPDS Catalog Root Example" {
		t.Errorf("expected title OPDS Catalog Root Example, got %s", feed.Title)
	}
	if feed.Updated != "2010-01-10T10:03:10Z" {
		t.Errorf("expected updated 2010-01-10T10:03:10Z, got %s", feed.Updated)
	}
	if feed.Author == nil || feed.Author.Name != "Spec Writer" {
		t.Errorf("expected author Spec Writer")
	}
	if len(feed.Links) != 2 {
		t.Errorf("expected 2 links, got %d", len(feed.Links))
	}
	if len(feed.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(feed.Entries))
	}
	if feed.Entries[0].Title != "Popular Publications" {
		t.Errorf("expected entry title Popular Publications, got %s", feed.Entries[0].Title)
	}
}

func TestBuildAndMarshal(t *testing.T) {
	t.Parallel()
	feed := NewFeedBuilder("urn:uuid:test-123", "Test Feed", "2024-01-01T00:00:00Z").
		Author("Test Author", "http://example.com").
		AddLink(NewLinkBuilder().Rel(RelSelf).Href("/test.xml").Type(TypeNavigation).Build()).
		AddEntry(NewEntryBuilder("urn:uuid:entry-1", "Test Entry", "2024-01-01T00:00:00Z").
			Summary("A test entry").
			AddLink(NewLinkBuilder().Rel(RelAcquisition).Href("/book.epub").Build()). // auto-detects MIMEEPUB!
			MustBuild()).
		MustBuild()

	xmlBytes, err := feed.ToXML()
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// Verify it's valid XML
	if !strings.Contains(string(xmlBytes), "<?xml version") {
		t.Error("missing XML header")
	}
	if !strings.Contains(string(xmlBytes), "Test Feed") {
		t.Error("missing feed title")
	}
	if !strings.Contains(string(xmlBytes), "Test Entry") {
		t.Error("missing entry title")
	}
}

func TestValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buildFn func() error
		wantErr error
	}{
		{
			name: "empty feed ID",
			buildFn: func() error {
				_, err := NewFeedBuilder("", "Title", "").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "empty feed title",
			buildFn: func() error {
				_, err := NewFeedBuilder("id", "", "").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "invalid feed timestamp",
			buildFn: func() error {
				_, err := NewFeedBuilder("id", "Title", "invalid-date").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "empty entry ID",
			buildFn: func() error {
				_, err := NewEntryBuilder("", "Title", "").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "empty entry title",
			buildFn: func() error {
				_, err := NewEntryBuilder("id", "", "").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "invalid entry timestamp",
			buildFn: func() error {
				_, err := NewEntryBuilder("id", "Title", "not-a-date").Build()
				return err
			},
			wantErr: ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.buildFn()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	t.Run("valid feed auto timestamp", func(t *testing.T) {
		t.Parallel()
		feed, err := NewFeedBuilder("id", "Title", "").Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if feed.Updated == "" {
			t.Error("expected auto-populated Updated timestamp")
		}
	})

	t.Run("link mime auto detection", func(t *testing.T) {
		t.Parallel()
		link := NewLinkBuilder().Rel(RelAcquisition).Href("/book.epub").Build()
		if link.Type != MIMEEPUB {
			t.Errorf("auto MIME detection failed: got %q, want %q", link.Type, MIMEEPUB)
		}
	})
}

func assertPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected %s to panic", name)
		}
		if err, ok := r.(error); !ok || !errors.Is(err, ErrValidation) {
			t.Errorf("expected panic with ErrValidation, got %v", r)
		}
	}()
	fn()
}

func TestMustBuildPanic(t *testing.T) {
	t.Parallel()
	t.Run("feed", func(t *testing.T) {
		t.Parallel()
		assertPanic(t, "feed", func() {
			NewFeedBuilder("", "Title", "").MustBuild()
		})
	})

	t.Run("entry", func(t *testing.T) {
		t.Parallel()
		assertPanic(t, "entry", func() {
			NewEntryBuilder("", "Title", "").MustBuild()
		})
	})
}

func TestParseAndRebuild(t *testing.T) {
	t.Parallel()
	feed, err := Parse([]byte(testFeedXML))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	xmlBytes, err := feed.ToXML()
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	feed2, err := Parse(xmlBytes)
	if err != nil {
		t.Fatalf("Parse again failed: %v", err)
	}

	if feed2.ID != feed.ID {
		t.Errorf("ID mismatch after roundtrip: %s != %s", feed2.ID, feed.ID)
	}
	if feed2.Title != feed.Title {
		t.Errorf("Title mismatch after roundtrip: %s != %s", feed2.Title, feed.Title)
	}
	if len(feed2.Entries) != len(feed.Entries) {
		t.Errorf("Entry count mismatch after roundtrip: %d != %d", len(feed2.Entries), len(feed.Entries))
	}
}

func TestSafety(t *testing.T) {
	t.Parallel()
	t.Run("nil receivers", func(t *testing.T) {
		t.Parallel()
		var f *Feed
		if err := f.Validate(); !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation for nil feed Validate, got %v", err)
		}
		if _, err := f.ToXML(); err == nil {
			t.Error("expected error for nil feed ToXML")
		}

		var e *Entry
		if err := e.Validate(); !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation for nil entry Validate, got %v", err)
		}
		if _, err := e.ToXML(); err == nil {
			t.Error("expected error for nil entry ToXML")
		}
	})

	t.Run("defensive slice cloning", func(t *testing.T) {
		t.Parallel()
		b := NewFeedBuilder("f1", "Feed 1", "")
		b.AddLink(NewLinkBuilder().Rel("self").Href("/feed").Build())
		b.AddEntryValue(Entry{ID: "e1", Title: "Entry 1"})

		feed1, err := b.Build()
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}

		feed1.Links = append(feed1.Links, Link{Rel: "next", Href: "/feed2"})
		feed1.Entries = append(feed1.Entries, Entry{ID: "e2", Title: "Entry 2"})

		feed2, err := b.Build()
		if err != nil {
			t.Fatalf("Build 2 failed: %v", err)
		}
		if len(feed2.Links) != 1 {
			t.Errorf("expected 1 link, got %d", len(feed2.Links))
		}
		if len(feed2.Entries) != 1 {
			t.Errorf("expected 1 entry, got %d", len(feed2.Entries))
		}
	})

	t.Run("unsafe link schemes rejected", func(t *testing.T) {
		t.Parallel()
		b := NewFeedBuilder("f1", "Feed 1", "").
			AddLink(Link{Rel: "acquisition", Href: "javascript:alert(1)"})
		_, err := b.Build()
		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation for javascript: href, got %v", err)
		}

		eb := NewEntryBuilder("e1", "Entry 1", "").
			AddLink(Link{Rel: "acquisition", Href: "data:text/html;base64,PHNjcmlwdD4="})
		_, err = eb.Build()
		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation for data:text/html href, got %v", err)
		}
	})
}

func TestWriteTo(t *testing.T) {
	t.Parallel()

	t.Run("feed WriteTo", func(t *testing.T) {
		t.Parallel()
		feed := NewFeedBuilder("urn:test:writeto", "WriteTo Feed", "2024-01-01T00:00:00Z").MustBuild()
		var buf bytes.Buffer
		n, err := feed.WriteTo(&buf)
		if err != nil || n == 0 {
			t.Fatalf("feed.WriteTo returned n=%d, err=%v", n, err)
		}
		if !strings.Contains(buf.String(), "WriteTo Feed") {
			t.Errorf("expected feed title in XML output")
		}

		var nilFeed *Feed
		if _, err := nilFeed.WriteTo(&buf); err == nil {
			t.Error("expected error for nil feed WriteTo")
		}
	})

	t.Run("entry WriteTo", func(t *testing.T) {
		t.Parallel()
		entry := NewEntryBuilder("urn:test:entry-writeto", "WriteTo Entry", "2024-01-01T00:00:00Z").MustBuild()
		var buf bytes.Buffer
		n, err := entry.WriteTo(&buf)
		if err != nil || n == 0 {
			t.Fatalf("entry.WriteTo returned n=%d, err=%v", n, err)
		}
		parsed, err := ParseEntry(buf.Bytes())
		if err != nil || parsed.Title != "WriteTo Entry" {
			t.Fatalf("ParseEntry failed or title mismatch: %v", err)
		}

		var nilEntry *Entry
		if _, err := nilEntry.WriteTo(&buf); err == nil {
			t.Error("expected error for nil entry WriteTo")
		}
	})
}

func TestParseReaderAndParseEntry(t *testing.T) {
	t.Parallel()

	t.Run("ParseReader success", func(t *testing.T) {
		t.Parallel()
		feed, err := ParseReader(strings.NewReader(testFeedXML))
		if err != nil {
			t.Fatalf("ParseReader failed: %v", err)
		}
		if feed.ID != "urn:uuid:2853dacf-ed79-42f5-8e8a-a7bb3d1ae6a2" {
			t.Errorf("unexpected feed ID: %s", feed.ID)
		}
	})

	t.Run("ParseReader malformed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseReader(strings.NewReader("<invalid-xml>"))
		if err == nil {
			t.Error("expected error for malformed XML reader")
		}
	})

	t.Run("Parse malformed bytes", func(t *testing.T) {
		t.Parallel()
		_, err := Parse([]byte("<unclosed"))
		if err == nil {
			t.Error("expected error for unclosed XML bytes")
		}
	})

	t.Run("ParseEntry malformed", func(t *testing.T) {
		t.Parallel()
		_, err := ParseEntry([]byte("not xml at all"))
		if err == nil {
			t.Error("expected error for invalid entry XML")
		}
	})
}

func TestFeedBuilderComprehensive(t *testing.T) {
	t.Parallel()
	feed := NewFeedBuilderNow("urn:fb:now", "Now Feed").
		Subtitle("My Subtitle").
		Author("Catalog Admin", "https://admin.example.com").
		AddLink(NewLinkBuilder().Rel(RelSelf).Href("/self").Title("Self Link").Length(512).Build()).
		MustBuild()

	if feed.Subtitle != "My Subtitle" {
		t.Errorf("Subtitle = %q, want %q", feed.Subtitle, "My Subtitle")
	}
	if feed.Author == nil || feed.Author.Name != "Catalog Admin" {
		t.Errorf("Author name = %v, want Catalog Admin", feed.Author)
	}
	if len(feed.Links) != 1 || feed.Links[0].Title != "Self Link" || feed.Links[0].Length != 512 {
		t.Errorf("Link attributes mismatch: %+v", feed.Links)
	}
	if feed.Updated == "" {
		t.Error("expected NewFeedBuilderNow to set non-empty Updated")
	}
}

func TestEntryBuilderComprehensive(t *testing.T) {
	t.Parallel()
	entry := NewEntryBuilderNow("urn:eb:now", "Full Entry").
		Published("2024-01-01T00:00:00Z").
		Author("Book Author", "https://author.example.com").
		Summary("Short summary").
		Content("text/html", "<p>Full content</p>").
		AddCategory(Category{Term: "fiction", Scheme: "http://example.com/scheme", Label: "Fiction"}).
		Rights("All Rights Reserved").
		DCLanguage("en").
		DCIssued("2024-01-01").
		DCIdentifier("urn:isbn:9780000000000").
		DCPublisher("Grand Publisher").
		MustBuild()

	if entry.Published != "2024-01-01T00:00:00Z" {
		t.Errorf("Published = %q", entry.Published)
	}
	if entry.Author == nil || entry.Author.Name != "Book Author" {
		t.Errorf("Author = %v", entry.Author)
	}
	if entry.Summary == nil || entry.Summary.Body != "Short summary" {
		t.Errorf("Summary = %v", entry.Summary)
	}
	if entry.Content == nil || entry.Content.Body != "<p>Full content</p>" {
		t.Errorf("Content = %v", entry.Content)
	}
	if len(entry.Category) != 1 || entry.Category[0].Term != "fiction" {
		t.Errorf("Category = %v", entry.Category)
	}
	if entry.Rights != "All Rights Reserved" {
		t.Errorf("Rights = %q", entry.Rights)
	}
	if entry.DCLanguage != "en" {
		t.Errorf("DCLanguage = %q", entry.DCLanguage)
	}
	if entry.DCIssued != "2024-01-01" {
		t.Errorf("DCIssued = %q", entry.DCIssued)
	}
	if entry.DCIdentifier != "urn:isbn:9780000000000" {
		t.Errorf("DCIdentifier = %q", entry.DCIdentifier)
	}
	if entry.DCPublisher != "Grand Publisher" {
		t.Errorf("DCPublisher = %q", entry.DCPublisher)
	}
}

func FuzzParse(f *testing.F) {
	f.Add([]byte(testFeedXML))
	f.Add([]byte("<feed></feed>"))
	f.Add([]byte("not xml at all"))
	f.Add([]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><id>id</id><title>t</title></feed>`))

	f.Fuzz(func(t *testing.T, data []byte) {
		feed, err := Parse(data)
		if err != nil {
			return
		}
		if feed == nil {
			t.Error("Parse returned nil feed with nil error")
		}
	})
}

func BenchmarkToXML(b *testing.B) {
	feed := NewFeedBuilder("urn:bench:1", "Benchmark Feed", "2024-01-01T00:00:00Z").
		AddEntry(NewEntryBuilder("urn:bench:e1", "Entry 1", "2024-01-01T00:00:00Z").MustBuild()).
		MustBuild()
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = feed.ToXML()
	}
}

func BenchmarkParse(b *testing.B) {
	data := []byte(testFeedXML)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = Parse(data)
	}
}

