package crawler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Returns Html,status code,content-type and Error Code
func getPageData(rawURL string) (string, int, string, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	// ---- Add realistic browser headers ----
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")
	req.Header.Set("Accept",
		"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode > 399 {
		return "", res.StatusCode, "", fmt.Errorf("HTTP error: %d %s",
			res.StatusCode, http.StatusText(res.StatusCode))
	}

	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return "", res.StatusCode, contentType,
			fmt.Errorf("invalid content type (only supports html): %s", contentType)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", res.StatusCode, "text/html",
			fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), res.StatusCode, contentType, nil
}
