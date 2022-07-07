// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"sync"

	"github.com/google/btree"
)

var _ StakerIterator = &treeIterator{}

type treeIterator struct {
	current     *Staker
	next        chan *Staker
	releaseOnce sync.Once
	release     chan struct{}
	wg          sync.WaitGroup
}

func NewTreeIterator(tree *btree.BTree) StakerIterator {
	it := &treeIterator{
		next:    make(chan *Staker),
		release: make(chan struct{}),
	}
	it.wg.Add(1)
	go func() {
		defer it.wg.Done()
		tree.Ascend(func(i btree.Item) bool {
			select {
			case it.next <- i.(*Staker):
				return true
			case <-it.release:
				return false
			}
		})
		close(it.next)
	}()
	return it
}

func (i *treeIterator) Next() bool {
	next, ok := <-i.next
	i.current = next
	return ok
}

func (i *treeIterator) Value() *Staker {
	return i.current
}

func (i *treeIterator) Release() {
	i.releaseOnce.Do(func() {
		close(i.release)
	})
	i.wg.Wait()
}
