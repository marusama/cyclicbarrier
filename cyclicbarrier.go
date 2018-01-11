// Copyright 2018 Maru Sama. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package cyclicbarrier provides an implementation of Cyclic Barrier primitive.
package cyclicbarrier // import "github.com/marusama/cyclicbarrier"

import (
	"context"
	"sync"
)

type CyclicBarrier interface {
	Await(ctx context.Context) error
	Reset()
}

type cyclicBarrier struct {
	parties       int
	barrierAction func()

	lock   sync.Mutex
	count  int
	waitCh chan struct{}
}

func New(parties int) CyclicBarrier {
	if parties <= 0 {
		panic("parties must be positive number")
	}
	waitCh := make(chan struct{})
	return &cyclicBarrier{
		parties: parties,
		waitCh:  waitCh,
	}
}

func NewWithAction(parties int, barrierAction func()) CyclicBarrier {
	if parties <= 0 {
		panic("parties must be positive number")
	}
	waitCh := make(chan struct{})
	return &cyclicBarrier{
		parties:       parties,
		barrierAction: barrierAction,
		waitCh:        waitCh,
	}
}

func (b *cyclicBarrier) Await(ctx context.Context) error {

	b.lock.Lock()
	b.count++

	// saving count and wait channel in local variables to prevent race
	waitCh := b.waitCh
	count := b.count

	b.lock.Unlock()

	if count > b.parties {
		panic("CyclicBarrier.Await is called more than total parties count times")
	}

	if count < b.parties {
		// wait other parties
		if ctx != nil {
			select {
			case <-waitCh:
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			<-waitCh
		}
	} else {
		// we are last, break the barrier
		if b.barrierAction != nil {
			b.barrierAction()
		}
		b.Reset()
	}

	return nil
}

func (b *cyclicBarrier) Reset() {
	b.lock.Lock()

	b.count = 0
	close(b.waitCh)
	b.waitCh = make(chan struct{})

	b.lock.Unlock()
}
