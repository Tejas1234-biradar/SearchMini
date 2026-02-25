package pages

import (
	"fmt"
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/utils"
	"time"
)

type Page struct {
	NormalizedURL string
	HTML          string
	ContentType   string
	LastCrawlTime time.Time
	StatusCode    int
}

func Constructor(normalizedUrl, html, contentype string, statuscode int) *Page {
	return &Page{
		NormalizedURL: normalizedUrl,
		HTML:          html,
		ContentType:   contentype,
		StatusCode:    statuscode,
		LastCrawlTime: time.Now(),
	}
}

func (p Page) ToString() string {
	htmlPreview := p.HTML

	// Truncate the HTML output
	if len(htmlPreview) > 15 {
		htmlPreview = htmlPreview[:15] + "..."
	}

	return fmt.Sprintf(
		"-------------------------------------------------\n"+
			"Normalized URL:    %-10s\n"+
			"HTML:              %-40s\n"+
			"Last Crawled:      %-30s\n"+
			"Status Code:       %-10d\n"+
			"Content Type:      %-20s\n"+
			"-------------------------------------------------\n",
		p.NormalizedURL, htmlPreview, p.LastCrawlTime.Format(time.RFC1123),
		p.StatusCode, p.ContentType,
	)
}
func HashPage(page *Page) (map[string]interface{}, error) {
	// Convert it to a redis hash
	return map[string]interface{}{
		"normalized_url": page.NormalizedURL,
		"html":           page.HTML,
		"content_type":   page.ContentType,
		"status_code":    page.StatusCode,
		"last_crawled":   page.LastCrawlTime.Format(time.RFC1123),
	}, nil
}

func DehashPage(data map[string]string) (*Page, error) {

	lastCrawled, err := utils.ParseTime(data["last_crawled"])
	if err != nil {
		return nil, fmt.Errorf("Error parsing 'LastCrawled' in hash: %w", err)
	}

	statusCode, err := utils.ParseInt(data["status_code"])
	if err != nil {
		return nil, fmt.Errorf("Error parsing 'StatusCode' in hash: %w", err)
	}

	return &Page{
		NormalizedURL: data["normalized_url"],
		HTML:          data["html"],
		ContentType:   data["content_type"],
		StatusCode:    statusCode,
		LastCrawlTime: lastCrawled,
	}, nil
}
