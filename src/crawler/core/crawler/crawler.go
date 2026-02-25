package crawler

import (
	"fmt"
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/pages"
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/utils"
	"log"
	"sync"
)

type CrawlerConfig struct {
	// only one goroutine can acess this
	Mu *sync.Mutex
	// WaitGroup wait for a collection of goroutines to finish
	Wg        *sync.WaitGroup
	Pages     map[string]*pages.Page
	BackLinks map[string]*pages.PageNode
	Outlinks  map[string]*pages.PageNode

	MaxPages       int
	MaxConcurrency int
}

func (c *CrawlerConfig) lenPages() int {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return len(c.Pages)
}
func (c *CrawlerConfig) maxPagesReached() bool {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if len(c.Pages) >= c.MaxPages {
		return true
	}
	return false
}
func (c *CrawlerConfig) addPage(page *pages.Page) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	normalizedURL := page.NormalizedURL
	if _, visited := c.Pages[normalizedURL]; visited {
		return fmt.Errorf("Page already visited")
	}
	if len(c.Pages) >= c.MaxPages {
		return fmt.Errorf("Max Pages reached cannot add more")
	}
	c.Pages[normalizedURL] = page
	return nil
}
func (c *CrawlerConfig) UpdateLinks(normalizedCurrentURL string, outgoingLinks []string) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.Outlinks[normalizedCurrentURL] = pages.PageNodeConstructor(normalizedCurrentURL)

	for _, link := range outgoingLinks {
		if utils.IsValidURL(link) {
			normalizedOutgoingURL, err := utils.NormalizeURL(link)
			if err != nil {
				log.Printf("Normalization failed for link: %s\n", link)
				continue
			}

			if normalizedOutgoingURL == normalizedCurrentURL {
				continue
			}

			if _, exists := c.BackLinks[normalizedOutgoingURL]; !exists {
				c.BackLinks[normalizedOutgoingURL] = pages.PageNodeConstructor(normalizedOutgoingURL)
			}

			c.BackLinks[normalizedOutgoingURL].AddLink(normalizedCurrentURL)
			c.Outlinks[normalizedCurrentURL].AddLink(normalizedOutgoingURL)
		}
	}
}
