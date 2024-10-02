package test

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/rollkit/go-sequencing"
)

// ErrInvalidRollupId is returned when the rollup id is invalid
var ErrInvalidRollupId = errors.New("invalid rollup id")

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

func totalBytes(data [][]byte) int {
	total := 0
	for _, sub := range data {
		total += len(sub)
	}
	return total
}

// GetNextBatch extracts a batch of transactions from the queue
func (tq *TransactionQueue) GetNextBatch(maxBytes uint64) *sequencing.Batch {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	var batch [][]byte
	batchSize := uint64(len(tq.queue))
	if batchSize == 0 {
		return &sequencing.Batch{Transactions: nil}
	}

	if maxBytes == 0 {
		maxBytes = math.MaxUint64
	}

	for batchSize > 0 {
		batch = tq.queue[:batchSize]
		blobSize := totalBytes(batch)
		if uint64(blobSize) <= maxBytes {
			break
		}
		batchSize = batchSize - 1
	}
	if batchSize == 0 {
		// No transactions can fit within maxBytes
		return &sequencing.Batch{Transactions: nil}
	}

	tq.queue = tq.queue[batchSize:]
	return &sequencing.Batch{Transactions: batch}
}

// DummySequencer is a dummy sequencer for testing that serves a single rollup
type DummySequencer struct {
	rollupId           []byte
	tq                 *TransactionQueue
	lastBatchHash      []byte
	lastBatchHashMutex sync.RWMutex

	seenBatches      map[string]struct{}
	seenBatchesMutex sync.Mutex
}

// SubmitRollupTransaction implements sequencing.Sequencer.
func (d *DummySequencer) SubmitRollupTransaction(ctx context.Context, req sequencing.SubmitRollupTransactionRequest) (*sequencing.SubmitRollupTransactionResponse, error) {
	if !d.isValid(req.RollupId) {
		return nil, ErrInvalidRollupId
	}
	d.tq.AddTransaction(req.Tx)
	return &sequencing.SubmitRollupTransactionResponse{}, nil
}

// GetNextBatch implements sequencing.Sequencer.
func (d *DummySequencer) GetNextBatch(ctx context.Context, req sequencing.GetNextBatchRequest) (*sequencing.GetNextBatchResponse, error) {
	if !d.isValid(req.RollupId) {
		return nil, ErrInvalidRollupId
	}
	now := time.Now()
	d.lastBatchHashMutex.RLock()
	lastBatchHash := d.lastBatchHash
	d.lastBatchHashMutex.RUnlock()

	if lastBatchHash == nil && req.LastBatchHash != nil {
		return nil, errors.New("lastBatch is supposed to be nil")
	} else if lastBatchHash != nil && req.LastBatchHash == nil {
		return nil, errors.New("lastBatch is not supposed to be nil")
	} else if !bytes.Equal(lastBatchHash, req.LastBatchHash) {
		return nil, errors.New("supplied lastBatch does not match with sequencer last batch")
	}

	batch := d.tq.GetNextBatch(req.MaxBytes)
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
	d.seenBatches[hex.EncodeToString(h)] = struct{}{}
	d.seenBatchesMutex.Unlock()
	return batchRes, nil
}

// VerifyBatch implements sequencing.Sequencer.
func (d *DummySequencer) VerifyBatch(ctx context.Context, req sequencing.VerifyBatchRequest) (*sequencing.VerifyBatchResponse, error) {
	if !d.isValid(req.RollupId) {
		return nil, ErrInvalidRollupId
	}
	d.seenBatchesMutex.Lock()
	defer d.seenBatchesMutex.Unlock()
	key := hex.EncodeToString(req.BatchHash)
	if _, exists := d.seenBatches[key]; exists {
		return &sequencing.VerifyBatchResponse{Status: true}, nil
	}
	return &sequencing.VerifyBatchResponse{Status: false}, nil
}

func (d *DummySequencer) isValid(rollupId []byte) bool {
	return bytes.Equal(d.rollupId, rollupId)
}

// NewDummySequencer creates a new DummySequencer
func NewDummySequencer(rollupId []byte) *DummySequencer {
	return &DummySequencer{
		rollupId:    rollupId,
		tq:          NewTransactionQueue(),
		seenBatches: make(map[string]struct{}, 0),
	}
}

var _ sequencing.Sequencer = &DummySequencer{}
