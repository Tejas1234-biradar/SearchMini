package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"sync"
)

type Result struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

func fetchURL(url string, wg *sync.WaitGroup, ch chan<- Result) {
	defer wg.Done() //executes when the function it is in returns a value
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching %s: %v", url, err)
		return

	}
	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {

		log.Printf("Error parsing url:", err)
		return
	}
	title := doc.Find("title").Text()
	ch <- Result{URL: url, Title: title}

}
func main() {
	urls := []string{"http://example.com","http://httpbin.org/html"}
	var wg sync.WaitGroup
	ch := make(chan Result, len(urls))

	// Spin up a goroutine for each URL
	for _, url := range urls {
		wg.Add(1)                 //dont end the main even after main function reaches end of program
		go fetchURL(url, &wg, ch) //go rouitine run this function in parallel with the main function
	}

	// Wait for completion and close channel
	wg.Wait()
	close(ch)

	// Collect and print results
	for result := range ch {
		fmt.Printf("URL: %s, Title: %s\n", result.URL, result.Title)

	}
}
