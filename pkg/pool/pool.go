package pool

import (
	"fmt"
	"sync"
	"time"
)

//Pool worker
type Pool struct {

	//Pool size
	size int

	//Running as server
	server bool

	//workers queued on this pool
	workers map[int]workerInterface

	//current workers running
	running map[int]bool

	//wait group for Run command
	wg sync.WaitGroup

	//Done channel called when an workerd finish
	chDone chan int

	//Shutdown channel called when want to shutdown this pool it will send a shutdown fignal for ,
	chShutdown chan bool

	//Quit is the last channel to be called, it will lead to end of running programimg.
	chQuit chan bool

	//The number of current tasks on pool
	runningNumber int

	//The number of queued tasks
	workersNumber int

	// Time of the run was called
	startedAt time.Time

	// started is true when the pool is already calling the workers ( Run() or Server() )
	started bool

	//shuttingDown is true when this pool is asked to shutdown
	shuttingDown bool
}

//IPool to create a pool
type IPool interface {

	// Add worker to pool
	Add(f func(chShutdown chan bool))

	// Run tasks keept into pool until there is no more queued
	Run()

	// Server will be a pool server wich will wait until a shootdown as send
	Server()

	// SetMaxPoolSize pool max size to a different number
	SetMaxPoolSize(s int)

	// Show status from pool
	Status() string

	// Shutdown()
	Shutdown()

	// done end a task on pool
	done()

	// start worker
	startNextWorker()
}

//New Pool
func New(size int) IPool {
	return &Pool{
		size:    size,
		workers: make(map[int]workerInterface),
		running: make(map[int]bool),
		chDone:  make(chan int),
	}
}

//Server will pool wait for shutdown
func (p *Pool) Server() {
	p.server = true
	p.Run()
	<-p.chQuit
}

//Run Workers on Pool
func (p *Pool) Run() {

	p.startedAt = time.Now()
	p.started = true

	go p.done()

	ln := len(p.workers)

	for workerID := 0; workerID < p.size && workerID < ln; workerID++ { // lock running go rotines
		p.running[workerID] = true
	}

	for workerID := 0; workerID < p.size && workerID < ln; workerID++ {
		go p.workers[workerID].Run()
	}

	// server will not wait for current pool finished
	if !p.server {
		p.wg.Wait()
	}

}

//Add Function into Pool
func (p *Pool) Add(f func(chShutdown chan bool)) {

	if p.shuttingDown {
		return
	}

	p.workersNumber++

	if !p.server {
		p.wg.Add(1)
	}

	workerID := len(p.workers)

	p.workers[workerID] = &worker{
		init: func(w *worker) {
			p.runningNumber++
			p.workersNumber--
			w.startedAt = time.Now()
			f(p.chShutdown)
			p.chDone <- workerID
		},
		queuedAt: time.Now(),
	}

	p.startNextWorker()

}

/*SetMaxPoolSize pool max size to a different number
 * It should change the pull size and start more workers
 * if it the new pool size is greater then the previous one
 * and if there is more workers to start.
 * if the new pull size is lower, it should do nothing,
 * because when a current task finish it will check if
 * it should call a new worker or not.
 */
func (p *Pool) SetMaxPoolSize(s int) {

	p.size = s

	for p.runningNumber < p.size && p.workersNumber > 0 {
		p.startNextWorker()
	}

}

//Shutdown make the pool shutdown gracefully
func (p *Pool) Shutdown() {
	p.shuttingDown = true
}

/* startNextWorker must start a new worker from queeue
 */
func (p *Pool) startNextWorker() {

	if p.started && !p.shuttingDown && p.runningNumber < p.size && p.workersNumber > 0 {
		for next := range p.workers {
			if _, running := p.running[next]; !running {
				p.running[next] = true
				go p.workers[next].Run()
				break
			}
		}
	}

}

func (p *Pool) done() {

	for workerID := range p.chDone {

		delete(p.workers, workerID)
		delete(p.running, workerID)

		p.runningNumber--

		if !p.server {
			p.wg.Done()
		}

		if p.shuttingDown && p.runningNumber == 0 {
			p.chQuit <- true
		}

		p.startNextWorker()

	}
}

//Status of pool
func (p *Pool) Status() string {
	return fmt.Sprintf("Remaining: %d",
		p.workersNumber,
	)
}
