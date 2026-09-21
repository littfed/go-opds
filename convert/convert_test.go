package convert_test

import (
	"testing"

	"github.com/littfed/go-opds/convert"
	"github.com/littfed/go-opds/opds1"
)

func TestNilSafety(t *testing.T) {
	t.Parallel()
	if got := convert.Catalog1To2(nil); got != nil {
		t.Errorf("Catalog1To2(nil) = %v; want nil", got)
	}
	if got := convert.Catalog2To1(nil); got != nil {
		t.Errorf("Catalog2To1(nil) = %v; want nil", got)
	}
}

func TestIsAcquisitionRel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		rel  string
		want bool
	}{
		{opds1.RelAcquisition, true},
		{opds1.RelAcquisitionBuy, true},
		{opds1.RelAcquisitionBorrow, true},
		{opds1.RelAcquisitionOpenAccess, true},
		{opds1.RelAcquisitionSample, true},
		{opds1.RelAcquisitionSubscribe, true},
		{opds1.RelSelf, false},
		{opds1.RelSubsection, false},
		{"custom-rel", false},
	}

	for _, tt := range tests {
		t.Run(tt.rel, func(t *testing.T) {
			t.Parallel()
			if got := convert.IsAcquisitionRel(tt.rel); got != tt.want {
				t.Errorf("IsAcquisitionRel(%q) = %v; want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestEntryPublicationSymmetric(t *testing.T) {
	t.Parallel()
	entry := opds1.Entry{
		Title:        "The Great Book",
		ID:           "urn:isbn:978-3-16-148410-0",
		DCIdentifier: "urn:isbn:978-3-16-148410-0",
		Updated:      "2024-01-01T12:00:00Z",
		DCLanguage:   "en",
		DCPublisher:  "Classic Publishing",
		Rights:       "Public Domain",
		Published:    "2024-01-01T00:00:00Z",
		Author: &opds1.Person{
			Name: "John Doe",
			URI:  "https://author.example.com",
		},
		Summary: &opds1.Text{
			Type: "text",
			Body: "A thrilling adventure story.",
		},
		Category: []opds1.Category{
			{Term: "Fiction"},
			{Term: "Adventure"},
		},
		Links: []opds1.Link{
			{
				Rel:  opds1.RelAcquisition,
				Href: "/books/great.epub",
				Type: opds1.MIMEEPUB,
			},
		},
	}

	pub := convert.EntryToPublication(entry)

	if pub.Metadata.Title != entry.Title {
		t.Errorf("pub.Title = %q; want %q", pub.Metadata.Title, entry.Title)
	}
	if pub.Metadata.Identifier != entry.ID {
		t.Errorf("pub.Identifier = %q; want %q", pub.Metadata.Identifier, entry.ID)
	}
	if pub.Metadata.Description != entry.Summary.Body {
		t.Errorf("pub.Description = %q; want %q", pub.Metadata.Description, entry.Summary.Body)
	}
	if pub.Metadata.Modified != entry.Updated {
		t.Errorf("pub.Modified = %q; want %q", pub.Metadata.Modified, entry.Updated)
	}
	if len(pub.Metadata.Author) != 1 || pub.Metadata.Author[0].Name != entry.Author.Name {
		t.Errorf("pub.Author = %+v; want name %q", pub.Metadata.Author, entry.Author.Name)
	}
	if len(pub.Metadata.Subject) != 2 {
		t.Errorf("pub.Subject len = %d; want 2", len(pub.Metadata.Subject))
	}
	if len(pub.Links) != 1 || pub.Links[0].Href != "/books/great.epub" {
		t.Errorf("pub.Links = %+v; want 1 link with href /books/great.epub", pub.Links)
	}

	backEntry := convert.PublicationToEntry(pub)
	if backEntry.Title != entry.Title {
		t.Errorf("backEntry.Title = %q; want %q", backEntry.Title, entry.Title)
	}
	if backEntry.ID != entry.ID {
		t.Errorf("backEntry.ID = %q; want %q", backEntry.ID, entry.ID)
	}
	if backEntry.Updated != entry.Updated {
		t.Errorf("backEntry.Updated = %q; want %q", backEntry.Updated, entry.Updated)
	}
	if backEntry.Summary == nil || backEntry.Summary.Body != entry.Summary.Body {
		t.Errorf("backEntry.Summary = %+v; want body %q", backEntry.Summary, entry.Summary.Body)
	}
}

func TestEntryNavigationSymmetric(t *testing.T) {
	t.Parallel()
	navEntry := opds1.Entry{
		Title:   "Fiction Section",
		ID:      "urn:section:fiction",
		Updated: "2026-09-21T10:00:00Z",
		Summary: &opds1.Text{
			Body: "Browse all fiction works",
		},
		Links: []opds1.Link{
			{
				Rel:  opds1.RelSubsection,
				Href: "/fiction",
				Type: opds1.TypeNavigation,
			},
		},
	}

	nav := convert.EntryToNavigation(navEntry)
	if nav.Metadata.Title != navEntry.Title {
		t.Errorf("nav.Title = %q; want %q", nav.Metadata.Title, navEntry.Title)
	}
	if nav.Metadata.Identifier != navEntry.ID {
		t.Errorf("nav.Identifier = %q; want %q", nav.Metadata.Identifier, navEntry.ID)
	}
	if nav.Metadata.Modified != navEntry.Updated {
		t.Errorf("nav.Modified = %q; want %q", nav.Metadata.Modified, navEntry.Updated)
	}
	if nav.Metadata.Description != navEntry.Summary.Body {
		t.Errorf("nav.Description = %q; want %q", nav.Metadata.Description, navEntry.Summary.Body)
	}

	backEntry := convert.NavigationToEntry(nav)
	if backEntry.Title != navEntry.Title {
		t.Errorf("backEntry.Title = %q; want %q", backEntry.Title, navEntry.Title)
	}
	if backEntry.ID != navEntry.ID {
		t.Errorf("backEntry.ID = %q; want %q", backEntry.ID, navEntry.ID)
	}
	if backEntry.Updated != navEntry.Updated {
		t.Errorf("backEntry.Updated = %q; want %q", backEntry.Updated, navEntry.Updated)
	}
	if backEntry.Summary == nil || backEntry.Summary.Body != navEntry.Summary.Body {
		t.Errorf("backEntry.Summary = %+v; want %q", backEntry.Summary, navEntry.Summary.Body)
	}
}

func TestCatalog1To2AndBack(t *testing.T) {
	t.Parallel()
	feed := &opds1.Feed{
		ID:      "urn:catalog:test",
		Title:   "Library Catalog",
		Updated: "2026-09-21T00:00:00Z",
		Links: []opds1.Link{
			{
				Rel:  opds1.RelSelf,
				Href: "/opds",
				Type: opds1.TypeNavigation,
			},
		},
		Entries: []opds1.Entry{
			{
				Title:        "Novel",
				ID:           "urn:book:1",
				DCIdentifier: "urn:book:1",
				Updated:      "2026-09-21T00:00:00Z",
				Links: []opds1.Link{
					{
						Rel:  opds1.RelAcquisition,
						Href: "/books/1.epub",
						Type: opds1.MIMEEPUB,
					},
				},
			},
			{
				Title:   "Non-Fiction",
				ID:      "urn:category:nonfiction",
				Updated: "2026-09-21T00:00:00Z",
				Links: []opds1.Link{
					{
						Rel:  opds1.RelSubsection,
						Href: "/categories/nonfiction",
						Type: opds1.TypeNavigation,
					},
				},
			},
		},
	}

	cat := convert.Catalog1To2(feed)
	if cat.Metadata.ID != feed.ID {
		t.Errorf("cat.ID = %q; want %q", cat.Metadata.ID, feed.ID)
	}
	if cat.Metadata.Title != feed.Title {
		t.Errorf("cat.Title = %q; want %q", cat.Metadata.Title, feed.Title)
	}
	if len(cat.Publications) != 1 {
		t.Fatalf("cat.Publications len = %d; want 1", len(cat.Publications))
	}
	if cat.Publications[0].Metadata.Title != "Novel" {
		t.Errorf("cat.Publications[0].Title = %q; want Novel", cat.Publications[0].Metadata.Title)
	}
	if cat.Publications[0].Metadata.Identifier != "urn:book:1" {
		t.Errorf("cat.Publications[0].Identifier = %q; want urn:book:1", cat.Publications[0].Metadata.Identifier)
	}
	if len(cat.Navigation) != 1 {
		t.Fatalf("cat.Navigation len = %d; want 1", len(cat.Navigation))
	}
	if cat.Navigation[0].Metadata.Title != "Non-Fiction" {
		t.Errorf("cat.Navigation[0].Title = %q; want Non-Fiction", cat.Navigation[0].Metadata.Title)
	}
	if cat.Navigation[0].Metadata.Identifier != "urn:category:nonfiction" {
		t.Errorf("cat.Navigation[0].Identifier = %q; want urn:category:nonfiction", cat.Navigation[0].Metadata.Identifier)
	}

	backFeed := convert.Catalog2To1(cat)
	if backFeed.ID != feed.ID {
		t.Errorf("backFeed.ID = %q; want %q", backFeed.ID, feed.ID)
	}
	if len(backFeed.Entries) != 2 {
		t.Fatalf("backFeed.Entries len = %d; want 2", len(backFeed.Entries))
	}
	if backFeed.Entries[0].ID != "urn:book:1" {
		t.Errorf("backFeed.Entries[0].ID = %q; want urn:book:1", backFeed.Entries[0].ID)
	}
	if backFeed.Entries[1].ID != "urn:category:nonfiction" {
		t.Errorf("backFeed.Entries[1].ID = %q; want urn:category:nonfiction", backFeed.Entries[1].ID)
	}
}

func BenchmarkCatalog1To2(b *testing.B) {
	feed := &opds1.Feed{
		ID:      "urn:bench:feed",
		Title:   "Bench Feed",
		Updated: "2026-09-21T00:00:00Z",
		Entries: []opds1.Entry{
			{
				Title: "Novel 1",
				ID:    "urn:book:1",
				Links: []opds1.Link{{Rel: opds1.RelAcquisition, Href: "/b.epub"}},
			},
			{
				Title: "Section 1",
				ID:    "urn:section:1",
				Links: []opds1.Link{{Rel: opds1.RelSubsection, Href: "/s1"}},
			},
		},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = convert.Catalog1To2(feed)
	}
}

func BenchmarkCatalog2To1(b *testing.B) {
	feed := &opds1.Feed{
		ID:      "urn:bench:feed",
		Title:   "Bench Feed",
		Updated: "2026-09-21T00:00:00Z",
		Entries: []opds1.Entry{
			{
				Title: "Novel 1",
				ID:    "urn:book:1",
				Links: []opds1.Link{{Rel: opds1.RelAcquisition, Href: "/b.epub"}},
			},
		},
	}
	cat := convert.Catalog1To2(feed)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = convert.Catalog2To1(cat)
	}
}
