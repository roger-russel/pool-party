package pool

import (
	"testing"
	"time"
)

func TestPool_NewPool(t *testing.T) {

	t.Run("Initializing a pool of size 1 with New Command", func(t *testing.T) {
		p := New(1)

		ch := make(chan struct{})

		xid, err := p.Add(func(taskID string) error {
			ch <- struct{}{}
			return nil
		})

		if xid == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err)
		}

		go p.Server()

		<-ch

	})

}

func TestPool_GettingDoneInfo(t *testing.T) {

	t.Run("A simple information", func(t *testing.T) {
		p := New(1)

		xid, err := p.Add(func(taskID string) error {
			return nil
		})

		if xid == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err)
		}

		ch := p.GetInfoChannel()

		go p.Server()

		wf := <-ch

		if wf.ID != xid {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}

	})

}

func TestPool_AddingOnRunningServer(t *testing.T) {

	t.Run("A simple information", func(t *testing.T) {

		ch := make(chan struct{ ID string })

		p := New(1)

		xid1, err1 := p.Add(func(taskID string) error {
			time.Sleep(5 * time.Millisecond)

			return nil
		})

		if xid1 == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err1 != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err1)
		}

		go p.Server()

		xid2, err2 := p.Add(func(taskID string) error {
			time.Sleep(10 * time.Millisecond)
			ch <- struct{ ID string }{ID: taskID}
			return nil
		})

		if err2 != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err1)
		}

		wf := <-ch

		if wf.ID != xid2 {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}

	})

}
