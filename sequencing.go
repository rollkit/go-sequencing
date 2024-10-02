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
	// RollupId is the unique identifier for the rollup chain
	// Tx is the transaction to submit
	// returns an error if any from the sequencer
	SubmitRollupTransaction(ctx context.Context, req SubmitRollupTransactionRequest) (*SubmitRollupTransactionResponse, error)
}

// SequencerOutput provides a method for getting the next batch of transactions from sequencer to rollup
type SequencerOutput interface {
	// GetNextBatch returns the next batch of transactions from sequencer to rollup
	// RollupId is the unique identifier for the rollup chain
	// LastBatchHash is the cryptographic hash of the last batch received by the rollup
	// MaxBytes is the maximum number of bytes to return in the batch
	// returns the next batch of transactions and an error if any from the sequencer
	GetNextBatch(ctx context.Context, req GetNextBatchRequest) (*GetNextBatchResponse, error)
}

// BatchVerifier provides a method for verifying a batch of transactions received from the sequencer
type BatchVerifier interface {
	// VerifyBatch verifies a batch of transactions received from the sequencer
	// RollupId is the unique identifier for the rollup chain
	// BatchHash is the cryptographic hash of the batch to verify
	// returns a boolean indicating if the batch is valid and an error if any from the sequencer
	VerifyBatch(ctx context.Context, req VerifyBatchRequest) (*VerifyBatchResponse, error)
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

// SubmitRollupTransactionRequest is a request to submit a transaction from rollup to sequencer
type SubmitRollupTransactionRequest struct {
	RollupId RollupId
	Tx       Tx
}

// SubmitRollupTransactionResponse is a response to submitting a transaction from rollup to sequencer
type SubmitRollupTransactionResponse struct {
}

// GetNextBatchRequest is a request to get the next batch of transactions from sequencer to rollup
type GetNextBatchRequest struct {
	RollupId      RollupId
	LastBatchHash Hash
	MaxBytes      uint64
}

// GetNextBatchResponse is a response to getting the next batch of transactions from sequencer to rollup
type GetNextBatchResponse struct {
	Batch     *Batch
	Timestamp time.Time
}

// VerifyBatchRequest is a request to verify a batch of transactions received from the sequencer
type VerifyBatchRequest struct {
	RollupId  RollupId
	BatchHash Hash
}

// VerifyBatchResponse is a response to verifying a batch of transactions received from the sequencer
type VerifyBatchResponse struct {
	Status bool
}
