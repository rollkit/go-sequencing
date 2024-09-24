package test

import (
	"bytes"
	"context"
	"crypto/sha256"
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

	tq            *TransactionQueue
	lastBatchHash []byte

	seenBatches map[string]struct{}
}

// SubmitRollupTransaction implements sequencing.Sequencer.
func (d *DummySequencer) SubmitRollupTransaction(ctx context.Context, rollupId []byte, tx []byte) error {
	if d.RollupId == nil {
		d.RollupId = rollupId
	} else {
		if !bytes.Equal(d.RollupId, rollupId) {
			return ErrorRollupIdMismatch
		}
	}
	d.tq.AddTransaction(tx)
	return nil
}

// GetNextBatch implements sequencing.Sequencer.
func (d *DummySequencer) GetNextBatch(ctx context.Context, lastBatchHash []byte) (*sequencing.Batch, time.Time, error) {
	now := time.Now()
	if d.lastBatchHash == nil {
		if lastBatchHash != nil {
			return nil, now, errors.New("lastBatch is supposed to be nil")
		}
	} else if lastBatchHash == nil {
		return nil, now, errors.New("lastBatch is not supposed to be nil")
	} else {
		if !bytes.Equal(d.lastBatchHash, lastBatchHash) {
			return nil, now, errors.New("supplied lastBatch does not match with sequencer last batch")
		}
	}

	batch := d.tq.GetNextBatch()
	// If there are no transactions, return empty batch without updating the last batch hash
	if batch.Transactions == nil {
		return batch, now, nil
	}

	batchBytes, err := batch.Marshal()
	if err != nil {
		return nil, now, err
	}

	d.lastBatchHash = hashSHA256(batchBytes)
	d.seenBatches[string(d.lastBatchHash)] = struct{}{}
	return batch, now, nil
}

// VerifyBatch implements sequencing.Sequencer.
func (d *DummySequencer) VerifyBatch(ctx context.Context, batchHash []byte) (bool, error) {
	_, ok := d.seenBatches[string(batchHash)]
	return ok, nil
}

// NewDummySequencer creates a new DummySequencer
func NewDummySequencer() *DummySequencer {
	return &DummySequencer{
		tq:          NewTransactionQueue(),
		seenBatches: make(map[string]struct{}, 0),
	}
}

func hashSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

var _ sequencing.Sequencer = &DummySequencer{}
