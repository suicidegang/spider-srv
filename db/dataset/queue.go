package dataset

// A buffered channel that we can send work requests on.
var Prepares = make(chan PrepareJob, 1000)
var PreparesQueue chan chan PrepareJob

func PQueueDispatcher(workers int) {
	PreparesQueue = make(chan chan PrepareJob, workers)

	// Start n workers using brand new queue channel
	for n := 0; n < workers; n++ {
		w := NewPWorker(n+1, PreparesQueue)
		w.Start()
	}

	go func() {
		for {
			select {
			case work := <-Prepares:
				go func() {
					worker := <-PreparesQueue
					worker <- work
				}()
			}
		}
	}()
}
