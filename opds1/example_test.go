package opds1_test

import (
	"fmt"

	"github.com/littfed/go-opds/opds1"
)

func ExampleNewFeedBuilder() {
	feed := opds1.NewFeedBuilder("urn:catalog:root", "Public Catalog", "2024-01-01T00:00:00Z").
		Subtitle("Free eBooks").
		Author("Library Staff", "https://library.example.com").
		AddLink(opds1.NewLinkBuilder().
			Rel(opds1.RelSelf).
			Href("/opds/root.xml").
			Type(opds1.TypeNavigation).
			Build()).
		AddEntry(opds1.NewEntryBuilder("urn:book:1", "Sample Novel", "2024-01-01T00:00:00Z").
			Summary("An exciting adventure novel.").
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelAcquisition).
				Href("/books/novel.epub").
				Build()).
			MustBuild()).
		MustBuild()

	fmt.Println(feed.Title)
	fmt.Println(feed.Subtitle)
	fmt.Println(len(feed.Entries))
	// Output:
	// Public Catalog
	// Free eBooks
	// 1
}

func ExampleParse() {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>urn:example:feed</id>
  <title>Example OPDS Feed</title>
  <updated>2024-01-01T00:00:00Z</updated>
</feed>`)

	feed, err := opds1.Parse(xmlData)
	if err != nil {
		panic(err)
	}

	fmt.Println(feed.ID)
	fmt.Println(feed.Title)
	// Output:
	// urn:example:feed
	// Example OPDS Feed
}
