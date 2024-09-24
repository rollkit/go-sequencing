package sequencing

import (
	"context"
	"time"
)

// Sequencer is a generic interface for a rollup sequencer
type Sequencer interface {
	SequencerInput
	SequencerOutput
	BatchVerifier
}

// SequencerInput provides a method for submitting a transaction from rollup to sequencer
type SequencerInput interface {
	// SubmitRollupTransaction submits a transaction from rollup to sequencer
	SubmitRollupTransaction(ctx context.Context, rollupId RollupId, tx Tx) error
}

// SequencerOutput provides a method for getting the next batch of transactions from sequencer to rollup
type SequencerOutput interface {
	// GetNextBatch returns the next batch of transactions from sequencer to rollup
	// lastBatch is the last batch of transactions received from the sequencer
	// returns the next batch of transactions and an error if any from the sequencer
	GetNextBatch(ctx context.Context, lastBatchHash Hash) (*Batch, time.Time, error)
}

// BatchVerifier provides a method for verifying a batch of transactions received from the sequencer
type BatchVerifier interface {
	// VerifyBatch verifies a batch of transactions received from the sequencer
	VerifyBatch(ctx context.Context, batchHash Hash) (bool, error)
}

// RollupId is a unique identifier for a rollup chain
type RollupId = []byte

// Tx is a rollup transaction
type Tx = []byte

// Hash is a cryptographic hash of the Batch
type Hash = []byte

// Batch is a collection of transactions
type Batch struct {
	Transactions []Tx
}
