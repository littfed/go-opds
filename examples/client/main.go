package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/littfed/go-opds/opds1"
)

// Client is an OPDS catalog client.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new OPDS client.
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}
}

func (c *Client) resolveURL(ref string) (string, error) {
	if ref == "" {
		return c.baseURL, nil
	}
	u, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("invalid reference URL %q: %w", ref, err)
	}
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL %q: %w", c.baseURL, err)
	}
	resolved := base.ResolveReference(u)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme %q in resolved URL", resolved.Scheme)
	}
	return resolved.String(), nil
}

const defaultUserAgent = "go-opds-client/1.0 (+https://github.com/littfed/go-opds)"

func (c *Client) get(targetURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "application/atom+xml,application/xml,text/xml,*/*")
	return c.httpClient.Do(req)
}

// FetchFeed fetches and parses an OPDS 1.2 feed from the given path.
func (c *Client) FetchFeed(path string) (*opds1.Feed, error) {
	targetURL, err := c.resolveURL(path)
	if err != nil {
		return nil, err
	}
	resp, err := c.get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("fetch feed %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch feed %s: status %d", targetURL, resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("read feed body: %w", err)
	}

	return opds1.Parse(data)
}

// FetchEntry fetches and parses an OPDS 1.2 entry from the given path.
func (c *Client) FetchEntry(path string) (*opds1.Entry, error) {
	targetURL, err := c.resolveURL(path)
	if err != nil {
		return nil, err
	}
	resp, err := c.get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("fetch entry %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch entry %s: status %d", targetURL, resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("read entry body: %w", err)
	}

	return opds1.ParseEntry(data)
}

// DownloadFile downloads a file from the given path and saves it to localPath.
func (c *Client) DownloadFile(remotePath, localPath string) error {
	targetURL, err := c.resolveURL(remotePath)
	if err != nil {
		return err
	}
	resp, err := c.get(targetURL)
	if err != nil {
		return fmt.Errorf("download %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: status %d", targetURL, resp.StatusCode)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", localPath, err)
	}
	defer out.Close()

	written, err := io.Copy(out, io.LimitReader(resp.Body, 500<<20))
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("close file %s: %w", localPath, err)
	}

	fmt.Printf("Downloaded %s (%d bytes)\n", localPath, written)
	return nil
}

// AcquisitionLinks returns all acquisition links from an entry.
func AcquisitionLinks(entry *opds1.Entry) []opds1.Link {
	var links []opds1.Link
	for _, link := range entry.Links {
		if strings.HasPrefix(link.Rel, "http://opds-spec.org/acquisition") {
			links = append(links, link)
		}
	}
	return links
}

// ThumbnailLink returns the thumbnail link from an entry, if any.
func ThumbnailLink(entry *opds1.Entry) *opds1.Link {
	for i := range entry.Links {
		if entry.Links[i].Rel == opds1.RelImageThumbnail {
			return &entry.Links[i]
		}
	}
	return nil
}

// ImageLink returns the full image link from an entry, if any.
func ImageLink(entry *opds1.Entry) *opds1.Link {
	for i := range entry.Links {
		if entry.Links[i].Rel == opds1.RelImage {
			return &entry.Links[i]
		}
	}
	return nil
}

// PrintFeed prints a summary of the feed contents.
func PrintFeed(feed *opds1.Feed) {
	fmt.Printf("=== %s ===\n", feed.Title)
	fmt.Printf("ID:      %s\n", feed.ID)
	fmt.Printf("Updated: %s\n", feed.Updated)
	fmt.Println()

	// Print navigation links
	for _, link := range feed.Links {
		if link.Rel == opds1.RelSelf || link.Rel == opds1.RelStart {
			fmt.Printf("[%s] %s\n", link.Rel, link.Href)
		}
	}
	fmt.Println()

	// Print entries
	for i, entry := range feed.Entries {
		fmt.Printf("%d. %s\n", i+1, entry.Title)
		fmt.Printf("   ID: %s\n", entry.ID)

		// Check if this is navigation or acquisition
		hasAcquisition := false
		for _, link := range entry.Links {
			if strings.HasPrefix(link.Rel, "http://opds-spec.org/acquisition") {
				hasAcquisition = true
				fmt.Printf("   Format: %s -> %s\n", link.Type, link.Href)
			}
		}

		if !hasAcquisition {
			// Navigation entry
			for _, link := range entry.Links {
				if link.Rel == opds1.RelSubsection || link.Rel == opds1.RelAcquisition {
					fmt.Printf("   Navigate: %s (%s)\n", link.Href, link.Type)
				}
			}
		}

		if entry.Author != nil {
			fmt.Printf("   Author: %s\n", entry.Author.Name)
		}
		if entry.Summary != nil {
			summary := entry.Summary.Body
			if len(summary) > 80 {
				summary = summary[:min(len(summary), 77)] + "..."
			}
			fmt.Printf("   Summary: %s\n", summary)
		}
		fmt.Println()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <opds-url>")
		fmt.Println("Example: client http://localhost:8080")
		fmt.Println("Example: client http://www.feedbooks.com/store/recent.atom")
		os.Exit(1)
	}

	url := os.Args[1]
	client := NewClient(url)

	fmt.Printf("Connecting to OPDS catalog: %s\n\n", url)

	// Fetch the root feed
	feed, err := client.FetchFeed("")
	if err != nil {
		slog.Error("failed to fetch root feed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	PrintFeed(feed)

	// If there are navigation entries, browse into the first one
	for _, entry := range feed.Entries {
		for _, link := range entry.Links {
			if link.Rel == opds1.RelSubsection || strings.HasPrefix(link.Rel, "http://opds-spec.org/sort") {
				fmt.Printf("--- Browsing: %s ---\n\n", entry.Title)
				subFeed, err := client.FetchFeed(link.Href)
				if err != nil {
					slog.Warn("failed to fetch subfeed", slog.String("href", link.Href), slog.String("error", err.Error()))
					continue
				}
				PrintFeed(subFeed)

				// Show acquisition links for first few books
				limit := min(3, len(subFeed.Entries))
				for _, book := range subFeed.Entries[:limit] {
					acqLinks := AcquisitionLinks(&book)
					if len(acqLinks) > 0 {
						fmt.Printf("Downloadable formats for '%s':\n", book.Title)
						for _, al := range acqLinks {
							fmt.Printf("  - %s: %s\n", al.Type, al.Href)
						}
					}
				}
				if len(subFeed.Entries) > 3 {
					fmt.Printf("... and %d more entries\n", len(subFeed.Entries)-3)
				}
				fmt.Println()
				return
			}
		}
	}

	// If no navigation, just print acquisition entries
	if len(feed.Entries) > 0 {
		fmt.Println("This catalog contains acquisition entries (books).")
		fmt.Println("Available formats:")
		for _, entry := range feed.Entries {
			acqLinks := AcquisitionLinks(&entry)
			if len(acqLinks) > 0 {
				fmt.Printf("\n'%s' by %s:\n", entry.Title, entry.Author.Name)
				for _, al := range acqLinks {
					fmt.Printf("  - %s (%s)\n", al.Type, al.Href)
				}
			}
		}
	}
}
