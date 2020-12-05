package pool

import "time"

type workerInterface interface {
	//Run worker
	Run()
}

type worker struct {
	init      func(w *worker)
	startedAt time.Time
	queuedAt  time.Time
}

func (w *worker) Run() {
	w.init(w)
}
