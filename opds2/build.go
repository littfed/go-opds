package opds2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// ToJSON serializes the Catalog to OPDS 2.0 JSON.
func (c *Catalog) ToJSON() ([]byte, error) {
	if c == nil {
		return nil, errors.New("opds2: catalog is nil")
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("opds2: marshal catalog: %w", err)
	}
	return data, nil
}

// ToJSONCompact serializes the Catalog to compact OPDS 2.0 JSON.
func (c *Catalog) ToJSONCompact() ([]byte, error) {
	if c == nil {
		return nil, errors.New("opds2: catalog is nil")
	}
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("opds2: marshal catalog compact: %w", err)
	}
	return data, nil
}

// WriteTo writes the Catalog as OPDS 2.0 JSON to the given writer.
func (c *Catalog) WriteTo(w io.Writer) (int64, error) {
	if c == nil {
		return 0, errors.New("opds2: catalog is nil")
	}
	data, err := c.ToJSON()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	return int64(n), err
}

// ToJSON serializes the Publication to OPDS 2.0 JSON.
func (p *Publication) ToJSON() ([]byte, error) {
	if p == nil {
		return nil, errors.New("opds2: publication is nil")
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("opds2: marshal publication: %w", err)
	}
	return data, nil
}

// ToJSONCompact serializes the Publication to compact OPDS 2.0 JSON.
func (p *Publication) ToJSONCompact() ([]byte, error) {
	if p == nil {
		return nil, errors.New("opds2: publication is nil")
	}
	data, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("opds2: marshal publication compact: %w", err)
	}
	return data, nil
}

// WriteTo writes the Publication as OPDS 2.0 JSON to the given writer.
func (p *Publication) WriteTo(w io.Writer) (int64, error) {
	if p == nil {
		return 0, errors.New("opds2: publication is nil")
	}
	data, err := p.ToJSON()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	return int64(n), err
}
