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

// cyclicBarrier impl CyclicBarrier intf
type cyclicBarrier struct {
	parties       int
	barrierAction func()

	lock   *sync.Mutex
	count  int
	waitCh chan struct{}

	bufCh chan struct{}
}

// New initializes a new instance of the CyclicBarrier, specifying the number of parties.
func New(parties int) CyclicBarrier {
	if parties <= 0 {
		panic("parties must be positive number")
	}
	waitCh := make(chan struct{})
	bufCh := make(chan struct{}, parties-1)
	return &cyclicBarrier{
		parties: parties,
		waitCh:  waitCh,
		lock:    &sync.Mutex{},

		bufCh: bufCh,
	}
}

// NewWithAction initializes a new instance of the CyclicBarrier,
// specifying the number of parties and the barrier action.
func NewWithAction(parties int, barrierAction func()) CyclicBarrier {
	if parties <= 0 {
		panic("parties must be positive number")
	}
	waitCh := make(chan struct{})
	return &cyclicBarrier{
		parties:       parties,
		barrierAction: barrierAction,
		waitCh:        waitCh,
		lock:          &sync.Mutex{},
	}
}

func (b *cyclicBarrier) Await(ctx context.Context) error {
	var ctxDoneCh <-chan struct{}
	if ctx != nil {
		ctxDoneCh = ctx.Done()
	}

	b.lock.Lock()
	select {
	case b.bufCh <- struct{}{}:
		b.lock.Unlock()
		select {
		case <-b.waitCh:
		case <-ctxDoneCh:
			return ctx.Err()
		}
	case <-ctxDoneCh:
		b.lock.Unlock()
		b.Reset()
		return ctx.Err()
	default:
		b.lock.Unlock()
		if b.barrierAction != nil {
			b.barrierAction()
		}
		b.Reset()
	}
	return nil
}

func (b *cyclicBarrier) Reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	for i := 0; i < b.parties-1; i++ {
		<-b.bufCh
	}

	close(b.waitCh)
	b.waitCh = make(chan struct{})
}

func (b *cyclicBarrier) AwaitOld(ctx context.Context, num int) error {

	var ctxDoneCh <-chan struct{}
	if ctx != nil {
		ctxDoneCh = ctx.Done()
	}

	// increment count
	b.lock.Lock()
	b.count++

	// saving count and wait channel in local variables to prevent race
	waitCh := b.waitCh
	count := b.count

	b.lock.Unlock()

	println(num, "count++", count)

	// decrement count
	decrementFunc := func() {
		b.lock.Lock()
		b.count--
		println(num, "count--", b.count)
		b.lock.Unlock()
	}

	if count > b.parties {
		panic("CyclicBarrier.Await is called more than total parties count times")
	}

	if count < b.parties {
		println(num, "wait")

		// wait other parties
		select {
		case <-waitCh:
			decrementFunc()
			return nil
		case <-ctxDoneCh:
			decrementFunc()
			return ctx.Err()
		}
	} else {
		println(num, "reset")
		// we are last, run the barrier action and break the barrier
		if b.barrierAction != nil {
			b.barrierAction()
		}
		decrementFunc()
		b.Reset()

		return nil
	}
}

func (b *cyclicBarrier) ResetOld() {
	b.lock.Lock()

	close(b.waitCh)
	b.waitCh = make(chan struct{})

	b.lock.Unlock()
}
