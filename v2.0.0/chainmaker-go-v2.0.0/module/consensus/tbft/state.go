/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tbft

import (
	"sync"

	"chainmaker.org/chainmaker-go/logger"
	tbftpb "chainmaker.org/chainmaker/pb-go/v2/consensus/tbft"
)

// ConsensusState represents the consensus state of the node
type ConsensusState struct {
	logger *logger.CMLogger
	Id     string

	Height uint64
	Round  int32
	Step   tbftpb.Step

	Proposal           *Proposal // proposal
	VerifingProposal   *Proposal // verifing proposal
	LockedRound        int32
	LockedProposal     *Proposal // locked proposal
	ValidRound         int32
	ValidProposal      *Proposal // valid proposal
	heightRoundVoteSet *heightRoundVoteSet
}

// NewConsensusState creates a new ConsensusState instance
func NewConsensusState(logger *logger.CMLogger, id string) *ConsensusState {
	cs := &ConsensusState{
		logger: logger,
		Id:     id,
	}
	return cs
}

func (cs *ConsensusState) resetFromProto(csProto *tbftpb.ConsensusState, validatorSet *validatorSet) {
	cs.Height = csProto.Height
	cs.Round = csProto.Round
	cs.Step = csProto.Step
	cs.Proposal = NewProposalFromProto(csProto.Proposal)
	cs.VerifingProposal = NewProposalFromProto(csProto.VerifingProposal)
	cs.heightRoundVoteSet = newHeightRoundVoteSetFromProto(cs.logger, csProto.HeightRoundVoteSet, validatorSet)
}

// toProto serializes the ConsensusState instance
func (cs *ConsensusState) toProto() *tbftpb.ConsensusState {
	if cs == nil {
		return nil
	}
	csProto := &tbftpb.ConsensusState{
		Id:                 cs.Id,
		Height:             cs.Height,
		Round:              cs.Round,
		Step:               cs.Step,
		Proposal:           cs.Proposal.ToProto(),
		VerifingProposal:   cs.VerifingProposal.ToProto(),
		HeightRoundVoteSet: cs.heightRoundVoteSet.ToProto(),
	}
	return csProto
}

type consensusStateCache struct {
	sync.Mutex
	size  uint64
	cache map[uint64]*ConsensusState
}

func newConsensusStateCache(size uint64) *consensusStateCache {
	return &consensusStateCache{
		size:  size,
		cache: make(map[uint64]*ConsensusState, size),
	}
}

func (cache *consensusStateCache) addConsensusState(state *ConsensusState) {
	if state == nil || state.Height <= 0 {
		return
	}

	cache.Lock()
	defer cache.Unlock()

	cache.cache[state.Height] = state
	cache.gc(state.Height)
}

func (cache *consensusStateCache) getConsensusState(height uint64) *ConsensusState {
	cache.Lock()
	defer cache.Unlock()

	if state, ok := cache.cache[height]; ok {
		return state
	}

	return nil
}

func (cache *consensusStateCache) gc(height uint64) {
	for k := range cache.cache {
		if k < (height - cache.size) {
			delete(cache.cache, k)
		}
	}
}
