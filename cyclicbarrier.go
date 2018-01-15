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

type round struct {
	count  int
	waitCh chan struct{}
}

// cyclicBarrier impl CyclicBarrier intf
type cyclicBarrier struct {
	parties       int
	barrierAction func()

	lock  sync.Mutex
	round *round
	// count  int
	// waitCh chan struct{}

	// waiters int32

	// bufCh   chan struct{}
	// inReset int32
}

// New initializes a new instance of the CyclicBarrier, specifying the number of parties.
func New(parties int) CyclicBarrier {
	if parties <= 0 {
		panic("parties must be positive number")
	}
	//waitCh := make(chan struct{})
	//bufCh := make(chan struct{}, parties-1)
	return &cyclicBarrier{
		parties: parties,
		lock:    sync.Mutex{},
		round: &round{
			waitCh: make(chan struct{}),
		},

		//bufCh: bufCh,
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

// func (b *cyclicBarrier) Await(ctx context.Context) error {
// 	var ctxDoneCh <-chan struct{}
// 	if ctx != nil {
// 		ctxDoneCh = ctx.Done()
// 	}

// 	select {
// 	case b.bufCh <- struct{}{}:
// 		select {
// 		case <-b.waitCh:
// 		case <-ctxDoneCh:
// 			return ctx.Err()
// 		}
// 	case <-ctxDoneCh:
// 		b.Reset()
// 		return ctx.Err()
// 	default:
// 		if b.barrierAction != nil {
// 			b.barrierAction()
// 		}
// 		b.Reset()
// 	}
// 	return nil
// }

// func (b *cyclicBarrier) Reset() {
// 	b.lock.Lock()
// 	defer b.lock.Unlock()

// 	for i := 0; i < b.parties-1; i++ {
// 		<-b.bufCh
// 	}

// 	close(b.waitCh)
// 	b.waitCh = make(chan struct{})
// }

//func (b *cyclicBarrier) Await(ctx context.Context, num int) error {
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

	//println(num, "count++", count)

	// decrement count
	// decrementFunc := func() {
	// 	b.lock.Lock()
	// 	b.count--
	// 	//println(num, "count--", b.count)
	// 	b.lock.Unlock()
	// }

	if count > b.parties {
		panic("CyclicBarrier.Await is called more than total parties count times")
	}

	if count < b.parties {
		//println(num, "wait")

		//atomic.AddInt32(&b.waiters, 1)

		// wait other parties
		select {
		case <-waitCh:
			//decrementFunc()
			//atomic.AddInt32(&b.waiters, -1)
			return nil
		case <-ctxDoneCh:
			//decrementFunc()
			//atomic.AddInt32(&b.waiters, -1)
			return ctx.Err()
		}
	} else {
		//println(num, "reset")
		// we are last, run the barrier action and break the barrier
		if b.barrierAction != nil {
			b.barrierAction()
		}
		//decrementFunc()
		b.Reset()

		return nil
	}
}

func (b *cyclicBarrier) Reset() {
	b.lock.Lock()

	//close(b.waitCh)
	//b.waitCh = make(chan struct{})
	close(b.round.waitCh)
	b.round = &round{
		waitCh: make(chan struct{}),
	}

	b.lock.Unlock()
}
