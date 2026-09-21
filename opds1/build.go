package opds1

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// ToXML serializes the Feed to OPDS 1.2 XML.
func (f *Feed) ToXML() ([]byte, error) {
	if f == nil {
		return nil, errors.New("opds1: feed is nil")
	}
	output, err := xml.MarshalIndent(f, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("opds1: marshal feed: %w", err)
	}
	return append([]byte(xml.Header), output...), nil
}

// WriteTo writes the Feed as OPDS 1.2 XML to the given writer.
func (f *Feed) WriteTo(w io.Writer) (int64, error) {
	data, err := f.ToXML()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	return int64(n), err
}

// ToXML serializes the Entry to OPDS 1.2 XML.
func (e *Entry) ToXML() ([]byte, error) {
	if e == nil {
		return nil, errors.New("opds1: entry is nil")
	}
	output, err := xml.MarshalIndent(e, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("opds1: marshal entry: %w", err)
	}
	return append([]byte(xml.Header), output...), nil
}

// WriteTo writes the Entry as OPDS 1.2 XML to the given writer.
func (e *Entry) WriteTo(w io.Writer) (int64, error) {
	data, err := e.ToXML()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	return int64(n), err
}
