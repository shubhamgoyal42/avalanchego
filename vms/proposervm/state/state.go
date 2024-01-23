// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/database/versiondb"
)

var (
	chainStatePrefix  = []byte("chain")
	blockStatePrefix  = []byte("block")
	heightIndexPrefix = []byte("height")
)

type State interface {
	ChainState
	BlockState
	HeightIndex
}

type state struct {
	ChainState
	BlockState
	HeightIndex
}

func New(db *versiondb.Database) (State, error) {
	chainDB := prefixdb.New(chainStatePrefix, db)
	blockDB := prefixdb.New(blockStatePrefix, db)
	heightDB := prefixdb.New(heightIndexPrefix, db)

	blockState := NewBlockState(blockDB)
	chainState, err := NewChainState(chainDB, blockState)
	if err != nil {
		return nil, err
	}

	return &state{
		ChainState:  chainState,
		BlockState:  blockState,
		HeightIndex: NewHeightIndex(heightDB, db),
	}, nil
}

func NewMetered(db *versiondb.Database, namespace string, metrics prometheus.Registerer) (State, error) {
	chainDB := prefixdb.New(chainStatePrefix, db)
	blockDB := prefixdb.New(blockStatePrefix, db)
	heightDB := prefixdb.New(heightIndexPrefix, db)

	blockState, err := NewMeteredBlockState(blockDB, namespace, metrics)
	if err != nil {
		return nil, err
	}
	chainState, err := NewChainState(chainDB, blockState)
	if err != nil {
		return nil, err
	}

	return &state{
		ChainState:  chainState,
		BlockState:  blockState,
		HeightIndex: NewHeightIndex(heightDB, db),
	}, nil
}
