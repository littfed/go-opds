package common_test

import (
	"testing"

	"github.com/littfed/go-opds/internal/common"
)

func TestDetectMIMEType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "epub", input: "book.epub", expected: common.MIMEEPUB},
		{name: "epub uppercase", input: "BOOK.EPUB", expected: common.MIMEEPUB},
		{name: "pdf", input: "document.pdf", expected: common.MIMEPDF},
		{name: "mobi", input: "kindle.mobi", expected: common.MIMEMOBI},
		{name: "azw3", input: "kindle.azw3", expected: common.MIMEAZW3},
		{name: "cbz", input: "comic.cbz", expected: common.MIMECBZ},
		{name: "fb2", input: "fiction.fb2", expected: common.MIMEFB2},
		{name: "html", input: "page.html", expected: common.MIMEHTML},
		{name: "htm", input: "page.htm", expected: common.MIMEHTML},
		{name: "png", input: "cover.png", expected: common.MIMEPNG},
		{name: "jpg", input: "cover.jpg", expected: common.MIMEJPEG},
		{name: "jpeg", input: "cover.jpeg", expected: common.MIMEJPEG},
		{name: "gif", input: "cover.gif", expected: common.MIMEGIF},
		{name: "unknown txt", input: "readme.txt", expected: ""},
		{name: "unknown extensionless", input: "manifest", expected: ""},
		{name: "empty string", input: "", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := common.DetectMIMEType(tt.input)
			if got != tt.expected {
				t.Errorf("DetectMIMEType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
