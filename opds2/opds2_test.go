package opds2

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

const testCatalogJSON = `{
  "metadata": {
    "title": "Test Catalog",
    "modified": "2024-01-01T00:00:00Z",
    "id": "urn:uuid:test-catalog"
  },
  "links": [
    {
      "rel": "self",
      "href": "/catalog.json",
      "type": "application/opds-catalog+json"
    }
  ],
  "publications": [
    {
      "metadata": {
        "title": "Test Book",
        "identifier": "urn:uuid:test-book",
        "author": [
          {"name": "Test Author"}
        ],
        "language": "en",
        "description": "A test book"
      },
      "links": [
        {
          "rel": "http://opds-spec.org/acquisition",
          "href": "/books/test.epub",
          "type": "application/epub+zip"
        }
      ]
    }
  ],
  "navigation": [
    {
      "metadata": {
        "title": "Browse All"
      },
      "links": [
        {
          "rel": "self",
          "href": "/all",
          "type": "application/opds-catalog+json"
        }
      ]
    }
  ]
}`

func TestParse(t *testing.T) {
	t.Parallel()
	catalog, err := Parse([]byte(testCatalogJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if catalog.Metadata.Title != "Test Catalog" {
		t.Errorf("expected title Test Catalog, got %s", catalog.Metadata.Title)
	}
	if catalog.Metadata.ID != "urn:uuid:test-catalog" {
		t.Errorf("expected ID urn:uuid:test-catalog, got %s", catalog.Metadata.ID)
	}
	if len(catalog.Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(catalog.Links))
	}
	if len(catalog.Publications) != 1 {
		t.Fatalf("expected 1 publication, got %d", len(catalog.Publications))
	}
	if catalog.Publications[0].Metadata.Title != "Test Book" {
		t.Errorf("expected publication title Test Book, got %s", catalog.Publications[0].Metadata.Title)
	}
	if len(catalog.Navigation) != 1 {
		t.Errorf("expected 1 navigation, got %d", len(catalog.Navigation))
	}
}

func TestBuildAndMarshal(t *testing.T) {
	t.Parallel()
	catalog := NewCatalogBuilder("Test Catalog").
		ID("urn:uuid:test-123").
		AddLink(NewLinkBuilder().Rel(RelSelf).Href("/catalog.json").Type(TypeNavigation).Build()).
		AddPublication(NewPublicationBuilder("Test Book").
			Identifier("urn:uuid:book-1").
			Author("Author Name", "urn:uuid:author-1").
			Language("en").
			AddAcquisitionLink("/book.epub", ""). // tests auto MIME detection!
			MustBuild()).
		AddNavigation(NewNavigationBuilder("Browse All").
			AddLink(NewLinkBuilder().Rel(RelSelf).Href("/all").Type(TypeNavigation).Build()).
			MustBuild()).
		MustBuild()

	jsonBytes, err := catalog.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if !strings.Contains(string(jsonBytes), "Test Catalog") {
		t.Error("missing catalog title")
	}
	if !strings.Contains(string(jsonBytes), "Test Book") {
		t.Error("missing publication title")
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
			name: "catalog empty title",
			buildFn: func() error {
				_, err := NewCatalogBuilderWithID("urn:id", "").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "catalog empty ID",
			buildFn: func() error {
				_, err := NewCatalogBuilder("Title").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "catalog bad modified timestamp",
			buildFn: func() error {
				_, err := NewCatalogBuilderWithID("urn:id", "Title").Modified("bad-timestamp").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "pub empty title",
			buildFn: func() error {
				_, err := NewPublicationBuilderWithID("urn:id", "").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "pub empty identifier",
			buildFn: func() error {
				_, err := NewPublicationBuilder("Title").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "pub invalid published date",
			buildFn: func() error {
				_, err := NewPublicationBuilderWithID("urn:id", "Title").Published("invalid-date").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "pub invalid modified timestamp",
			buildFn: func() error {
				_, err := NewPublicationBuilderWithID("urn:id", "Title").Modified("bad-modified").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "nav empty title",
			buildFn: func() error {
				_, err := NewNavigationBuilder("").Build()
				return err
			},
			wantErr: ErrValidation,
		},
		{
			name: "nav invalid modified timestamp",
			buildFn: func() error {
				_, err := NewNavigationBuilder("Title").Modified("bad-date").Build()
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

	t.Run("pub valid published dates", func(t *testing.T) {
		t.Parallel()
		for _, dateStr := range []string{"2024-01-01T12:00:00Z", "2024-01-01"} {
			pub, err := NewPublicationBuilderWithID("urn:id", "Title").Published(dateStr).Build()
			if err != nil {
				t.Errorf("expected valid for published %q, got: %v", dateStr, err)
			}
			if pub.Metadata.Published != dateStr {
				t.Errorf("pub.Published = %q, want %q", pub.Metadata.Published, dateStr)
			}
		}
	})

	t.Run("catalog auto modified timestamp", func(t *testing.T) {
		t.Parallel()
		built, err := NewCatalogBuilderWithID("urn:id", "Valid Title").Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if built.Metadata.Modified == "" {
			t.Error("expected auto-populated Modified timestamp")
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
			t.Errorf("expected ErrValidation panic, got %v", r)
		}
	}()
	fn()
}

func TestMustBuildPanic(t *testing.T) {
	t.Parallel()
	t.Run("catalog", func(t *testing.T) {
		t.Parallel()
		assertPanic(t, "catalog", func() {
			NewCatalogBuilder("").ID("").MustBuild()
		})
	})

	t.Run("publication", func(t *testing.T) {
		t.Parallel()
		assertPanic(t, "publication", func() {
			NewPublicationBuilder("").Identifier("").MustBuild()
		})
	})

	t.Run("navigation", func(t *testing.T) {
		t.Parallel()
		assertPanic(t, "navigation", func() {
			NewNavigationBuilder("").MustBuild()
		})
	})
}

func TestPublicationSerialization(t *testing.T) {
	t.Parallel()
	pub := NewPublicationBuilderWithID("urn:id:1", "My Title").
		Author("Author", "").
		MustBuild()

	data, err := pub.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !strings.Contains(string(data), "My Title") {
		t.Errorf("missing title in serialized publication: %s", string(data))
	}

	compact, err := pub.ToJSONCompact()
	if err != nil {
		t.Fatalf("ToJSONCompact failed: %v", err)
	}
	if !strings.Contains(string(compact), "urn:id:1") {
		t.Errorf("missing identifier in compact publication: %s", string(compact))
	}
}

func TestParseAndRebuild(t *testing.T) {
	t.Parallel()
	catalog, err := Parse([]byte(testCatalogJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsonBytes, err := catalog.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	catalog2, err := Parse(jsonBytes)
	if err != nil {
		t.Fatalf("Parse again failed: %v", err)
	}

	if catalog2.Metadata.Title != catalog.Metadata.Title {
		t.Errorf("Title mismatch after roundtrip: %s != %s", catalog2.Metadata.Title, catalog.Metadata.Title)
	}
	if len(catalog2.Publications) != len(catalog.Publications) {
		t.Errorf("Publications mismatch after roundtrip: %d != %d", len(catalog2.Publications), len(catalog.Publications))
	}
}

func TestSafety(t *testing.T) {
	t.Parallel()
	t.Run("nil receivers", func(t *testing.T) {
		t.Parallel()
		var c *Catalog
		if err := c.Validate(); !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation for nil catalog Validate, got %v", err)
		}
		if _, err := c.ToJSON(); err == nil {
			t.Error("expected error for nil catalog ToJSON")
		}
		if _, err := c.ToJSONCompact(); err == nil {
			t.Error("expected error for nil catalog ToJSONCompact")
		}

		var p *Publication
		if err := p.Validate(); !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation for nil publication Validate, got %v", err)
		}
		if _, err := p.ToJSON(); err == nil {
			t.Error("expected error for nil publication ToJSON")
		}
		if _, err := p.ToJSONCompact(); err == nil {
			t.Error("expected error for nil publication ToJSONCompact")
		}

		var n *Navigation
		if err := n.Validate(); !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation for nil navigation Validate, got %v", err)
		}
	})

	t.Run("defensive slice cloning", func(t *testing.T) {
		t.Parallel()
		b := NewCatalogBuilderWithID("c1", "Cat 1")
		b.AddLink(NewLinkBuilder().Rel("self").Href("/cat").Build())
		b.AddPublicationValue(Publication{Metadata: PublicationMetadata{Identifier: "p1", Title: "Pub 1"}})

		cat1, err := b.Build()
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}

		cat1.Links = append(cat1.Links, Link{Rel: "next", Href: "/cat2"})
		cat1.Publications = append(cat1.Publications, Publication{Metadata: PublicationMetadata{Identifier: "p2", Title: "Pub 2"}})

		cat2, err := b.Build()
		if err != nil {
			t.Fatalf("Build 2 failed: %v", err)
		}
		if len(cat2.Links) != 1 {
			t.Errorf("expected 1 link, got %d", len(cat2.Links))
		}
		if len(cat2.Publications) != 1 {
			t.Errorf("expected 1 publication, got %d", len(cat2.Publications))
		}
	})

	t.Run("unsafe link schemes rejected", func(t *testing.T) {
		t.Parallel()
		b := NewCatalogBuilderWithID("c1", "Cat 1").
			AddLink(Link{Rel: "self", Href: "javascript:alert(1)"})
		_, err := b.Build()
		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation for javascript: href, got %v", err)
		}

		pb := NewPublicationBuilderWithID("p1", "Pub 1").
			AddLink(Link{Rel: "acquisition", Href: "data:text/html;base64,PHNjcmlwdD4="})
		_, err = pb.Build()
		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation for data:text/html href, got %v", err)
		}
	})
}

func TestWriteToAndCompact(t *testing.T) {
	t.Parallel()

	t.Run("catalog WriteTo and ToJSONCompact", func(t *testing.T) {
		t.Parallel()
		cat := NewCatalogBuilderWithID("urn:test:cat-write", "Cat Write").MustBuild()

		var buf bytes.Buffer
		n, err := cat.WriteTo(&buf)
		if err != nil {
			t.Fatalf("cat.WriteTo failed: %v", err)
		}
		if n <= 0 || int64(buf.Len()) != n {
			t.Errorf("bytes mismatch: wrote %d, buf %d", n, buf.Len())
		}

		compact, err := cat.ToJSONCompact()
		if err != nil {
			t.Fatalf("cat.ToJSONCompact failed: %v", err)
		}
		if !strings.Contains(string(compact), "urn:test:cat-write") {
			t.Errorf("compact missing ID: %s", string(compact))
		}

		var nilCat *Catalog
		if _, err := nilCat.WriteTo(&buf); err == nil {
			t.Error("expected error on nilCat.WriteTo")
		}
	})

	t.Run("publication WriteTo", func(t *testing.T) {
		t.Parallel()
		pub := NewPublicationBuilderWithID("urn:test:pub-write", "Pub Write").MustBuild()

		var buf bytes.Buffer
		n, err := pub.WriteTo(&buf)
		if err != nil {
			t.Fatalf("pub.WriteTo failed: %v", err)
		}
		if n <= 0 || int64(buf.Len()) != n {
			t.Errorf("bytes mismatch: wrote %d, buf %d", n, buf.Len())
		}

		var nilPub *Publication
		if _, err := nilPub.WriteTo(&buf); err == nil {
			t.Error("expected error on nilPub.WriteTo")
		}
	})
}

func TestParseReader(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		cat, err := ParseReader(strings.NewReader(testCatalogJSON))
		if err != nil {
			t.Fatalf("ParseReader failed: %v", err)
		}
		if cat.Metadata.Title != "Test Catalog" {
			t.Errorf("parsed Title = %q, want %q", cat.Metadata.Title, "Test Catalog")
		}
	})

	t.Run("malformed json reader", func(t *testing.T) {
		t.Parallel()
		_, err := ParseReader(strings.NewReader("{invalid-json"))
		if err == nil {
			t.Error("expected error on malformed json reader")
		}
	})

	t.Run("malformed json bytes", func(t *testing.T) {
		t.Parallel()
		_, err := Parse([]byte("{invalid-json"))
		if err == nil {
			t.Error("expected error on malformed json bytes")
		}
	})
}

func TestCatalogBuilderComprehensive(t *testing.T) {
	t.Parallel()
	group := Group{
		Metadata: GroupMetadata{Title: "Featured Books"},
	}
	navVal := Navigation{Metadata: NavigationMetadata{Title: "By Author"}}

	cat := NewCatalogBuilderWithID("urn:cat:comp", "Comprehensive Catalog").
		AddNavigation(&Navigation{Metadata: NavigationMetadata{Title: "By Genre"}}).
		AddNavigationValue(navVal).
		AddGroup(group).
		MustBuild()

	if len(cat.Navigation) != 2 {
		t.Errorf("len(Navigation) = %d, want 2", len(cat.Navigation))
	}
	if len(cat.Groups) != 1 || cat.Groups[0].Metadata.Title != "Featured Books" {
		t.Errorf("Groups mismatch: %+v", cat.Groups)
	}
}

func TestPublicationBuilderComprehensive(t *testing.T) {
	t.Parallel()
	pub := NewPublicationBuilderWithID("urn:pub:comp", "Comp Title").
		Publisher("Tech Books Inc").
		Description("Detailed book description").
		Subject("Computing").
		Rights("All Rights Reserved").
		Modified("2024-01-01T12:00:00Z").
		NumberOfPages(450).
		AddImage(Image{Href: "/covers/c.png", Type: MIMEPNG, Height: 600, Width: 400}).
		MustBuild()

	if pub.Metadata.Publisher != "Tech Books Inc" {
		t.Errorf("Publisher = %q", pub.Metadata.Publisher)
	}
	if pub.Metadata.Description != "Detailed book description" {
		t.Errorf("Description = %q", pub.Metadata.Description)
	}
	if len(pub.Metadata.Subject) != 1 || pub.Metadata.Subject[0] != "Computing" {
		t.Errorf("Subject = %v", pub.Metadata.Subject)
	}
	if pub.Metadata.Rights != "All Rights Reserved" {
		t.Errorf("Rights = %q", pub.Metadata.Rights)
	}
	if pub.Metadata.Modified != "2024-01-01T12:00:00Z" {
		t.Errorf("Modified = %q", pub.Metadata.Modified)
	}
	if pub.Metadata.NumberOfPages != 450 {
		t.Errorf("NumberOfPages = %d", pub.Metadata.NumberOfPages)
	}
	if len(pub.Images) != 1 || pub.Images[0].Href != "/covers/c.png" {
		t.Errorf("Images = %v", pub.Images)
	}
}

func TestNavigationBuilderComprehensive(t *testing.T) {
	t.Parallel()
	nav := NewNavigationBuilder("Genres").
		Identifier("urn:nav:genres").
		Description("List of genres").
		Modified("2024-01-01T12:00:00Z").
		AddLink(NewLinkBuilder().Rel("self").Href("/genres.json").Build()).
		MustBuild()

	nav2 := NewNavigationBuilderWithID("urn:nav:2", "Nav 2").MustBuild()
	if nav2.Metadata.Identifier != "urn:nav:2" {
		t.Errorf("nav2 Identifier = %q", nav2.Metadata.Identifier)
	}

	if nav.Metadata.Identifier != "urn:nav:genres" {
		t.Errorf("Identifier = %q", nav.Metadata.Identifier)
	}
	if nav.Metadata.Description != "List of genres" {
		t.Errorf("Description = %q", nav.Metadata.Description)
	}
	if nav.Metadata.Modified != "2024-01-01T12:00:00Z" {
		t.Errorf("Modified = %q", nav.Metadata.Modified)
	}
	if len(nav.Links) != 1 {
		t.Errorf("Links len = %d", len(nav.Links))
	}
}

func TestLinkBuilderTemplated(t *testing.T) {
	t.Parallel()
	link := NewLinkBuilder().
		Rel("search").
		Href("/search{?query}").
		Title("Search Link").
		Templated(true).
		Build()

	if !link.Templated {
		t.Error("expected Templated to be true")
	}
	if link.Title != "Search Link" {
		t.Errorf("Title = %q", link.Title)
	}
}

func FuzzParse(f *testing.F) {
	f.Add([]byte(testCatalogJSON))
	f.Add([]byte("{}"))
	f.Add([]byte("not valid json"))
	f.Add([]byte(`{"metadata":{"title":"T","id":"urn:id"}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		cat, err := Parse(data)
		if err != nil {
			return
		}
		if cat == nil {
			t.Error("Parse returned nil catalog with nil error")
		}
	})
}

func BenchmarkToJSON(b *testing.B) {
	catalog := NewCatalogBuilderWithID("urn:bench:cat", "Bench Catalog").
		AddPublication(NewPublicationBuilderWithID("urn:bench:p1", "Bench Pub").MustBuild()).
		MustBuild()
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = catalog.ToJSON()
	}
}

func BenchmarkParse(b *testing.B) {
	data := []byte(testCatalogJSON)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = Parse(data)
	}
}

