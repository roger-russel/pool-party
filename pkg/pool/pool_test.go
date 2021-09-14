package pool

import (
	"testing"
)

func TestPool_NewPool(t *testing.T) {
	t.Run("Initializing a pool of size 1 with New Command", func(t *testing.T) {
		p := New(1)

		ch := make(chan struct{})

		add(t, p, func(taskID string) error {
			ch <- struct{}{}
			return nil
		})

		go p.Server()

		<-ch
	})
}

func TestPool_GettingDoneInfo(t *testing.T) {
	t.Run("A simple information", func(t *testing.T) {
		p := New(1)

		xid := add(t, p, func(taskID string) error {
			return nil
		})

		ch := p.GetInfoChannel()

		go p.Server()

		wf := <-ch

		if wf.ID != xid {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}
	})
}

func add(t *testing.T, p Pool, f func(taskID string) error) string {
	xid, err := p.Add(f)

	if xid == "" {
		t.Error("p.Add() returned an empty ID")
	}

	if err != nil {
		t.Errorf("p.Add() returned an inexpected error: %s", err)
	}

	return xid
}

func toMap(t *testing.T, slice []string) map[string]int {
	m := make(map[string]int)

	for i, v := range slice {
		if ip, ok := m[v]; ok {
			t.Errorf("duplicated index detected: %s - %d, %d", v, i, ip)
		}

		m[v] = i
	}
	return m
}
