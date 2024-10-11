package test

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/rollkit/go-sequencing"
)

// MultiRollupSequencer is a sequencer for testing that serves multiple rollups
type MultiRollupSequencer struct {
	rollups      map[string]*RollupData
	rollupsMutex sync.RWMutex
}

// RollupData holds the data for a specific rollup, including its transaction queue, last batch hash, and seen batches.
type RollupData struct {
	tq                 *TransactionQueue
	lastBatchHash      []byte
	lastBatchHashMutex sync.RWMutex

	seenBatches      map[string]struct{}
	seenBatchesMutex sync.Mutex
}

// SubmitRollupTransaction implements sequencing.Sequencer.
func (d *MultiRollupSequencer) SubmitRollupTransaction(ctx context.Context, req sequencing.SubmitRollupTransactionRequest) (*sequencing.SubmitRollupTransactionResponse, error) {
	rollup, err := d.getOrCreateRollup(req.RollupId)
	if err != nil {
		return nil, err
	}
	rollup.tq.AddTransaction(req.Tx)
	return &sequencing.SubmitRollupTransactionResponse{}, nil
}

// GetNextBatch implements sequencing.Sequencer.
func (d *MultiRollupSequencer) GetNextBatch(ctx context.Context, req sequencing.GetNextBatchRequest) (*sequencing.GetNextBatchResponse, error) {
	rollup, err := d.getOrCreateRollup(req.RollupId)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	rollup.lastBatchHashMutex.RLock()
	lastBatchHash := rollup.lastBatchHash
	rollup.lastBatchHashMutex.RUnlock()

	if lastBatchHash == nil && req.LastBatchHash != nil {
		return nil, errors.New("lastBatch is supposed to be nil")
	} else if lastBatchHash != nil && req.LastBatchHash == nil {
		return nil, errors.New("lastBatch is not supposed to be nil")
	} else if !bytes.Equal(lastBatchHash, req.LastBatchHash) {
		return nil, errors.New("supplied lastBatch does not match with sequencer last batch")
	}

	batch := rollup.tq.GetNextBatch(req.MaxBytes)
	batchRes := &sequencing.GetNextBatchResponse{Batch: batch, Timestamp: now}
	// If there are no transactions, return empty batch without updating the last batch hash
	if len(batch.Transactions) == 0 {
		return batchRes, nil
	}

	h, err := batch.Hash()
	if err != nil {
		return nil, err
	}

	rollup.lastBatchHashMutex.Lock()
	rollup.lastBatchHash = h
	rollup.lastBatchHashMutex.Unlock()

	rollup.seenBatchesMutex.Lock()
	rollup.seenBatches[hex.EncodeToString(h)] = struct{}{}
	rollup.seenBatchesMutex.Unlock()
	return batchRes, nil
}

// VerifyBatch implements sequencing.Sequencer.
func (d *MultiRollupSequencer) VerifyBatch(ctx context.Context, req sequencing.VerifyBatchRequest) (*sequencing.VerifyBatchResponse, error) {
	rollup, err := d.getOrCreateRollup(req.RollupId)
	if err != nil {
		return nil, err
	}

	rollup.seenBatchesMutex.Lock()
	defer rollup.seenBatchesMutex.Unlock()
	key := hex.EncodeToString(req.BatchHash)
	if _, exists := rollup.seenBatches[key]; exists {
		return &sequencing.VerifyBatchResponse{Status: true}, nil
	}
	return &sequencing.VerifyBatchResponse{Status: false}, nil
}

// getOrCreateRollup returns the RollupData for a given rollupId, creating it if necessary.
func (d *MultiRollupSequencer) getOrCreateRollup(rollupId []byte) (*RollupData, error) {
	rollupKey := hex.EncodeToString(rollupId)

	d.rollupsMutex.RLock()
	rollup, exists := d.rollups[rollupKey]
	d.rollupsMutex.RUnlock()

	if exists {
		return rollup, nil
	}

	d.rollupsMutex.Lock()
	defer d.rollupsMutex.Unlock()

	// Double-check existence after acquiring write lock
	if rollup, exists := d.rollups[rollupKey]; exists {
		return rollup, nil
	}

	// Create a new RollupData if it doesn't exist
	rollup = &RollupData{
		tq:          NewTransactionQueue(),
		seenBatches: make(map[string]struct{}, 0),
	}
	d.rollups[rollupKey] = rollup
	return rollup, nil
}

// NewMultiRollupSequencer creates a new MultiRollupSequencer
func NewMultiRollupSequencer() *MultiRollupSequencer {
	return &MultiRollupSequencer{
		rollups: make(map[string]*RollupData),
	}
}

var _ sequencing.Sequencer = &MultiRollupSequencer{}
