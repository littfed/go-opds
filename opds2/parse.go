package opds2

import (
	"encoding/json"
	"fmt"
	"io"
)

// Parse parses an OPDS 2.0 catalog from JSON bytes.
func Parse(data []byte) (*Catalog, error) {
	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("opds2: parse catalog: %w", err)
	}
	return &catalog, nil
}

// DefaultMaxReadLimit is the maximum number of bytes ParseReader reads to prevent memory exhaustion (10 MB).
const DefaultMaxReadLimit = 10 << 20

// ParseReader parses an OPDS 2.0 catalog from an io.Reader, limited to DefaultMaxReadLimit bytes.
func ParseReader(r io.Reader) (*Catalog, error) {
	var catalog Catalog
	decoder := json.NewDecoder(io.LimitReader(r, DefaultMaxReadLimit))
	if err := decoder.Decode(&catalog); err != nil {
		return nil, fmt.Errorf("opds2: parse catalog: %w", err)
	}
	return &catalog, nil
}
