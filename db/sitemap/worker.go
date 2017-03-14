package sitemap

import (
	"log"
)

type Worker struct {
	ID       int
	Work     chan SitemapRequest
	Queue    chan chan SitemapRequest
	QuitChan chan bool
}

func (worker *Worker) Start() {
	go func() {
		log.Printf("Starting worker %v", worker.ID)

		for {
			log.Printf("Worker %b ready again :D", worker.ID)

			// Add the worker into workers queue
			worker.Queue <- worker.Work

			select {
			case work := <-worker.Work:
				log.Printf("Crawled %s : depth %v", work.Url, work.Depth)
				work.Work()

			case <-worker.QuitChan:
				log.Printf("worker %v stopping", worker.ID)
			}
		}
	}()
}

func (worker *Worker) Stop() {
	go func() {
		worker.QuitChan <- true
	}()
}

func NewWorker(id int, queue chan chan SitemapRequest) Worker {
	w := Worker{
		ID:       id,
		Work:     make(chan SitemapRequest),
		Queue:    queue,
		QuitChan: make(chan bool),
	}

	return w
}
