package crawler

import (
	"sync"
	"testing"

	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/database"
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/pages"
	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/utils"
)

func TestBFS_GoDev(t *testing.T) {

	// -------- COMMENTED OUT LOCAL TEST SERVER --------
	/*
		var server *httptest.Server

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			html := `
		<html>
			<body>
				<a href="` + server.URL + `/page1">Page1</a>
				<a href="` + server.URL + `/page2">Page2</a>
				<a href="` + server.URL + `/page3">Page3</a>
				<a href="` + server.URL + `/page4">Page4</a>
				<a href="` + server.URL + `/page5">Page5</a>
			</body>
		</html>`
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			w.Write([]byte(html))
		})

		server = httptest.NewServer(handler)
		defer server.Close()
	*/

	// -------- Setup Redis --------
	db := &database.Database{}
	err := db.ConnectToRedis("localhost", "6379", "", "0")
	if err != nil {
		t.Fatalf("Redis connection failed: %v", err)
	}

	db.Client.Del(db.Context, utils.CrawlerQueueKey)

	// -------- Use go.dev as seed --------
	seedURL := "https://www.delugerpg.com/"

	err = db.PushURL(seedURL, 0)
	if err != nil {
		t.Fatalf("PushURL failed: %v", err)
	}

	t.Logf("Pushed seed URL: %v", seedURL)

	// -------- Setup Crawler --------
	c := &CrawlerConfig{
		Mu:        &sync.Mutex{},
		Wg:        &sync.WaitGroup{},
		Pages:     make(map[string]*pages.Page),
		BackLinks: make(map[string]*pages.PageNode),
		Outlinks:  make(map[string]*pages.PageNode),
		MaxPages:  3, // keep small to avoid large crawl
	}

	c.Wg.Add(1)
	go c.BFS(db)
	c.Wg.Wait()

	// -------- Assertions --------
	if len(c.Pages) == 0 {
		t.Fatalf("Expected pages to be crawled, got 0")
	}

	if len(c.Outlinks) == 0 {
		t.Fatalf("Outlinks were not updated")
	}

	if len(c.BackLinks) == 0 {
		t.Fatalf("BackLinks were not updated")
	}

	t.Logf("Tachyon I dit it!!! integration test passed: crawled %d pages", len(c.Pages))
}
