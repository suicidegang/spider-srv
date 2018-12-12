package sitemap

import (
	"fmt"
)

// Queue is a buffered channel where we receive work requests...
var Queue = make(chan Request, 20)

// Dispatcher worker
func Dispatcher() {
	go func() {
		for {
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
