package crawler

import (
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/utils"
	"golang.org/x/net/html"
	"net/url"
	"regexp"
	"strings"
)

var nonASCIIRegex = regexp.MustCompile(`[^\x20-\x7E]`)

// Get all the html links in a website
func getURLsFromHTML(htmlBody string, rawURL string) ([]string, error) {
	baseUrl, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	node, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}
	linksSet := make(map[string]struct{})
	traverse(node, baseUrl, linksSet)
	links := make([]string, 0, len(linksSet))
	for link := range linksSet {
		links = append(links, link)
	}
	return links, nil
}

// traverse in the html tags in a webpage take all the <a href="">  links
func traverse(node *html.Node, baseURL *url.URL, linksSet map[string]struct{}) {
	//base cases
	if node == nil {
		return
	}

	// type Attribute struct {
	// 	Namespace, Key, Val string
	// }
	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				rawHref := attr.Val
				if strings.ContainsAny(rawHref, "<>\"") || nonASCIIRegex.MatchString(rawHref) {
					continue
				}
				u, err := url.Parse(rawHref)
				if err != nil {
					continue
				}
				var resolved string

				// Absolute means that it has a non-empty scheme.
				if u.IsAbs() {
					resolved = u.String()
				} else {
					resolved = baseURL.ResolveReference(u).String()
				}
				resolved, err = utils.NormalizeURL(resolved)
				if err != nil {
					continue
				}
				linksSet[resolved] = struct{}{}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		traverse(c, baseURL, linksSet)
	}
}
