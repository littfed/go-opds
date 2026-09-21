package opds2

// Catalog represents an OPDS 2.0 catalog document.
type Catalog struct {
	Metadata     CatalogMetadata `json:"metadata"`
	Links        []Link          `json:"links,omitempty"`
	Publications []Publication   `json:"publications,omitempty"`
	Navigation   []Navigation    `json:"navigation,omitempty"`
	Groups       []Group         `json:"groups,omitempty"`
}

// CatalogMetadata contains metadata for the catalog.
type CatalogMetadata struct {
	Title    string `json:"title"`
	Modified string `json:"modified,omitempty"`
	ID       string `json:"id,omitempty"`
}

// Publication represents an OPDS 2.0 publication.
type Publication struct {
	Metadata PublicationMetadata `json:"metadata"`
	Links    []Link              `json:"links"`
	Images   []Image             `json:"images,omitempty"`
}

// PublicationMetadata contains metadata for a publication.
type PublicationMetadata struct {
	Title         string   `json:"title"`
	Identifier    string   `json:"identifier,omitempty"`
	Author        []Author `json:"author,omitempty"`
	Publisher     string   `json:"publisher,omitempty"`
	Language      string   `json:"language,omitempty"`
	Description   string   `json:"description,omitempty"`
	Subject       []string `json:"subject,omitempty"`
	Rights        string   `json:"rights,omitempty"`
	NumberOfPages int      `json:"numberOfPages,omitempty"`
	Published     string   `json:"published,omitempty"`
	Modified      string   `json:"modified,omitempty"`
}

// Navigation represents an OPDS 2.0 navigation element.
type Navigation struct {
	Metadata NavigationMetadata `json:"metadata"`
	Links    []Link             `json:"links"`
}

// NavigationMetadata contains metadata for a navigation element.
type NavigationMetadata struct {
	Title       string `json:"title"`
	Identifier  string `json:"identifier,omitempty"`
	Description string `json:"description,omitempty"`
	Modified    string `json:"modified,omitempty"`
}

// Link represents an OPDS 2.0 link.
type Link struct {
	Rel      string `json:"rel,omitempty"`
	Href     string `json:"href"`
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Templated bool  `json:"templated,omitempty"`
}

// Author represents an OPDS 2.0 author.
type Author struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier,omitempty"`
}

// Image represents an OPDS 2.0 image.
type Image struct {
	Href   string `json:"href"`
	Type   string `json:"type,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// Group represents an OPDS 2.0 group.
type Group struct {
	Metadata   GroupMetadata `json:"metadata"`
	Links      []Link        `json:"links,omitempty"`
	Navigation []Navigation  `json:"navigation,omitempty"`
}

// GroupMetadata contains metadata for a group.
type GroupMetadata struct {
	Title string `json:"title"`
}
