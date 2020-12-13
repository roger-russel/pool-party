package pool

import (
	"testing"
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
