package sitemap

import (
	"log"

	"github.com/satori/go.uuid"
)

type Worker struct {
	ID       string
	Queue    chan SitemapRequest
	QuitChan chan bool
}

func (worker *Worker) Start() {
	go func() {
		log.Printf("Spawning worker id %s", worker.ID)

		for {
			select {
			case work := <-worker.Queue:
				log.Printf("Crawled %s : depth %v", work.Url, work.Depth)
				work.Work(worker.Enqueue)

			case <-worker.QuitChan:
				log.Printf("worker %v stopping", worker.ID)
			}
		}
	}()
}

func (worker *Worker) Enqueue(job SitemapRequest) {
	go func() {
		worker.Queue <- job
	}()
}

func (worker *Worker) Stop() {
	go func() {
		worker.QuitChan <- true
	}()
}

func NewWorker() Worker {
	w := Worker{
		ID:       uuid.NewV4().String(),
		Queue:    make(chan SitemapRequest, 500),
		QuitChan: make(chan bool),
	}

	return w
}
