// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
)

var (
	verifiedPrefix          = []byte("verified")
	preferredMetadataPrefix = []byte("preferred_metadata")
	preferredKey            = []byte("preferred")
)

type VerifiedIndex interface {
	PutVerifiedBlock(blkID ids.ID) error
	PutPreference(preferredID ids.ID) error
}

type verifiedIndex struct {
	verifiedBlkIDsDB database.Database
	metadataDB       database.Database
}

func NewVerifiedIndex(db database.Database, blkState BlockState) (*verifiedIndex, error) {
	v := &verifiedIndex{
		verifiedBlkIDsDB: prefixdb.New(verifiedPrefix, db),
		metadataDB:       prefixdb.New(preferredMetadataPrefix, db),
	}

	preferredID, err := v.metadataDB.Get(preferredKey)
	switch {
	case err != nil && err != database.ErrNotFound:
		return nil, err
	case err == database.ErrNotFound:
		return v, nil
	}

	// Add the preferred chain up to the last accepted block to be skipped.
	preferredBlkIDs := make(map[string]struct{})
	preferredBlk, status, err := blkState.GetBlock(ids.ID(preferredID))
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

	// Delete any verified blocks that are not in the preferred chain.
	iter := v.verifiedBlkIDsDB.NewIterator()
	defer iter.Release()

	for iter.Next() {
		if _, ok := preferredBlkIDs[string(iter.Key())]; ok {
			continue
		}
		if err := v.verifiedBlkIDsDB.Delete(iter.Key()); err != nil {
			return nil, err
		}
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *verifiedIndex) PutVerifiedBlock(blkID ids.ID) error {
	return v.verifiedBlkIDsDB.Put(blkID[:], nil)
}

func (v *verifiedIndex) PutPreference(preferredID ids.ID) error {
	return v.metadataDB.Put(preferredKey, preferredID[:])
}
