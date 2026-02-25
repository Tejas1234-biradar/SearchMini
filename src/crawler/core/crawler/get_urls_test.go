package crawler

import (
	"fmt"
	"testing"
)

func TestGetURLsFromHTML_RealURL(t *testing.T) {
	url := "https://youtube.com" // Replace with any live URL if you want

	fmt.Println("=== Starting test for URL:", url, "===")

	// Fetch page content using your getPageData function
	htmlBody, statusCode, contentType, err := getPageData(url)
	if err != nil {
		t.Fatalf("Failed to fetch page: %v", err)
	}

	fmt.Println("Status Code:", statusCode)
	fmt.Println("Content-Type:", contentType)

	// Extract links
	links, err := getURLsFromHTML(htmlBody, url)
	if err != nil {
		t.Fatalf("Failed to extract links: %v", err)
	}

	fmt.Println("=== Extracted & Normalized Links ===")
	for i, link := range links {
		fmt.Printf("%d: %s\n", i+1, link)
	}

	fmt.Println("=== Test finished ===")
}
