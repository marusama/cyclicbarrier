// Copyright 2018 Maru Sama. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package cyclicbarrier provides an implementation of Cyclic Barrier primitive.
package cyclicbarrier // import "github.com/marusama/cyclicbarrier"

import (
	"context"
	"errors"
	"sync"
)

// CyclicBarrier is a synchronizer that allows a set of goroutines to wait for each other
// to reach a common execution point, also called a barrier.
// CyclicBarriers are useful in programs involving a fixed sized party of goroutines
// that must occasionally wait for each other.
// The barrier is called cyclic because it can be re-used after the waiting goroutines are released.
// A CyclicBarrier supports an optional Runnable command that is run once per barrier point,
// after the last goroutine in the party arrives, but before any goroutines are released.
// This barrier action is useful for updating shared-state before any of the parties continue.
type CyclicBarrier interface {
	// Await waits until all parties have invoked await on this barrier.
	// If the barrier is reset while any goroutine is waiting, or if the barrier is broken when await is invoked,
	// or while any goroutine is waiting, then ErrBrokenBarrier is returned.
	// If any goroutine is interrupted by ctx.Done() while waiting, then all other waiting goroutines
	// will return ErrBrokenBarrier and the barrier is placed in the broken state.
	// If the current goroutine is the last goroutine to arrive, and a non-nil barrier action was supplied in the constructor,
	// then the current goroutine runs the action before allowing the other goroutines to continue.
	// If an error occurs during the barrier action then that error will be returned and the barrier is placed in the broken state.
	Await(ctx context.Context) error

	// Reset resets the barrier to its initial state.
	// If any parties are currently waiting at the barrier, they will return with a ErrBrokenBarrier.
	Reset()

	// GetNumberWaiting returns the number of parties currently waiting at the barrier.
	GetNumberWaiting() int

	// GetParties returns the number of parties required to trip this barrier.
	GetParties() int

	// IsBroken queries if this barrier is in a broken state.
	// Returns true if one or more parties broke out of this barrier due to interruption by ctx.Done() or the last reset,
	// or a barrier action failed due to an error; false otherwise.
	IsBroken() bool
}

var (
	// ErrBrokenBarrier error used when a goroutine tries to wait upon a barrier that is in a broken state,
	// or which enters the broken state while the goroutine is waiting.
	ErrBrokenBarrier = errors.New("broken barrier")
)

// round
type round struct {
	count    int           // count of goroutines for this roundtrip
	waitCh   chan struct{} // wait channel for this roundtrip
	brokeCh  chan struct{} // channel for isBroken broadcast
	isBroken bool          // is barrier broken
}

// cyclicBarrier impl CyclicBarrier intf
type cyclicBarrier struct {
	parties       int
	barrierAction func() error

	lock  sync.RWMutex
	round *round
}

// New initializes a new instance of the CyclicBarrier, specifying the number of parties.
func New(parties int) CyclicBarrier {
	if parties <= 0 {
		panic("parties must be positive number")
	}
	return &cyclicBarrier{
		parties: parties,
		lock:    sync.RWMutex{},
		round: &round{
			waitCh:  make(chan struct{}),
			brokeCh: make(chan struct{}),
		},
	}
}

// NewWithAction initializes a new instance of the CyclicBarrier,
// specifying the number of parties and the barrier action.
func NewWithAction(parties int, barrierAction func() error) CyclicBarrier {
	if parties <= 0 {
		panic("parties must be positive number")
	}
	return &cyclicBarrier{
		parties: parties,
		lock:    sync.RWMutex{},
		round: &round{
			waitCh:  make(chan struct{}),
			brokeCh: make(chan struct{}),
		},
		barrierAction: barrierAction,
	}
}

func (b *cyclicBarrier) Await(ctx context.Context) error {
	var (
		ctxDoneCh <-chan struct{}
	)
	if ctx != nil {
		ctxDoneCh = ctx.Done()
	}

	// check if context is done
	select {
	case <-ctxDoneCh:
		return ctx.Err()
	default:
	}

	b.lock.Lock()

	// check if broken
	if b.round.isBroken {
		b.lock.Unlock()
		return ErrBrokenBarrier
	}

	// increment count of waiters
	b.round.count++

	// saving in local variables to prevent race
	waitCh := b.round.waitCh
	brokeCh := b.round.brokeCh
	count := b.round.count

	b.lock.Unlock()

	if count > b.parties {
		panic("CyclicBarrier.Await is called more than count of parties")
	}

	if count < b.parties {
		// wait other parties
		select {
		case <-waitCh:
			return nil
		case <-brokeCh:
			return ErrBrokenBarrier
		case <-ctxDoneCh:
			b.breakBarrier(true)
			return ctx.Err()
		}
	} else {
		// we are last, run the barrier action and reset the barrier
		if b.barrierAction != nil {
			err := b.barrierAction()
			if err != nil {
				b.breakBarrier(true)
				return err
			}
		}
		b.reset(true)
		return nil
	}
}

func (b *cyclicBarrier) breakBarrier(needLock bool) {
	if needLock {
		b.lock.Lock()
		defer b.lock.Unlock()
	}

	if !b.round.isBroken {
		b.round.isBroken = true

		// broadcast
		close(b.round.brokeCh)
	}
}

func (b *cyclicBarrier) reset(safe bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if safe {
		// broadcast to pass waiting goroutines
		close(b.round.waitCh)

	} else if b.round.count > 0 {
		b.breakBarrier(false)
	}

	// create new round
	b.round = &round{
		waitCh:  make(chan struct{}),
		brokeCh: make(chan struct{}),
	}
}

func (b *cyclicBarrier) Reset() {
	b.reset(false)
}

func (b *cyclicBarrier) GetParties() int {
	return b.parties
}

func (b *cyclicBarrier) GetNumberWaiting() int {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.round.count
}

func (b *cyclicBarrier) IsBroken() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.round.isBroken
}
