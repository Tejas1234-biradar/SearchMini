package crawler

import (
	"sync"
	"testing"

	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/pages"
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/utils"
)

func newTestCrawler(maxPages int) *CrawlerConfig {
	return &CrawlerConfig{
		Mu:             &sync.Mutex{},
		Wg:             &sync.WaitGroup{},
		Pages:          make(map[string]*pages.Page),
		BackLinks:      make(map[string]*pages.PageNode),
		Outlinks:       make(map[string]*pages.PageNode),
		MaxPages:       maxPages,
		MaxConcurrency: 5,
	}
}
func TestLenPages(t *testing.T) {
	c := newTestCrawler(10)

	if c.lenPages() != 0 {
		t.Errorf("expected 0 pages, got %d", c.lenPages())
	}

	page := &pages.Page{NormalizedURL: "https://example.com"}
	c.Pages[page.NormalizedURL] = page

	if c.lenPages() != 1 {
		t.Errorf("expected 1 page, got %d", c.lenPages())
	}
}
func TestMaxPagesReached(t *testing.T) {
	c := newTestCrawler(1)

	if c.maxPagesReached() {
		t.Errorf("should not have reached max pages")
	}

	page := &pages.Page{NormalizedURL: "https://example.com"}
	c.Pages[page.NormalizedURL] = page

	if !c.maxPagesReached() {
		t.Errorf("should have reached max pages")
	}
}
func TestAddPage(t *testing.T) {
	c := newTestCrawler(2)

	page1 := &pages.Page{NormalizedURL: "https://example.com"}
	page2 := &pages.Page{NormalizedURL: "https://example2.com"}

	// add first page
	if err := c.addPage(page1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// duplicate page
	if err := c.addPage(page1); err == nil {
		t.Errorf("expected error for duplicate page")
	}

	// add second page
	if err := c.addPage(page2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// exceed max pages
	page3 := &pages.Page{NormalizedURL: "https://example3.com"}
	if err := c.addPage(page3); err == nil {
		t.Errorf("expected max pages error")
	}
}
func TestUpdateLinks(t *testing.T) {
	c := newTestCrawler(10)

	currentURL := "https://example.com"

	outgoing := []string{
		"https://google.com",
		"https://github.com",
		"https://example.com", // self link (should be ignored)
	}

	c.UpdateLinks(currentURL, outgoing)

	// Normalize current URL (since graph stores normalized keys)
	normalizedCurrentURL, err := utils.NormalizeURL(currentURL)
	if err != nil {
		t.Fatalf("unexpected normalization error: %v", err)
	}

	// Outlinks should exist
	node, exists := c.Outlinks[normalizedCurrentURL]
	if !exists {
		t.Fatalf("expected outlinks entry for current URL")
	}

	// Should have 2 valid outgoing links (self excluded)
	if len(node.NormalizedURLs) != 2 {
		t.Errorf("expected 2 outgoing links, got %d", len(node.NormalizedURLs))
	}

	// Backlinks map should contain google + github
	if len(c.BackLinks) != 2 {
		t.Errorf("expected 2 backlinks entries, got %d", len(c.BackLinks))
	}

	// Each backlink should contain normalized current URL
	for _, backlinkNode := range c.BackLinks {
		if _, exists := backlinkNode.NormalizedURLs[normalizedCurrentURL]; !exists {
			t.Errorf("expected backlink to contain %s", normalizedCurrentURL)
		}
	}
}
