package test

import (
	"context"
	"testing"
	"time"

	"github.com/rollkit/go-sequencing"
	"github.com/stretchr/testify/assert"
)

func TestTransactionQueue_AddTransaction(t *testing.T) {
	queue := NewTransactionQueue()

	tx1 := []byte("transaction_1")
	queue.AddTransaction(tx1)

	// Check that the transaction was added
	batch := queue.GetNextBatch()
	assert.Equal(t, 1, len(batch.Transactions))
	assert.Equal(t, tx1, batch.Transactions[0])
}

func TestTransactionQueue_GetNextBatch(t *testing.T) {
	queue := NewTransactionQueue()

	// Add multiple transactions
	tx1 := []byte("transaction_1")
	tx2 := []byte("transaction_2")
	queue.AddTransaction(tx1)
	queue.AddTransaction(tx2)

	// Get next batch and check if all transactions are retrieved
	batch := queue.GetNextBatch()
	assert.Equal(t, 2, len(batch.Transactions))
	assert.Equal(t, tx1, batch.Transactions[0])
	assert.Equal(t, tx2, batch.Transactions[1])

	// Ensure the queue is empty after retrieving the batch
	emptyBatch := queue.GetNextBatch()
	assert.Equal(t, 0, len(emptyBatch.Transactions))
}

func TestDummySequencer_SubmitRollupTransaction(t *testing.T) {
	sequencer := NewDummySequencer()

	// Define a test rollup ID and transaction
	rollupId := []byte("test_rollup_id")
	tx := []byte("test_transaction")

	// Submit a transaction
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	resp, err := sequencer.SubmitRollupTransaction(context.Background(), req)

	// Assert no error and check if transaction was added
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Check that the transaction is in the queue
	batch := sequencer.tq.GetNextBatch()
	assert.Equal(t, 1, len(batch.Transactions))
	assert.Equal(t, tx, batch.Transactions[0])
}

func TestDummySequencer_GetNextBatch(t *testing.T) {
	sequencer := NewDummySequencer()

	// Add a transaction to the queue
	rollupId := []byte("test_rollup_id")
	tx := []byte("test_transaction")
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	sequencer.SubmitRollupTransaction(context.Background(), req)

	// Retrieve the next batch
	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: nil,
	}
	batchResp, err := sequencer.GetNextBatch(context.Background(), getBatchReq)

	// Assert no error
	assert.NoError(t, err)
	// Ensure the batch contains the transaction
	assert.NotNil(t, batchResp)
	assert.Equal(t, 1, len(batchResp.Batch.Transactions))
	assert.Equal(t, tx, batchResp.Batch.Transactions[0])

	// Assert timestamp is recent
	assert.WithinDuration(t, time.Now(), batchResp.Timestamp, time.Second)
}

func TestDummySequencer_GetNextBatch_LastBatchHashMismatch(t *testing.T) {
	sequencer := NewDummySequencer()

	// Submit a transaction
	rollupId := []byte("test_rollup_id")
	tx := []byte("test_transaction")
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	sequencer.SubmitRollupTransaction(context.Background(), req)

	// Retrieve the next batch
	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: []byte("invalid_hash"),
	}
	_, err := sequencer.GetNextBatch(context.Background(), getBatchReq)

	// Assert that the batch hash mismatch error is returned
	assert.Error(t, err)
	assert.Equal(t, "lastBatch is supposed to be nil", err.Error())
}

func TestDummySequencer_VerifyBatch(t *testing.T) {
	sequencer := NewDummySequencer()

	// Add and retrieve a batch
	rollupId := []byte("test_rollup_id")
	tx := []byte("test_transaction")
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	sequencer.SubmitRollupTransaction(context.Background(), req)

	// Get the next batch to generate batch hash
	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: nil,
	}
	batchResp, err := sequencer.GetNextBatch(context.Background(), getBatchReq)
	assert.NoError(t, err)

	batchHash, err := batchResp.Batch.Hash()
	assert.NoError(t, err)
	// Verify the batch using the batch hash
	verifyReq := sequencing.VerifyBatchRequest{
		BatchHash: batchHash,
	}
	verifyResp, err := sequencer.VerifyBatch(context.Background(), verifyReq)

	// Assert no error and that the batch is verified successfully
	assert.NoError(t, err)
	assert.True(t, verifyResp.Status)
}

func TestDummySequencer_VerifyBatch_NotFound(t *testing.T) {
	sequencer := NewDummySequencer()

	// Try verifying a batch with an invalid hash
	verifyReq := sequencing.VerifyBatchRequest{
		BatchHash: []byte("invalid_hash"),
	}
	verifyResp, err := sequencer.VerifyBatch(context.Background(), verifyReq)

	// Assert no error and that the batch is not found
	assert.NoError(t, err)
	assert.False(t, verifyResp.Status)
}
