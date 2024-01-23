// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
)

const (
	lastAcceptedByte byte = iota
	preferredByte
	verifiedByte
)

var (
	lastAcceptedKey = []byte{lastAcceptedByte}
	preferredKey    = []byte{preferredByte}

	_ ChainState = (*chainState)(nil)
)

type ChainState interface {
	SetLastAccepted(blkID ids.ID) error
	DeleteLastAccepted() error
	GetLastAccepted() (ids.ID, error)

	PutVerifiedBlock(blkID ids.ID) error
	DeleteVerifiedBlock(blkID ids.ID) error
	SetPreference(preferredID ids.ID) error
	Preferred() ids.ID
}

type chainState struct {
	lastAccepted ids.ID
	preferred    ids.ID
	db           database.Database
}

func NewChainState(db database.Database, blkState BlockState) (ChainState, error) {
	s := &chainState{db: db}
	return s, s.initPreferredChain(blkState)
}

func (s *chainState) initPreferredChain(blkState BlockState) error {
	preferredID, err := s.db.Get(preferredKey)
	switch {
	case err != nil && err != database.ErrNotFound:
		return err
	case err == database.ErrNotFound:
		return nil
	}

	// Add the preferred chain up to the last accepted block to be skipped.
	preferredBlkIDs := make(map[string]struct{})
	preferredBlk, status, err := blkState.GetBlock(ids.ID(preferredID))
	if err != nil {
		return err
	}
	for status != choices.Accepted {
		preferredID := preferredBlk.ID()
		preferredBlkIDs[string(preferredID[:])] = struct{}{}

		preferredBlk, status, err = blkState.GetBlock(preferredBlk.ParentID())
		if err != nil {
			return err
		}
	}

	// Delete any verified blocks that are not in the preferred chain.
	iter := s.db.NewIteratorWithPrefix([]byte{verifiedByte})
	defer iter.Release()

	for iter.Next() {
		blkID := ids.ID(iter.Key()[1:]) // Drop the one byte prefix
		if _, ok := preferredBlkIDs[string(blkID[:])]; ok {
			continue
		}
		if err := s.DeleteVerifiedBlock(blkID); err != nil {
			return err
		}
	}
	if err := iter.Error(); err != nil {
		return err
	}

	return nil
}

func (s *chainState) SetLastAccepted(blkID ids.ID) error {
	if s.lastAccepted == blkID {
		return nil
	}
	s.lastAccepted = blkID
	return s.db.Put(lastAcceptedKey, blkID[:])
}

func (s *chainState) DeleteLastAccepted() error {
	s.lastAccepted = ids.Empty
	return s.db.Delete(lastAcceptedKey)
}

func (s *chainState) GetLastAccepted() (ids.ID, error) {
	if s.lastAccepted != ids.Empty {
		return s.lastAccepted, nil
	}
	lastAcceptedBytes, err := s.db.Get(lastAcceptedKey)
	if err != nil {
		return ids.ID{}, err
	}
	lastAccepted, err := ids.ToID(lastAcceptedBytes)
	if err != nil {
		return ids.ID{}, err
	}
	s.lastAccepted = lastAccepted
	return lastAccepted, nil
}

func (s *chainState) PutVerifiedBlock(blkID ids.ID) error {
	return s.db.Put(makePreferreBlkIDKey(blkID), nil)
}

func (s *chainState) DeleteVerifiedBlock(blkID ids.ID) error {
	return s.db.Delete(makePreferreBlkIDKey(blkID))
}

func makePreferreBlkIDKey(blkID ids.ID) []byte {
	preferredBlkIDKey := make([]byte, ids.IDLen+1)
	preferredBlkIDKey[0] = verifiedByte
	copy(preferredBlkIDKey[1:], blkID[:])
	return preferredBlkIDKey
}

func (s *chainState) SetPreference(preferredID ids.ID) error {
	s.preferred = preferredID
	return s.db.Put(preferredKey, preferredID[:])
}

// XXX: returns ids.Empty if SetPreference has not been called yet
// currently handled differently from GetLastAccepted
func (s *chainState) Preferred() ids.ID {
	return s.preferred
}
