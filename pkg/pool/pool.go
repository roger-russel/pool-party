package pool

import (
	"sync"
	"time"

	"github.com/rs/xid"
)

//New Pool Server
func New(size int) Pool {
	return &Impl{
		size:       size,
		chDone:     make(chan WorkerInfo),
		chInfo:     make(chan WorkerInfo),
		chQuit:     make(chan bool),
		chShutdown: make(chan bool),
		queue:      &queueImpl{},
	}
}

//Pool to create a pool
type Pool interface {

	// Server will be a pool server wich will wait until a shootdown as send
	Server()

	// Add worker to pool
	Add(f func() error) (id string, err error)

	// SetMaxPoolSize pool max size to a different number
	SetMaxPoolSize(s int)

	// Show status from pool
	// Status() string

	// Shutdown start shutdown pool server
	Shutdown() (killedTasksIds []string)
}

//Impl is the Pool interface implementation
type Impl struct {

	//Pool size
	size int

	// started is true when the pool is already calling the workers ( Run() or Server() )
	started bool

	// Time of the run was called
	startedAt time.Time

	//shuttingDown is true when this pool is asked to shutdown
	shuttingDown bool

	//Quit is the last channel to be called, it will lead to end of server.
	chQuit chan bool

	//Done channel called when an workerd finish
	chDone chan WorkerInfo

	//chInfo will send realtime info about what is happening
	chInfo chan WorkerInfo

	//Shutdown channel called when want to shutdown this pool it will send a shutdown fignal for ,
	chShutdown chan bool

	//muQueue is to add and remove thread safe from queued slice
	muQueue sync.Mutex

	//queued workers slice
	queued []worker

	//queue
	queue queue

	//number of worker running
	runningWorkers int
}

//Server start pool as service
func (p *Impl) Server() {

	p.started = true
	p.startedAt = time.Now()

	go p.done()

	for p.callNextWorker() {
		// call until false
	}

	<-p.chQuit

	close(p.chQuit)
	close(p.chInfo)
	close(p.chDone)
	close(p.chShutdown)

}

//Add new workers to pool
func (p *Impl) Add(f func() error) (id string, err error) {

	if p.shuttingDown {
		return "", errAddONShuttingDown
	}

	id = xid.New().String()

	p.muQueue.Lock()

	p.queued = append(p.queued, &workerImpl{
		id:       id,
		queuedAt: time.Now(),
		run: func() error {
			return f()
		},
	})

	p.muQueue.Unlock()

	p.callNextWorker()

	return id, nil
}

/*SetMaxPoolSize pool max size to a different number
 * It should change the pull size and start more workers
 * if it the new pool size is greater then the previous one
 * and if there is more workers to start.
 * if the new pull size is lower, it should do nothing,
 * because when a current task finish it will check if
 * it should call a new worker or not.
 */
func (p *Impl) SetMaxPoolSize(size int) {

	current := p.size
	p.size = size

	if current >= size {
		return
	}

	for p.callNextWorker() {
		// call until false
	}

}

//Call Next Worker to be executed
func (p *Impl) callNextWorker() bool {

	if p.shuttingDown || p.size <= p.runningWorkers || p.queue.len() <= 0 {
		return false
	}

	w := p.queue.get(&p.runningWorkers)

	if w != nil {
		go w.Start(p.chDone)
	}

	return len(p.queued) > 0

}

/*Shutdown force a shutdown on application, it will wait only for
 *the running tasks finish, for a Graceful shutdown take a look into
 *which will try to run all tasks on pool which are been queued ShutdownGraceful
 */
func (p *Impl) Shutdown() (killedTasksIds []string) {

	p.shuttingDown = true
	p.chShutdown <- true

	p.muQueue.Lock()
	defer p.muQueue.Unlock()

	for _, w := range p.queued {
		killedTasksIds = append(killedTasksIds, w.ID())
	}

	return killedTasksIds

}

func (p *Impl) done() {

	for wInfo := range p.chDone {

		p.runningWorkers--

		p.chInfo <- wInfo

		if p.shuttingDown && p.runningWorkers == 0 {
			p.chQuit <- true
			break
		}

		p.callNextWorker()

	}
}
