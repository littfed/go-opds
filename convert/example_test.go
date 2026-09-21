package convert_test

import (
	"fmt"

	"github.com/littfed/go-opds/convert"
	"github.com/littfed/go-opds/opds1"
	"github.com/littfed/go-opds/opds2"
)

func ExampleCatalog1To2() {
	feed := opds1.NewFeedBuilder("urn:catalog:root", "Public Catalog", "2024-01-01T00:00:00Z").
		AddEntry(opds1.NewEntryBuilder("urn:book:1", "Sample Novel", "2024-01-01T00:00:00Z").
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelAcquisition).
				Href("/books/novel.epub").
				Build()).
			MustBuild()).
		MustBuild()

	catalog := convert.Catalog1To2(feed)

	fmt.Println(catalog.Metadata.ID)
	fmt.Println(catalog.Metadata.Title)
	fmt.Println(len(catalog.Publications))
	// Output:
	// urn:catalog:root
	// Public Catalog
	// 1
}

func ExampleCatalog2To1() {
	catalog := opds2.NewCatalogBuilderWithID("urn:catalog:root", "Public Catalog").
		AddPublication(opds2.NewPublicationBuilderWithID("urn:book:1", "Sample Novel").
			AddAcquisitionLink("/books/novel.epub", "").
			MustBuild()).
		MustBuild()

	feed := convert.Catalog2To1(catalog)

	fmt.Println(feed.ID)
	fmt.Println(feed.Title)
	fmt.Println(len(feed.Entries))
	// Output:
	// urn:catalog:root
	// Public Catalog
	// 1
}
