// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/vms/proposervm/block"
)

type VerifiedIndex interface {
	PutVerifiedBlock(block block.Block) error
}

type verifiedIndex struct {
	db database.Database
}

func NewVerifiedIndex(db database.Database, blkState BlockState, preferenceID ids.ID) (*verifiedIndex, error) {
	preferredBlkIDs := make(map[string]struct{})
	_, err := db.Get(preferenceID[:])
	switch {
	case err != nil && err != database.ErrNotFound:
		return nil, err
	case err == database.ErrNotFound:
	default:
		preferredBlk, status, err := blkState.GetBlock(preferenceID)
		if err != nil {
			return nil, err
		}
		for status != choices.Accepted {
			preferredID := preferredBlk.ID()
			preferredBlkIDs[string(preferredID[:])] = struct{}{}

			preferredBlk, status, err = blkState.GetBlock(preferredBlk.ParentID())
			if err != nil {
				return nil, err
			}
		}
	}

	iter := db.NewIterator()
	defer iter.Release()

	for iter.Next() {
		if _, ok := preferredBlkIDs[string(iter.Key())]; ok {
			continue
		}
		if err := db.Delete(iter.Key()); err != nil {
			return nil, err
		}
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}

	return &verifiedIndex{db: db}, nil
}

func (v *verifiedIndex) PutVerifiedBlock(block block.Block) error {
	blkID := block.ID()
	return v.db.Put(blkID[:], nil)
}
