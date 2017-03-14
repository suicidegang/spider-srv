package sitemap

import (
	"github.com/orcaman/concurrent-map"

	"fmt"
)

// A buffered channel where we send initial work requests to...
var Queue = make(chan SitemapRequest, 20)

// Urls hash table for o(1) concurrent checks
var SitemapTable cmap.ConcurrentMap
var SitemapRetries cmap.ConcurrentMap

func Dispatcher() {
	SitemapTable = cmap.New()
	SitemapRetries = cmap.New()

	go func() {
		for {
			fmt.Printf("Dispatcher cycle ran.")

			select {
			case work := <-Queue:
				fmt.Printf("Received work request in generic queue. Spawning worker.")

				worker := NewWorker()
				worker.Start()
				worker.Queue <- work
			}
		}
	}()
}
