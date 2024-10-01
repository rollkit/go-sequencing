package test

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rollkit/go-sequencing"
)

// ErrorRollupIdMismatch is returned when the rollup id does not match
var ErrorRollupIdMismatch = errors.New("rollup id mismatch")

// TransactionQueue is a queue of transactions
type TransactionQueue struct {
	queue []sequencing.Tx
	mu    sync.Mutex
}

// NewTransactionQueue creates a new TransactionQueue
func NewTransactionQueue() *TransactionQueue {
	return &TransactionQueue{
		queue: make([]sequencing.Tx, 0),
	}
}

// AddTransaction adds a new transaction to the queue
func (tq *TransactionQueue) AddTransaction(tx sequencing.Tx) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tq.queue = append(tq.queue, tx)
}

// GetNextBatch extracts a batch of transactions from the queue
func (tq *TransactionQueue) GetNextBatch() *sequencing.Batch {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	size := len(tq.queue)
	if size == 0 {
		return &sequencing.Batch{Transactions: nil}
	}
	batch := tq.queue[:size]
	tq.queue = tq.queue[size:]
	return &sequencing.Batch{Transactions: batch}
}

// DummySequencer is a dummy sequencer for testing that serves a single rollup
type DummySequencer struct {
	sequencing.RollupId

	tq                 *TransactionQueue
	lastBatchHash      []byte
	lastBatchHashMutex sync.RWMutex

	seenBatches      map[string]struct{}
	seenBatchesMutex sync.Mutex
}

// SubmitRollupTransaction implements sequencing.Sequencer.
func (d *DummySequencer) SubmitRollupTransaction(ctx context.Context, req sequencing.SubmitRollupTransactionRequest) (*sequencing.SubmitRollupTransactionResponse, error) {
	if d.RollupId == nil {
		d.RollupId = req.RollupId
	} else {
		if !bytes.Equal(d.RollupId, req.RollupId) {
			return nil, ErrorRollupIdMismatch
		}
	}
	d.tq.AddTransaction(req.Tx)
	return &sequencing.SubmitRollupTransactionResponse{}, nil
}

// GetNextBatch implements sequencing.Sequencer.
func (d *DummySequencer) GetNextBatch(ctx context.Context, req sequencing.GetNextBatchRequest) (*sequencing.GetNextBatchResponse, error) {
	now := time.Now()
	d.lastBatchHashMutex.RLock()
	lastBatchHash := d.lastBatchHash
	d.lastBatchHashMutex.RUnlock()
	if lastBatchHash == nil {
		if req.LastBatchHash != nil {
			return nil, errors.New("lastBatch is supposed to be nil")
		}
	} else if req.LastBatchHash == nil {
		return nil, errors.New("lastBatch is not supposed to be nil")
	} else {
		if !bytes.Equal(lastBatchHash, req.LastBatchHash) {
			return nil, errors.New("supplied lastBatch does not match with sequencer last batch")
		}
	}

	batch := d.tq.GetNextBatch()
	batchRes := &sequencing.GetNextBatchResponse{Batch: batch, Timestamp: now}
	// If there are no transactions, return empty batch without updating the last batch hash
	if batch.Transactions == nil {
		return batchRes, nil
	}

	h, err := batch.Hash()
	if err != nil {
		return nil, err
	}

	d.lastBatchHashMutex.Lock()
	d.lastBatchHash = h
	d.lastBatchHashMutex.Unlock()

	d.seenBatchesMutex.Lock()
	d.seenBatches[string(h)] = struct{}{}
	d.seenBatchesMutex.Unlock()
	return batchRes, nil
}

// VerifyBatch implements sequencing.Sequencer.
func (d *DummySequencer) VerifyBatch(ctx context.Context, req sequencing.VerifyBatchRequest) (*sequencing.VerifyBatchResponse, error) {
	d.seenBatchesMutex.Lock()
	defer d.seenBatchesMutex.Unlock()
	for batchHash := range d.seenBatches {
		if bytes.Equal([]byte(batchHash), req.BatchHash) {
			return &sequencing.VerifyBatchResponse{Status: true}, nil
		}
	}
	return &sequencing.VerifyBatchResponse{Status: false}, nil
}

// NewDummySequencer creates a new DummySequencer
func NewDummySequencer() *DummySequencer {
	return &DummySequencer{
		tq:          NewTransactionQueue(),
		seenBatches: make(map[string]struct{}, 0),
	}
}

var _ sequencing.Sequencer = &DummySequencer{}
