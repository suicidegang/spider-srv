package dataset

import (
	"log"
)

type PrepareWorker struct {
	ID       int
	Work     chan PrepareJob
	Queue    chan chan PrepareJob
	QuitChan chan bool
}

func (worker *PrepareWorker) Start() {
	go func() {
		log.Printf("Starting pworker %v", worker.ID)

		for {
			// Add the worker into workers queue
			worker.Queue <- worker.Work

			select {
			case work := <-worker.Work:
				work.Work()

			case <-worker.QuitChan:
				log.Printf("pworker %v stopping", worker.ID)
			}
		}
	}()
}

func (worker *PrepareWorker) Stop() {
	go func() {
		worker.QuitChan <- true
	}()
}

func NewPWorker(id int, queue chan chan PrepareJob) PrepareWorker {
	w := PrepareWorker{
		ID:       id,
		Work:     make(chan PrepareJob),
		Queue:    queue,
		QuitChan: make(chan bool),
	}

	return w
}
