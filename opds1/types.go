package opds1

import "encoding/xml"

// Feed represents an OPDS 1.2 Catalog Feed Document (atom:feed).
type Feed struct {
	XMLName  xml.Name  `xml:"http://www.w3.org/2005/Atom feed"`
	ID       string    `xml:"id"`
	Title    string    `xml:"title"`
	Subtitle string    `xml:"subtitle,omitempty"`
	Updated  string    `xml:"updated"`
	Author   *Person   `xml:"author,omitempty"`
	Links    []Link    `xml:"link"`
	Entries  []Entry   `xml:"entry,omitempty"`
}

// Entry represents an OPDS 1.2 Catalog Entry (atom:entry).
type Entry struct {
	Title        string     `xml:"title"`
	ID           string     `xml:"id"`
	Updated      string     `xml:"updated"`
	Published    string     `xml:"published,omitempty"`
	Author       *Person    `xml:"author,omitempty"`
	Summary      *Text      `xml:"summary,omitempty"`
	Content      *Text      `xml:"content,omitempty"`
	Links        []Link     `xml:"link"`
	Category     []Category `xml:"category,omitempty"`
	Rights       string     `xml:"rights,omitempty"`
	DCLanguage        string     `xml:"dc:language,omitempty"`
	DCIssued          string     `xml:"dc:issued,omitempty"`
	DCIdentifier      string     `xml:"dc:identifier,omitempty"`
	DCPublisher       string     `xml:"dc:publisher,omitempty"`
	DCTermsLanguage   string     `xml:"http://purl.org/dc/terms/ language,omitempty"`
	DCTermsIssued     string     `xml:"http://purl.org/dc/terms/ issued,omitempty"`
	DCTermsIdentifier string     `xml:"http://purl.org/dc/terms/ identifier,omitempty"`
	DCTermsPublisher  string     `xml:"http://purl.org/dc/terms/ publisher,omitempty"`
}

// Link represents an atom:link element with optional OPDS facet attributes.
type Link struct {
	Rel         string `xml:"rel,attr,omitempty"`
	Href        string `xml:"href,attr"`
	Type        string `xml:"type,attr,omitempty"`
	Title       string `xml:"title,attr,omitempty"`
	HrefLang    string `xml:"hreflang,attr,omitempty"`
	Length      int64  `xml:"length,attr,omitempty"`
	FacetGroup  string `xml:"http://opds-spec.org/2010/catalog facetGroup,attr,omitempty"`
	ActiveFacet string `xml:"http://opds-spec.org/2010/catalog activeFacet,attr,omitempty"`
}

// Person represents an Atom person construct (author/contributor).
type Person struct {
	Name  string `xml:"name"`
	URI   string `xml:"uri,omitempty"`
	Email string `xml:"email,omitempty"`
}

// Text represents an Atom text construct (summary/content).
type Text struct {
	Type string `xml:"type,attr,omitempty"`
	Body string `xml:",chardata"`
}

// Category represents an Atom category element.
type Category struct {
	Term   string `xml:"term,attr"`
	Label  string `xml:"label,attr,omitempty"`
	Scheme string `xml:"scheme,attr,omitempty"`
}

// Price represents an opds:price element.
type Price struct {
	CurrencyCode string `xml:"currencycode,attr"`
	Value        string `xml:",chardata"`
}

// IndirectAcquisition represents an opds:indirectAcquisition element.
type IndirectAcquisition struct {
	Type               string                 `xml:"type,attr"`
	IndirectAcquisition []IndirectAcquisition `xml:"indirectAcquisition,omitempty"`
}

// Language returns the language code, checking DCLanguage and DCTermsLanguage.
func (e *Entry) Language() string {
	if e == nil {
		return ""
	}
	if e.DCLanguage != "" {
		return e.DCLanguage
	}
	return e.DCTermsLanguage
}

// Issued returns the publication/issued date string, checking DCIssued, DCTermsIssued, and Published.
func (e *Entry) Issued() string {
	if e == nil {
		return ""
	}
	if e.DCIssued != "" {
		return e.DCIssued
	}
	if e.DCTermsIssued != "" {
		return e.DCTermsIssued
	}
	return e.Published
}

// Identifier returns the catalog identifier, checking DCIdentifier, DCTermsIdentifier, and ID.
func (e *Entry) Identifier() string {
	if e == nil {
		return ""
	}
	if e.DCIdentifier != "" {
		return e.DCIdentifier
	}
	if e.DCTermsIdentifier != "" {
		return e.DCTermsIdentifier
	}
	return e.ID
}

// Publisher returns the publisher name, checking DCPublisher and DCTermsPublisher.
func (e *Entry) Publisher() string {
	if e == nil {
		return ""
	}
	if e.DCPublisher != "" {
		return e.DCPublisher
	}
	return e.DCTermsPublisher
}

