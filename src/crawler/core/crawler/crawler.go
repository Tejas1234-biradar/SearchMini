package crawler

import (
	"fmt"
	"sync"

	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/pages"
)
type CrawlerConfig struct{
	// only one goroutine can acess this
Mu *sync.Mutex 
//WaitGroup wait for a collection of goroutines to finish
Wg *sync.WaitGroup
Pages map[string]*pages.Page

MaxPages int
MaxConcurrency int
}
func (c *CrawlerConfig) lenPages()int{
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return len(c.Pages)
}
func (c* CrawlerConfig)maxPagesReached()(bool){
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if len(c.Pages)>=c.MaxPages{
		return true
	}
	return false
}


