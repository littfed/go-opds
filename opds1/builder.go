package opds1

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/littfed/go-opds/internal/common"
)

// ErrValidation indicates that a model failed OPDS 1.2 protocol invariant validation.
var ErrValidation = errors.New("opds1: validation error")

// Validate checks that the Feed satisfies OPDS 1.2 protocol invariants.
func (f *Feed) Validate() error {
	if f == nil {
		return fmt.Errorf("%w: feed is nil", ErrValidation)
	}
	if f.ID == "" {
		return fmt.Errorf("%w: feed ID is required", ErrValidation)
	}
	if f.Title == "" {
		return fmt.Errorf("%w: feed Title is required", ErrValidation)
	}
	if f.Updated != "" {
		if err1 := validateTimestamp(f.Updated); err1 != nil {
			return fmt.Errorf("%w: invalid Updated timestamp %q: %w", ErrValidation, f.Updated, err1)
		}
	}
	return validateLinks(f.Links)
}

// Validate checks that the Entry satisfies OPDS 1.2 protocol invariants.
func (e *Entry) Validate() error {
	if e == nil {
		return fmt.Errorf("%w: entry is nil", ErrValidation)
	}
	if e.ID == "" {
		return fmt.Errorf("%w: entry ID is required", ErrValidation)
	}
	if e.Title == "" {
		return fmt.Errorf("%w: entry Title is required", ErrValidation)
	}
	if e.Updated != "" {
		if err1 := validateTimestamp(e.Updated); err1 != nil {
			return fmt.Errorf("%w: invalid Updated timestamp %q: %w", ErrValidation, e.Updated, err1)
		}
	}
	return validateLinks(e.Links)
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

// FeedBuilder provides a fluent API for building OPDS 1.2 feeds.
type FeedBuilder struct {
	feed Feed
}

// NewFeedBuilder creates a new FeedBuilder with the required fields.
func NewFeedBuilder(id, title, updated string) *FeedBuilder {
	return &FeedBuilder{
		feed: Feed{
			ID:      id,
			Title:   title,
			Updated: updated,
		},
	}
}

// NewFeedBuilderNow creates a new FeedBuilder using the current time for Updated.
func NewFeedBuilderNow(id, title string) *FeedBuilder {
	return NewFeedBuilder(id, title, time.Now().UTC().Format(time.RFC3339))
}

// Subtitle sets the feed subtitle.
func (b *FeedBuilder) Subtitle(subtitle string) *FeedBuilder {
	b.feed.Subtitle = subtitle
	return b
}

// Author sets the feed author.
func (b *FeedBuilder) Author(name, uri string) *FeedBuilder {
	b.feed.Author = &Person{Name: name, URI: uri}
	return b
}

// AddLink adds a link to the feed.
func (b *FeedBuilder) AddLink(link Link) *FeedBuilder {
	b.feed.Links = append(b.feed.Links, link)
	return b
}

// AddEntry adds an entry pointer to the feed.
func (b *FeedBuilder) AddEntry(entry *Entry) *FeedBuilder {
	if entry != nil {
		b.feed.Entries = append(b.feed.Entries, *entry)
	}
	return b
}

// AddEntryValue adds an entry value to the feed.
func (b *FeedBuilder) AddEntryValue(entry Entry) *FeedBuilder {
	b.feed.Entries = append(b.feed.Entries, entry)
	return b
}

// Build validates invariants and returns the constructed Feed pointer.
func (b *FeedBuilder) Build() (*Feed, error) {
	b.feed.Updated = cmp.Or(b.feed.Updated, time.Now().UTC().Format(time.RFC3339))
	if err := b.feed.Validate(); err != nil {
		return nil, err
	}
	cp := b.feed
	cp.Links = slices.Clone(b.feed.Links)
	cp.Entries = slices.Clone(b.feed.Entries)
	return &cp, nil
}

// MustBuild returns the constructed Feed pointer, panicking if invariant validation fails.
func (b *FeedBuilder) MustBuild() *Feed {
	feed, err := b.Build()
	if err != nil {
		panic(err)
	}
	return feed
}

// EntryBuilder provides a fluent API for building OPDS 1.2 entries.
type EntryBuilder struct {
	entry Entry
}

// NewEntryBuilder creates a new EntryBuilder with the required fields.
func NewEntryBuilder(id, title, updated string) *EntryBuilder {
	return &EntryBuilder{
		entry: Entry{
			ID:      id,
			Title:   title,
			Updated: updated,
		},
	}
}

// NewEntryBuilderNow creates a new EntryBuilder using the current time for Updated.
func NewEntryBuilderNow(id, title string) *EntryBuilder {
	return NewEntryBuilder(id, title, time.Now().UTC().Format(time.RFC3339))
}

// Published sets the entry published date.
func (b *EntryBuilder) Published(published string) *EntryBuilder {
	b.entry.Published = published
	return b
}

// Author sets the entry author.
func (b *EntryBuilder) Author(name, uri string) *EntryBuilder {
	b.entry.Author = &Person{Name: name, URI: uri}
	return b
}

// Summary sets the entry summary.
func (b *EntryBuilder) Summary(text string) *EntryBuilder {
	b.entry.Summary = &Text{Type: "text", Body: text}
	return b
}

// Content sets the entry content.
func (b *EntryBuilder) Content(typ, body string) *EntryBuilder {
	b.entry.Content = &Text{Type: typ, Body: body}
	return b
}

// AddLink adds a link to the entry.
func (b *EntryBuilder) AddLink(link Link) *EntryBuilder {
	b.entry.Links = append(b.entry.Links, link)
	return b
}

// AddCategory adds a category to the entry.
func (b *EntryBuilder) AddCategory(cat Category) *EntryBuilder {
	b.entry.Category = append(b.entry.Category, cat)
	return b
}

// Rights sets the entry rights.
func (b *EntryBuilder) Rights(rights string) *EntryBuilder {
	b.entry.Rights = rights
	return b
}

// DCLanguage sets the Dublin Core language.
func (b *EntryBuilder) DCLanguage(lang string) *EntryBuilder {
	b.entry.DCLanguage = lang
	return b
}

// DCIssued sets the Dublin Core issued date.
func (b *EntryBuilder) DCIssued(issued string) *EntryBuilder {
	b.entry.DCIssued = issued
	return b
}

// DCIdentifier sets the Dublin Core identifier.
func (b *EntryBuilder) DCIdentifier(id string) *EntryBuilder {
	b.entry.DCIdentifier = id
	return b
}

// DCPublisher sets the Dublin Core publisher.
func (b *EntryBuilder) DCPublisher(publisher string) *EntryBuilder {
	b.entry.DCPublisher = publisher
	return b
}

// Build validates invariants and returns the constructed Entry pointer.
func (b *EntryBuilder) Build() (*Entry, error) {
	b.entry.Updated = cmp.Or(b.entry.Updated, time.Now().UTC().Format(time.RFC3339))
	if err := b.entry.Validate(); err != nil {
		return nil, err
	}
	cp := b.entry
	cp.Links = slices.Clone(b.entry.Links)
	cp.Category = slices.Clone(b.entry.Category)
	return &cp, nil
}

// MustBuild returns the constructed Entry pointer, panicking if invariant validation fails.
func (b *EntryBuilder) MustBuild() *Entry {
	entry, err := b.Build()
	if err != nil {
		panic(err)
	}
	return entry
}

// LinkBuilder provides a fluent API for building OPDS 1.2 links.
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

// Length sets the link content length.
func (b *LinkBuilder) Length(length int64) *LinkBuilder {
	b.link.Length = length
	return b
}

// Build returns the constructed Link.
func (b *LinkBuilder) Build() Link {
	return b.link
}
