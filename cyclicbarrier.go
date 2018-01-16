// Copyright 2018 Maru Sama. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package cyclicbarrier provides an implementation of Cyclic Barrier primitive.
package cyclicbarrier // import "github.com/marusama/cyclicbarrier"

import (
	"context"
	"sync"
)

// CyclicBarrier is a synchronizer that allows a set of goroutines to wait for each other
// to reach a common execution point, also called a barrier.
type CyclicBarrier interface {
	// Await waits until all parties have invoked await on this barrier.
	Await(ctx context.Context) error

	// Reset resets the barrier to its initial state.
	Reset()
}

// round
type round struct {
	count  int           // count of goroutines for this roundtrip
	waitCh chan struct{} // wait channel for this roundtrip
}

// cyclicBarrier impl CyclicBarrier intf
type cyclicBarrier struct {
	parties       int
	barrierAction func()

	lock  sync.Mutex
	round *round
}

// New initializes a new instance of the CyclicBarrier, specifying the number of parties.
func New(parties int) CyclicBarrier {
	if parties <= 0 {
		panic("parties must be positive number")
	}
	return &cyclicBarrier{
		parties: parties,
		lock:    sync.Mutex{},
		round: &round{
			waitCh: make(chan struct{}),
		},
	}
}

// NewWithAction initializes a new instance of the CyclicBarrier,
// specifying the number of parties and the barrier action.
func NewWithAction(parties int, barrierAction func()) CyclicBarrier {
	if parties <= 0 {
		panic("parties must be positive number")
	}
	return &cyclicBarrier{
		parties: parties,
		lock:    sync.Mutex{},
		round: &round{
			waitCh: make(chan struct{}),
		},
		barrierAction: barrierAction,
	}
}

func (b *cyclicBarrier) Await(ctx context.Context) error {

	var ctxDoneCh <-chan struct{}
	if ctx != nil {
		ctxDoneCh = ctx.Done()
	}

	// increment count
	b.lock.Lock()
	b.round.count++

	// saving count and wait channel in local variables to prevent race
	waitCh := b.round.waitCh
	count := b.round.count

	b.lock.Unlock()

	if count > b.parties {
		panic("CyclicBarrier.Await is called more than total parties count times")
	}

	if count < b.parties {
		select {
		case <-waitCh:
			return nil
		case <-ctxDoneCh:
			return ctx.Err()
		}
	} else {
		// we are last, run the barrier action and break the barrier
		if b.barrierAction != nil {
			b.barrierAction()
		}
		b.Reset()
		return nil
	}
}

func (b *cyclicBarrier) Reset() {
	b.lock.Lock()

	// broadcast to pass waiting goroutines
	close(b.round.waitCh)

	// create new round
	b.round = &round{
		waitCh: make(chan struct{}),
	}

	b.lock.Unlock()
}
