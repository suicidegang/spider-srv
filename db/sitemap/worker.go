package sitemap

import (
	"log"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/satori/go.uuid"
)

// Worker def.
type Worker struct {
	ID       string
	Queue    chan Request
	QuitChan chan bool
}

// Start worker queue processing.
func (worker *Worker) Start() {
	go func() {
		log.Printf("Spawning worker id %s", worker.ID)
		var lastRequest *Request
		done := 0

		// Concurrent maps that will live as long as the worker still processing...
		visited := new(sync.Map)
		retries := new(sync.Map)
		for {
			select {
			case work := <-worker.Queue:
				log.Printf("Crawled %s : depth %v", work.URL, work.Depth)

				// Perform work
				work.Done = visited
				work.Retries = retries
				work.Work(worker.Enqueue)
				lastRequest = &work
				lastRequest.DB.Model(Sitemap{}).Where("id = ?", lastRequest.SitemapID).UpdateColumn("done", gorm.Expr("done + ?", 1))
				done++

			case <-worker.QuitChan:
				log.Printf("worker %v stopping", worker.ID)

			case <-time.After(4 * time.Second):
				log.Println("Stopping worker [pid:", worker.ID, "]")
				log.Println("Processed", done, "jobs")

				// Mark sitemap as done.
				if done > 0 && lastRequest != nil {
					lastRequest.DB.Model(Sitemap{}).Where("id = ?", lastRequest.SitemapID).Update("updating", false)
				}
				return
			}
		}
	}()
}

// Enqueue a job on the same worker.
func (worker *Worker) Enqueue(job Request) {
	// It the case the internal buffer gets overfilled,
	// the additional goroutine wrapping the job sending
	// will prevent a hard blocking condition.
	go func() {
		worker.Queue <- job
	}()
}

// Stop the current worker if running...
func (worker *Worker) Stop() {
	go func() {
		worker.QuitChan <- true
	}()
}

func NewWorker() Worker {
	w := Worker{
		ID:       uuid.Must(uuid.NewV4()).String(),
		Queue:    make(chan Request, 500),
		QuitChan: make(chan bool),
	}

	return w
}
