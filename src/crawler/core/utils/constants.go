package utils
import (
	"time"
)
const(
	Timeout = 5*time.Second
	MaxScore = 10000
	MinScore = -1000
	//Message Queues
	CrawlerQueueKey = "crawler_queue"
	IndexerQueueKey ="pages_queue"
	SignalQueueKey = "signal_queue"
	ResumeCrawl = "RESUME_CRAWL"
	MaxIndexerQueueSize = 5000
	//redis data
  NormalizedURLPrefix = "normalized_url"	// Stays in Redis indefinitely
	PagePrefix          = "page_data"		// Transferred by the indexer
	ImagePrefix         = "image_data"		// Transferred by the image indexer
	PageImagesPrefix    = "page_images"		// Transferred by the image indexer
	BacklinksPrefix		= "backlinks"		// Transferred by the backlinks processor
	OutlinksPrefix 		= "outlinks"		// Transferred by the indexer
)
