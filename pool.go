package party

import (
	"fmt"
	"sync"
)

//Pool worker
type Pool struct {
	size    int
	workers map[int]func()
	running map[int]bool
	wg      sync.WaitGroup
	chDone  chan int
	tasks   int //The number of current tasks on pool
}

//PoolInterface to create a pool
type PoolInterface interface {

	// Add worker to pool
	Add(f func())

	// Run tasks keept into pool
	Run()

	// SetMaxPoolSize pool max size to a different number
	SetMaxPoolSize(s int)

	// Show status from pool
	Status() string

	// done end a task on pool
	done()
}

//New Pool
func New(size int) PoolInterface {
	return &Pool{
		size:    size,
		workers: make(map[int]func()),
		running: make(map[int]bool),
		chDone:  make(chan int, 1),
	}
}

//Run Workers on Pool
func (p *Pool) Run() {

	//ToDo Add here time when run started

	go p.done()

	ln := len(p.workers)

	for workerID := 0; workerID < p.size && workerID < ln; workerID++ { // lock running go rotines
		p.running[workerID] = true
	}

	for workerID := 0; workerID < p.size && workerID < ln; workerID++ { // run it
		go p.workers[workerID]()
	}

	p.wg.Wait()

}

//Add Function into Pool
func (p *Pool) Add(f func()) {

	p.tasks++
	p.wg.Add(1)
	workerID := len(p.workers)

	//Todo add here time when it was queued
	p.workers[workerID] = func() {
		//ToDO add here time when it started up
		f()
		p.chDone <- workerID
	}

	//Todo add here when a new task was added to pool, but there is non task running anymore
	//It should know that it must run imediatly when there is less than the pool max size running

}

// SetMaxPoolSize pool max size to a different number
func (p *Pool) SetMaxPoolSize(s int) {
	/*
	   ToDO: set max pool size, must change the pool size size.
	   It should check if the new max number is grater then the currenctly
	   If it was it should call more workers to start by itself
	   If it was lesser then the done funcion should check every task done
	   if call a new one would be greater then the max size
	*/
}

func (p *Pool) done() {

	for workerID := range p.chDone {
		delete(p.workers, workerID)
		delete(p.running, workerID)

		p.tasks--
		p.wg.Done()

		if len(p.workers) > 0 {
			for next := range p.workers {
				if _, running := p.running[next]; !running {
					p.running[next] = true
					go p.workers[next]()
					break
				}
			}
		}
	}
}

//Status of pool
func (p *Pool) Status() string {
	return fmt.Sprintf("Remaining: %v",
		p.tasks,
	)
}
