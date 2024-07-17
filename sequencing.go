package sequencing

// SequencerInput provides a method for submitting a transaction from rollup to sequencer
type SequencerInput interface {
	// SubmitRollupTransaction submits a transaction from rollup to sequencer
	SubmitRollupTransaction(rollupId RollupId, tx Tx) error
}

// SequencerOutput provides a method for getting the next batch of transactions from sequencer to rollup
type SequencerOutput interface {
	// GetNextBatch returns the next batch of transactions from sequencer to rollup
	// lastBatch is the last batch of transactions received from the sequencer
	// returns the next batch of transactions and an error if any from the sequencer
	GetNextBatch(lastBatch []Tx) ([]Tx, error)
}

// BatchVerifier provides a method for verifying a batch of transactions received from the sequencer
type BatchVerifier interface {
	// VerifyBatch verifies a batch of transactions received from the sequencer
	VerifyBatch(batch []Tx) (bool, error)
}

// RollupId is a unique identifier for a rollup chain
type RollupId = []byte

// Tx is a rollup transaction
type Tx = []byte
