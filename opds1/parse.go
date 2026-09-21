package opds1

import (
	"encoding/xml"
	"fmt"
	"io"
)

// Parse parses an OPDS 1.2 feed from XML bytes.
func Parse(data []byte) (*Feed, error) {
	var feed Feed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("opds1: parse feed: %w", err)
	}
	return &feed, nil
}

// DefaultMaxReadLimit is the maximum number of bytes ParseReader reads to prevent memory exhaustion (10 MB).
const DefaultMaxReadLimit = 10 << 20

// ParseReader parses an OPDS 1.2 feed from an io.Reader, limited to DefaultMaxReadLimit bytes.
func ParseReader(r io.Reader) (*Feed, error) {
	decoder := xml.NewDecoder(io.LimitReader(r, DefaultMaxReadLimit))
	var feed Feed
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("opds1: parse feed: %w", err)
	}
	return &feed, nil
}

// ParseEntry parses an OPDS 1.2 entry from XML bytes.
func ParseEntry(data []byte) (*Entry, error) {
	var entry Entry
	if err := xml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("opds1: parse entry: %w", err)
	}
	return &entry, nil
}
