package opds2

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/littfed/go-opds/internal/common"
)

// ErrValidation indicates that a model failed OPDS 2.0 protocol invariant validation.
var ErrValidation = errors.New("opds2: validation error")

// Validate checks that the Catalog satisfies OPDS 2.0 protocol invariants.
func (c *Catalog) Validate() error {
	if c == nil {
		return fmt.Errorf("%w: catalog is nil", ErrValidation)
	}
	if c.Metadata.ID == "" {
		return fmt.Errorf("%w: catalog ID is required", ErrValidation)
	}
	if c.Metadata.Title == "" {
		return fmt.Errorf("%w: catalog Title is required", ErrValidation)
	}
	if c.Metadata.Modified != "" {
		if err := validateTimestamp(c.Metadata.Modified); err != nil {
			return fmt.Errorf("%w: invalid Modified timestamp %q: %w", ErrValidation, c.Metadata.Modified, err)
		}
	}
	return validateLinks(c.Links)
}

// Validate checks that the Publication satisfies OPDS 2.0 protocol invariants.
func (p *Publication) Validate() error {
	if p == nil {
		return fmt.Errorf("%w: publication is nil", ErrValidation)
	}
	if p.Metadata.Identifier == "" {
		return fmt.Errorf("%w: publication Identifier is required", ErrValidation)
	}
	if p.Metadata.Title == "" {
		return fmt.Errorf("%w: publication Title is required", ErrValidation)
	}
	if p.Metadata.Published != "" {
		if err := validatePublishedDate(p.Metadata.Published); err != nil {
			return fmt.Errorf("%w: invalid Published date %q: %w", ErrValidation, p.Metadata.Published, err)
		}
	}
	if p.Metadata.Modified != "" {
		if err := validateTimestamp(p.Metadata.Modified); err != nil {
			return fmt.Errorf("%w: invalid Modified timestamp %q: %w", ErrValidation, p.Metadata.Modified, err)
		}
	}
	return validateLinks(p.Links)
}

// Validate checks that the Navigation element satisfies OPDS 2.0 protocol invariants.
func (n *Navigation) Validate() error {
	if n == nil {
		return fmt.Errorf("%w: navigation is nil", ErrValidation)
	}
	if n.Metadata.Title == "" {
		return fmt.Errorf("%w: navigation Title is required", ErrValidation)
	}
	if n.Metadata.Modified != "" {
		if err := validateTimestamp(n.Metadata.Modified); err != nil {
			return fmt.Errorf("%w: invalid Modified timestamp %q: %w", ErrValidation, n.Metadata.Modified, err)
		}
	}
	return validateLinks(n.Links)
}

func validateLinks(links []Link) error {
	for _, link := range links {
		if isUnsafeScheme(link.Href) {
			return fmt.Errorf("%w: unsafe URI scheme in link href %q", ErrValidation, link.Href)
		}
	}
	return nil
}

func isUnsafeScheme(href string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(href))
	return strings.HasPrefix(trimmed, "javascript:") ||
		strings.HasPrefix(trimmed, "vbscript:") ||
		strings.HasPrefix(trimmed, "data:text/html")
}

func validateTimestamp(val string) error {
	_, err := time.Parse(time.RFC3339, val)
	if err == nil {
		return nil
	}
	_, errNano := time.Parse(time.RFC3339Nano, val)
	if errNano == nil {
		return nil
	}
	return fmt.Errorf("%w (also tried RFC3339Nano: %w)", err, errNano)
}

func validatePublishedDate(val string) error {
	_, err := time.Parse(time.RFC3339, val)
	if err == nil {
		return nil
	}
	_, errDate := time.Parse("2006-01-02", val)
	if errDate == nil {
		return nil
	}
	return fmt.Errorf("%w (also tried YYYY-MM-DD: %w)", err, errDate)
}

// CatalogBuilder provides a fluent API for building OPDS 2.0 catalogs.
type CatalogBuilder struct {
	catalog Catalog
}

