package convert

import (
	"cmp"
	"slices"
	"time"

	"github.com/littfed/go-opds/opds1"
	"github.com/littfed/go-opds/opds2"
)

// Catalog1To2 converts an OPDS 1.2 Feed to an OPDS 2.0 Catalog.
func Catalog1To2(feed *opds1.Feed) *opds2.Catalog {
	if feed == nil {
		return nil
	}

	catalog := &opds2.Catalog{
		Metadata: opds2.CatalogMetadata{
			ID:       feed.ID,
			Title:    feed.Title,
			Modified: feed.Updated,
		},
	}

	// Convert links
	for _, link := range feed.Links {
		catalog.Links = append(catalog.Links, convertLink1To2(link))
	}

	// Convert entries to publications and navigation
	for _, entry := range feed.Entries {
		if hasAcquisitionLink(entry.Links) {
			catalog.Publications = append(catalog.Publications, EntryToPublication(entry))
		} else {
			catalog.Navigation = append(catalog.Navigation, EntryToNavigation(entry))
		}
	}

	return catalog
}

// Catalog2To1 converts an OPDS 2.0 Catalog to an OPDS 1.2 Feed.
func Catalog2To1(catalog *opds2.Catalog) *opds1.Feed {
	if catalog == nil {
		return nil
	}

	feed := &opds1.Feed{
		ID:      catalog.Metadata.ID,
		Title:   catalog.Metadata.Title,
		Updated: catalog.Metadata.Modified,
	}

	// Convert links
	for _, link := range catalog.Links {
		feed.Links = append(feed.Links, convertLink2To1(link))
	}

	// Convert publications to entries
	for _, pub := range catalog.Publications {
		feed.Entries = append(feed.Entries, PublicationToEntry(pub))
	}

	// Convert navigation to entries
	for _, nav := range catalog.Navigation {
		feed.Entries = append(feed.Entries, NavigationToEntry(nav))
	}

	return feed
}

func hasAcquisitionLink(links []opds1.Link) bool {
	return slices.ContainsFunc(links, func(l opds1.Link) bool {
		return IsAcquisitionRel(l.Rel)
	})
}

func convertLink1To2(l opds1.Link) opds2.Link {
	return opds2.Link{
		Rel:   l.Rel,
		Href:  l.Href,
		Type:  l.Type,
		Title: l.Title,
	}
}

func convertLink2To1(l opds2.Link) opds1.Link {
	return opds1.Link{
		Rel:   l.Rel,
		Href:  l.Href,
		Type:  l.Type,
		Title: l.Title,
	}
}

// IsAcquisitionRel returns true if the relation string is an OPDS acquisition relation.
func IsAcquisitionRel(rel string) bool {
	return rel == opds1.RelAcquisition ||
		rel == opds1.RelAcquisitionBuy ||
		rel == opds1.RelAcquisitionBorrow ||
		rel == opds1.RelAcquisitionOpenAccess ||
		rel == opds1.RelAcquisitionSample ||
		rel == opds1.RelAcquisitionSubscribe
}

// EntryToPublication translates an OPDS 1.2 Entry into an OPDS 2.0 Publication without data loss.
func EntryToPublication(entry opds1.Entry) opds2.Publication {
	pub := opds2.Publication{
		Metadata: opds2.PublicationMetadata{
			Title:      entry.Title,
			Identifier: entry.Identifier(),
			Publisher:  entry.Publisher(),
			Language:   entry.Language(),
			Rights:     entry.Rights,
			Published:  entry.Issued(),
			Modified:   entry.Updated,
		},
	}

	if entry.Author != nil {
		pub.Metadata.Author = append(pub.Metadata.Author, opds2.Author{
			Name:       entry.Author.Name,
			Identifier: entry.Author.URI,
		})
	}

	if entry.Summary != nil {
		pub.Metadata.Description = entry.Summary.Body
	} else if entry.Content != nil {
		pub.Metadata.Description = entry.Content.Body
	}

	for _, cat := range entry.Category {
		pub.Metadata.Subject = append(pub.Metadata.Subject, cat.Term)
	}

	// Convert links using helper
	for _, link := range entry.Links {
		pub.Links = append(pub.Links, convertLink1To2(link))
	}

	return pub
}

// EntryToNavigation translates an OPDS 1.2 Entry into an OPDS 2.0 Navigation item preserving ID and timestamps.
func EntryToNavigation(entry opds1.Entry) opds2.Navigation {
	nav := opds2.Navigation{
		Metadata: opds2.NavigationMetadata{
			Title:       entry.Title,
			Identifier:  entry.ID,
			Modified:    entry.Updated,
		},
	}

	if entry.Summary != nil {
		nav.Metadata.Description = entry.Summary.Body
	}

	for _, link := range entry.Links {
		nav.Links = append(nav.Links, convertLink1To2(link))
	}

	return nav
}

// PublicationToEntry translates an OPDS 2.0 Publication into an OPDS 1.2 Entry preserving timestamps and metadata.
func PublicationToEntry(pub opds2.Publication) opds1.Entry {
	updated := cmp.Or(pub.Metadata.Modified, pub.Metadata.Published, time.Now().UTC().Format(time.RFC3339))

	entry := opds1.Entry{
		Title:        pub.Metadata.Title,
		ID:           pub.Metadata.Identifier,
		DCIdentifier: pub.Metadata.Identifier,
		Updated:      updated,
		DCLanguage:   pub.Metadata.Language,
		DCPublisher:  pub.Metadata.Publisher,
		Rights:       pub.Metadata.Rights,
		Published:    pub.Metadata.Published,
		DCIssued:     pub.Metadata.Published,
	}

	if len(pub.Metadata.Author) > 0 {
		entry.Author = &opds1.Person{
			Name: pub.Metadata.Author[0].Name,
			URI:  pub.Metadata.Author[0].Identifier,
		}
	}

	if pub.Metadata.Description != "" {
		entry.Summary = &opds1.Text{
			Type: "text",
			Body: pub.Metadata.Description,
		}
	}

	for _, subject := range pub.Metadata.Subject {
		entry.Category = append(entry.Category, opds1.Category{
			Term: subject,
		})
	}

	for _, link := range pub.Links {
		entry.Links = append(entry.Links, convertLink2To1(link))
	}

	return entry
}

// NavigationToEntry translates an OPDS 2.0 Navigation item into an OPDS 1.2 Entry with non-empty ID and timestamps.
func NavigationToEntry(nav opds2.Navigation) opds1.Entry {
	id := cmp.Or(nav.Metadata.Identifier, "urn:nav:"+nav.Metadata.Title)
	updated := cmp.Or(nav.Metadata.Modified, time.Now().UTC().Format(time.RFC3339))

	entry := opds1.Entry{
		Title:   nav.Metadata.Title,
		ID:      id,
		Updated: updated,
	}

	if nav.Metadata.Description != "" {
		entry.Summary = &opds1.Text{
			Type: "text",
			Body: nav.Metadata.Description,
		}
	}

	for _, link := range nav.Links {
		entry.Links = append(entry.Links, convertLink2To1(link))
	}

	return entry
}
