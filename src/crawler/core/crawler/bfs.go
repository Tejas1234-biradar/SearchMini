package crawler

import (
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/database"
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/pages"
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/utils"
	"log"
	"math"
)

func (c *CrawlerConfig) BFS(db *database.Database) {
	defer c.Wg.Done()
	for {
		log.Printf("Crawling....\n")
		if c.maxPagesReached() {
			log.Printf("MaxPagesReached\n")
			return
		}
		rawCurrentURL, depthLevel, normalizedCurrentURL, err := db.PopURL()
		if err != nil {
			log.Printf("Queue is Empty:%v", err)
			return
		}

		visited, err := db.HasURLBeenVisited(normalizedCurrentURL)
		if err != nil {
			log.Printf("Error: [%v]--skipping..\n", err)
			continue
		}
		if visited {
			continue
		}
		log.Printf("Crawling from %v..\n", normalizedCurrentURL)
		html, statusCode, contentType, err := getPageData(rawCurrentURL)
		if err != nil {
			log.Printf("Error getting links: %v", err)
			continue
		}
		outgoingLinks, err := getURLsFromHTML(html, rawCurrentURL)
		c.UpdateLinks(normalizedCurrentURL, outgoingLinks)
		log.Printf("Extracted links: %v\n", outgoingLinks)
		pg := pages.Constructor(normalizedCurrentURL, html, contentType, statusCode)
		err = c.addPage(pg)
		if err != nil {
			log.Print("Error adding pages\n")
			continue
		}
		err = db.VisitPage(normalizedCurrentURL)
		if err != nil {
			continue
		}
		//Add neighbours
		for _, rawCurrentLink := range outgoingLinks {
			if !utils.IsValidURL(rawCurrentLink) {
				continue
			}

			score, exists := db.ExistsInQueue(rawCurrentLink)
			if exists {
			} else {
				score = depthLevel + 1
			}
			score = math.Max(utils.MinScore, math.Min(score, utils.MaxScore))
			err := db.PushURL(rawCurrentLink, score)
			if err != nil {
				log.Printf("Push failed for %s: %v\n", rawCurrentLink, err)
			} else {
				log.Printf("Successfully pushed %s with score %v\n", rawCurrentLink, score)
			}
		}

	}

}