// NewCatalogBuilder creates a new CatalogBuilder with a title.
func NewCatalogBuilder(title string) *CatalogBuilder {
	return &CatalogBuilder{
		catalog: Catalog{
			Metadata: CatalogMetadata{
				Title:    title,
				Modified: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
}

// NewCatalogBuilderWithID creates a new CatalogBuilder with an ID and title.
func NewCatalogBuilderWithID(id, title string) *CatalogBuilder {
	b := NewCatalogBuilder(title)
	b.catalog.Metadata.ID = id
	return b
}

// ID sets the catalog ID.
func (b *CatalogBuilder) ID(id string) *CatalogBuilder {
	b.catalog.Metadata.ID = id
	return b
}

// Modified sets the catalog modified time.
func (b *CatalogBuilder) Modified(modified string) *CatalogBuilder {
	b.catalog.Metadata.Modified = modified
	return b
}

// AddLink adds a link to the catalog.
func (b *CatalogBuilder) AddLink(link Link) *CatalogBuilder {
	b.catalog.Links = append(b.catalog.Links, link)
	return b
}

// AddPublication adds a publication pointer to the catalog.
func (b *CatalogBuilder) AddPublication(pub *Publication) *CatalogBuilder {
	if pub != nil {
		b.catalog.Publications = append(b.catalog.Publications, *pub)
	}
	return b
}

// AddPublicationValue adds a publication value to the catalog.
func (b *CatalogBuilder) AddPublicationValue(pub Publication) *CatalogBuilder {
	b.catalog.Publications = append(b.catalog.Publications, pub)
	return b
}

// AddNavigation adds a navigation pointer to the catalog.
func (b *CatalogBuilder) AddNavigation(nav *Navigation) *CatalogBuilder {
	if nav != nil {
		b.catalog.Navigation = append(b.catalog.Navigation, *nav)
	}
	return b
}

// AddNavigationValue adds a navigation value to the catalog.
func (b *CatalogBuilder) AddNavigationValue(nav Navigation) *CatalogBuilder {
	b.catalog.Navigation = append(b.catalog.Navigation, nav)
	return b
}

// AddGroup adds a group to the catalog.
func (b *CatalogBuilder) AddGroup(group Group) *CatalogBuilder {
	b.catalog.Groups = append(b.catalog.Groups, group)
	return b
}

// Build validates invariants and returns the constructed Catalog pointer.
func (b *CatalogBuilder) Build() (*Catalog, error) {
	b.catalog.Metadata.Modified = cmp.Or(b.catalog.Metadata.Modified, time.Now().UTC().Format(time.RFC3339))
	if err := b.catalog.Validate(); err != nil {
		return nil, err
	}
	cp := b.catalog
	cp.Links = slices.Clone(b.catalog.Links)
	cp.Publications = slices.Clone(b.catalog.Publications)
	cp.Navigation = slices.Clone(b.catalog.Navigation)
	cp.Groups = slices.Clone(b.catalog.Groups)
	return &cp, nil
}

// MustBuild returns the constructed Catalog pointer, panicking if invariant validation fails.
func (b *CatalogBuilder) MustBuild() *Catalog {
	cat, err := b.Build()
	if err != nil {
		panic(err)
	}
	return cat
}

// PublicationBuilder provides a fluent API for building OPDS 2.0 publications.
type PublicationBuilder struct {
	pub Publication
}

// NewPublicationBuilder creates a new PublicationBuilder.
func NewPublicationBuilder(title string) *PublicationBuilder {
	return &PublicationBuilder{
		pub: Publication{
			Metadata: PublicationMetadata{
				Title: title,
			},
		},
	}
}

// NewPublicationBuilderWithID creates a new PublicationBuilder with an identifier and title.
func NewPublicationBuilderWithID(identifier, title string) *PublicationBuilder {
	b := NewPublicationBuilder(title)
	b.pub.Metadata.Identifier = identifier
	return b
}

// Identifier sets the publication identifier.
func (b *PublicationBuilder) Identifier(id string) *PublicationBuilder {
	b.pub.Metadata.Identifier = id
	return b
}

// Author adds an author to the publication.
func (b *PublicationBuilder) Author(name, identifier string) *PublicationBuilder {
	b.pub.Metadata.Author = append(b.pub.Metadata.Author, Author{Name: name, Identifier: identifier})
	return b
}

// Publisher sets the publisher.
func (b *PublicationBuilder) Publisher(publisher string) *PublicationBuilder {
	b.pub.Metadata.Publisher = publisher
	return b
}

// Language sets the language.
func (b *PublicationBuilder) Language(lang string) *PublicationBuilder {
	b.pub.Metadata.Language = lang
	return b
}

// Description sets the description.
func (b *PublicationBuilder) Description(desc string) *PublicationBuilder {
	b.pub.Metadata.Description = desc
	return b
}

// Subject adds a subject.
func (b *PublicationBuilder) Subject(subject string) *PublicationBuilder {
	b.pub.Metadata.Subject = append(b.pub.Metadata.Subject, subject)
	return b
}

// Rights sets the rights.
func (b *PublicationBuilder) Rights(rights string) *PublicationBuilder {
	b.pub.Metadata.Rights = rights
	return b
}

// Published sets the published date.
func (b *PublicationBuilder) Published(published string) *PublicationBuilder {
	b.pub.Metadata.Published = published
	return b
}

// Modified sets the modified timestamp.
func (b *PublicationBuilder) Modified(modified string) *PublicationBuilder {
	b.pub.Metadata.Modified = modified
	return b
}

// NumberOfPages sets the number of pages.
func (b *PublicationBuilder) NumberOfPages(pages int) *PublicationBuilder {
	b.pub.Metadata.NumberOfPages = pages
	return b
}

// AddLink adds a link to the publication.
func (b *PublicationBuilder) AddLink(link Link) *PublicationBuilder {
	b.pub.Links = append(b.pub.Links, link)
	return b
}

// AddAcquisitionLink adds an acquisition link to the publication, auto-detecting media type from href if empty.
func (b *PublicationBuilder) AddAcquisitionLink(href, mediaType string) *PublicationBuilder {
	mediaType = cmp.Or(mediaType, common.DetectMIMEType(href))
	b.pub.Links = append(b.pub.Links, Link{
		Rel:  RelAcquisition,
		Href: href,
		Type: mediaType,
	})
	return b
}

// AddImage adds an image to the publication.
func (b *PublicationBuilder) AddImage(img Image) *PublicationBuilder {
	b.pub.Images = append(b.pub.Images, img)
	return b
}

// Build validates invariants and returns the constructed Publication pointer.
func (b *PublicationBuilder) Build() (*Publication, error) {
	if err := b.pub.Validate(); err != nil {
		return nil, err
	}
	cp := b.pub
	cp.Links = slices.Clone(b.pub.Links)
	cp.Images = slices.Clone(b.pub.Images)
	cp.Metadata.Author = slices.Clone(b.pub.Metadata.Author)
	cp.Metadata.Subject = slices.Clone(b.pub.Metadata.Subject)
	return &cp, nil
}

// MustBuild returns the constructed Publication pointer, panicking if invariant validation fails.
func (b *PublicationBuilder) MustBuild() *Publication {
	pub, err := b.Build()
	if err != nil {
		panic(err)
	}
	return pub
}

// NavigationBuilder provides a fluent API for building OPDS 2.0 navigation elements.
type NavigationBuilder struct {
	nav Navigation
}

// NewNavigationBuilder creates a new NavigationBuilder.
func NewNavigationBuilder(title string) *NavigationBuilder {
	return &NavigationBuilder{
		nav: Navigation{
			Metadata: NavigationMetadata{
				Title: title,
			},
		},
	}
}

// NewNavigationBuilderWithID creates a new NavigationBuilder with an ID and title.
func NewNavigationBuilderWithID(identifier, title string) *NavigationBuilder {
	b := NewNavigationBuilder(title)
	b.nav.Metadata.Identifier = identifier
	return b
}

// Identifier sets the navigation identifier.
func (b *NavigationBuilder) Identifier(identifier string) *NavigationBuilder {
	b.nav.Metadata.Identifier = identifier
	return b
}

// Description sets the navigation description.
func (b *NavigationBuilder) Description(desc string) *NavigationBuilder {
	b.nav.Metadata.Description = desc
	return b
}

// Modified sets the navigation modified timestamp.
func (b *NavigationBuilder) Modified(modified string) *NavigationBuilder {
	b.nav.Metadata.Modified = modified
	return b
}

// AddLink adds a link to the navigation.
func (b *NavigationBuilder) AddLink(link Link) *NavigationBuilder {
	b.nav.Links = append(b.nav.Links, link)
	return b
}

// Build validates invariants and returns the constructed Navigation pointer.
func (b *NavigationBuilder) Build() (*Navigation, error) {
	if err := b.nav.Validate(); err != nil {
		return nil, err
	}
	cp := b.nav
	cp.Links = slices.Clone(b.nav.Links)
	return &cp, nil
}

// MustBuild returns the constructed Navigation pointer, panicking if invariant validation fails.
func (b *NavigationBuilder) MustBuild() *Navigation {
	nav, err := b.Build()
	if err != nil {
		panic(err)
	}
	return nav
}

// LinkBuilder provides a fluent API for building OPDS 2.0 links.
type LinkBuilder struct {
	link Link
}

// NewLinkBuilder creates a new empty LinkBuilder.
func NewLinkBuilder() *LinkBuilder {
	return &LinkBuilder{}
}

// Rel sets the link relation.
func (b *LinkBuilder) Rel(rel string) *LinkBuilder {
	b.link.Rel = rel
	return b
}

// Href sets the link href, and auto-detects Type from extension if not yet set.
func (b *LinkBuilder) Href(href string) *LinkBuilder {
	b.link.Href = href
	b.link.Type = cmp.Or(b.link.Type, common.DetectMIMEType(href))
	return b
}

// Type sets the link type.
func (b *LinkBuilder) Type(typ string) *LinkBuilder {
	b.link.Type = typ
	return b
}

// Title sets the link title.
func (b *LinkBuilder) Title(title string) *LinkBuilder {
	b.link.Title = title
	return b
}

// Templated marks this link as a URI template.
func (b *LinkBuilder) Templated(templated bool) *LinkBuilder {
	b.link.Templated = templated
	return b
}

// Build returns the constructed Link.
func (b *LinkBuilder) Build() Link {
	return b.link
}
