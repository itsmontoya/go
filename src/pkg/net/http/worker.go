package http

import (
	"sync"

	//	"fmt"
	"github.com/missionMeteora/toolkit/errors"
)

func newWorker(queueLen int) *worker {
	var w worker
	w.cq = make(chan *conn, queueLen)
	go w.listen()
	return &w
}

type worker struct {
	wg sync.WaitGroup
	cq chan *conn

	closed bool
}

func (w *worker) listen() {
	w.wg.Add(1)

	for c := range w.cq {
		//fmt.Println("About to serve!")
		if c == nil {
			panic("hmm")
		}
		c.serve()
		//fmt.Println("Served!")
	}

	w.wg.Done()
}

func (w *worker) Close() (err error) {
	if w.closed {
		return errors.ErrIsClosed
	}

	close(w.cq)
	// Still deciding if waiting until complete is needed
	//	w.wg.Wait()
	return
}

func newWorkers(workerCnt, queueLen int) *workers {
	var w workers
	w.ws = make([]*worker, 0, workerCnt)

	for i := 0; i < workerCnt; i++ {
		w.ws = append(w.ws, newWorker(queueLen))
	}

	w.last = workerCnt - 1
	return &w
}

type workers struct {
	cur  int
	last int

	ws []*worker
}

func (w *workers) queue(c *conn) {
	wkr := w.ws[w.cur]
	wkr.cq <- c

	if w.cur++; w.cur > w.last {
		w.cur = 0
	}
}

func (w *workers) closeWorkers(n int) {
	var li int
	if li = len(w.ws) - 1; n > li {
		// Our count is greater than our total workers, set to cap
		n = li
	}

	for ; n > 0; n-- {
		// Close worker
		wkr := w.ws[li]
		wkr.Close()
		// Remove worker from slice
		w.ws = w.ws[:li]
		// Set new last index
		li--
	}
}
