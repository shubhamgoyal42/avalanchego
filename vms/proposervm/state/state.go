// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/database/versiondb"
)

var (
	chainStatePrefix    = []byte("chain")
	blockStatePrefix    = []byte("block")
	heightIndexPrefix   = []byte("height")
	verifiedIndexPrefix = []byte("verified")
)

type State interface {
	ChainState
	BlockState
	HeightIndex
	VerifiedIndex
}

type state struct {
	ChainState
	BlockState
	HeightIndex
	VerifiedIndex
}

func New(db *versiondb.Database) (State, error) {
	chainDB := prefixdb.New(chainStatePrefix, db)
	blockDB := prefixdb.New(blockStatePrefix, db)
	heightDB := prefixdb.New(heightIndexPrefix, db)
	verifiedDB := prefixdb.New(verifiedIndexPrefix, db)

	blockState := NewBlockState(blockDB)
	verifiedIndex, err := NewVerifiedIndex(verifiedDB, blockState)
	if err != nil {
		return nil, err
	}

	return &state{
		ChainState:    NewChainState(chainDB),
		BlockState:    blockState,
		HeightIndex:   NewHeightIndex(heightDB, db),
		VerifiedIndex: verifiedIndex,
	}, nil
}

func NewMetered(db *versiondb.Database, namespace string, metrics prometheus.Registerer) (State, error) {
	chainDB := prefixdb.New(chainStatePrefix, db)
	blockDB := prefixdb.New(blockStatePrefix, db)
	heightDB := prefixdb.New(heightIndexPrefix, db)
	verifiedDB := prefixdb.New(verifiedIndexPrefix, db)

	blockState, err := NewMeteredBlockState(blockDB, namespace, metrics)
	if err != nil {
		return nil, err
	}
	verifiedIndex, err := NewVerifiedIndex(verifiedDB, blockState)
	if err != nil {
		return nil, err
	}

	return &state{
		ChainState:    NewChainState(chainDB),
		BlockState:    blockState,
		HeightIndex:   NewHeightIndex(heightDB, db),
		VerifiedIndex: verifiedIndex,
	}, nil
}
