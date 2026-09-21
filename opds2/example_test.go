package opds2_test

import (
	"fmt"

	"github.com/littfed/go-opds/opds2"
)

func ExampleNewCatalogBuilder() {
	catalog := opds2.NewCatalogBuilderWithID("urn:catalog:root", "Public Catalog").
		AddLink(opds2.NewLinkBuilder().
			Rel(opds2.RelSelf).
			Href("/opds/catalog.json").
			Type(opds2.TypeNavigation).
			Build()).
		AddPublication(opds2.NewPublicationBuilderWithID("urn:book:1", "Sample Novel").
			Language("en").
			AddAcquisitionLink("/books/novel.epub", "").
			MustBuild()).
		MustBuild()

	fmt.Println(catalog.Metadata.ID)
	fmt.Println(catalog.Metadata.Title)
	fmt.Println(len(catalog.Publications))
	// Output:
	// urn:catalog:root
	// Public Catalog
	// 1
}

func ExampleParse() {
	jsonData := []byte(`{
  "metadata": {
    "title": "Example Catalog",
    "id": "urn:example:cat"
  }
}`)

	catalog, err := opds2.Parse(jsonData)
	if err != nil {
		panic(err)
	}

	fmt.Println(catalog.Metadata.ID)
	fmt.Println(catalog.Metadata.Title)
	// Output:
	// urn:example:cat
	// Example Catalog
}
