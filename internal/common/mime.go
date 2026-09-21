package common

import "strings"

// Publication format MIME types.
const (
	MIMEEPUB = "application/epub+zip"
	MIMEPDF  = "application/pdf"
	MIMEMOBI = "application/x-mobipocket-ebook"
	MIMEAZW3 = "application/vnd.amazon.ebook"
	MIMECBZ  = "application/x-cbz"
	MIMEFB2  = "application/x-fictionbook+xml"
	MIMEHTML = "text/html"
	MIMEPNG  = "image/png"
	MIMEJPEG = "image/jpeg"
	MIMEGIF  = "image/gif"
)

// DetectMIMEType returns the MIME type corresponding to a file extension in href.
func DetectMIMEType(href string) string {
	lower := strings.ToLower(href)
	switch {
	case strings.HasSuffix(lower, ".epub"):
		return MIMEEPUB
	case strings.HasSuffix(lower, ".pdf"):
		return MIMEPDF
	case strings.HasSuffix(lower, ".mobi"):
		return MIMEMOBI
	case strings.HasSuffix(lower, ".azw3"):
		return MIMEAZW3
	case strings.HasSuffix(lower, ".cbz"):
		return MIMECBZ
	case strings.HasSuffix(lower, ".fb2"):
		return MIMEFB2
	case strings.HasSuffix(lower, ".html") || strings.HasSuffix(lower, ".htm"):
		return MIMEHTML
	case strings.HasSuffix(lower, ".png"):
		return MIMEPNG
	case strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg"):
		return MIMEJPEG
	case strings.HasSuffix(lower, ".gif"):
		return MIMEGIF
	default:
		return ""
	}
}
