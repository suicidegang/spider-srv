package sitemap

import (
	"github.com/orcaman/concurrent-map"

	"fmt"
)

// A buffered channel that we can send work requests on.
var Queue = make(chan SitemapRequest, 100000)

// Queue channel
var SitemapQueue chan chan SitemapRequest

// Urls hash table for o(1) concurrent checks
var SitemapTable cmap.ConcurrentMap
var SitemapRetries cmap.ConcurrentMap

func Dispatcher(workers int) {
	SitemapQueue = make(chan chan SitemapRequest, workers)
	SitemapTable = cmap.New()
	SitemapRetries = cmap.New()

	// Start n workers using brand new queue channel
	for n := 0; n < workers; n++ {
		w := NewWorker(n+1, SitemapQueue)
		w.Start()
	}

	go func() {
		for {
			select {
			case work := <-Queue:
				fmt.Printf("+")

				go func() {
					worker := <-SitemapQueue
					fmt.Printf("-")

					worker <- work
				}()
			}
		}
	}()
}
