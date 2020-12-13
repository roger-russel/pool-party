package pool

import (
	"time"
)

//WorkerInfo return success information
type WorkerInfo struct {
	ID            string
	QueuedAt      time.Time
	WaitingTime   time.Duration
	ExecutionTime time.Duration
	FinishedTime  time.Duration
	Err           error
	CalledToRun   bool
}
