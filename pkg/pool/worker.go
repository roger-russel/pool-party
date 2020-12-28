package pool

import (
	"time"
)

type worker interface {
	Start()
	ID() string
	Info(CalledToRun bool, err error) *WorkerInfo
}

//workerImpl Struct
type workerImpl struct {
	run       func() error
	id        string
	queuedAt  time.Time
	startedAt time.Time
	chDone    chan *WorkerInfo
}

func (w *workerImpl) Start() {

	w.startedAt = time.Now()
	err := w.run()

	w.chDone <- w.Info(true, err)

}

func (w *workerImpl) ID() string {
	return w.id
}

func (w *workerImpl) Info(CalledToRun bool, err error) *WorkerInfo {
	return &WorkerInfo{
		ID:            w.id,
		QueuedAt:      w.queuedAt,
		WaitingTime:   w.startedAt.Sub(w.queuedAt),
		ExecutionTime: time.Since(w.startedAt),
		CalledToRun:   CalledToRun,
		Err:           err,
	}
}
