package crawler

import (
	"fmt"
	"sync"
)
type CrawlerConfig struct{
	// only one goroutine can acess this
Mu *sync.Mutex 
//WaitGroup wait for a collection of goroutines to finish
Wg *sync.WaitGroup
MaxPages int
MaxConcurrency int
}


