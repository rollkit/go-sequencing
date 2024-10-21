package test

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rollkit/go-sequencing"
)

func TestTransactionQueue_AddTransaction(t *testing.T) {
	queue := NewTransactionQueue()

	tx1, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	queue.AddTransaction(tx1)

	// Check that the transaction was added
	batch := queue.GetNextBatch(math.MaxUint64)
	assert.Equal(t, 1, len(batch.Transactions))
	assert.Equal(t, tx1, batch.Transactions[0])
}

func TestTransactionQueue_GetNextBatch(t *testing.T) {
	queue := NewTransactionQueue()

	// Add multiple transactions
	tx1, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	tx2, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	queue.AddTransaction(tx1)
	queue.AddTransaction(tx2)

	// Get next batch and check if all transactions are retrieved
	batch := queue.GetNextBatch(math.MaxUint64)
	assert.Equal(t, 2, len(batch.Transactions))
	assert.Equal(t, tx1, batch.Transactions[0])
	assert.Equal(t, tx2, batch.Transactions[1])

	// Ensure the queue is empty after retrieving the batch
	emptyBatch := queue.GetNextBatch(math.MaxUint64)
	assert.Equal(t, 0, len(emptyBatch.Transactions))
}

func TestDummySequencer_SubmitRollupTransaction(t *testing.T) {
	// Define a test rollup ID and transaction
	rollupId := []byte("test_rollup_id")
	tx, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	sequencer := NewDummySequencer(rollupId)

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
	batch := sequencer.tq.GetNextBatch(math.MaxUint64)
	assert.Equal(t, 1, len(batch.Transactions))
	assert.Equal(t, tx, batch.Transactions[0])
}

func TestDummySequencer_SubmitEmptyTransaction(t *testing.T) {
	// Define a test rollup ID and an empty transaction
	rollupId := []byte("test_rollup_id")
	tx := []byte("")
	sequencer := NewDummySequencer(rollupId)

	// Submit an empty transaction
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	resp, err := sequencer.SubmitRollupTransaction(context.Background(), req)

	// Assert no error (since the sequencer may accept empty transactions) and ensure the transaction was added
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Check that the transaction is in the queue (even if empty)
	batch := sequencer.tq.GetNextBatch(math.MaxUint64)
	assert.Equal(t, 1, len(batch.Transactions))
	assert.Equal(t, tx, batch.Transactions[0])
}

func TestDummySequencer_SubmitMultipleTransactions(t *testing.T) {
	// Define a test rollup ID and multiple transactions
	rollupId := []byte("test_rollup_id")
	tx1, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	tx2, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	tx3, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	sequencer := NewDummySequencer(rollupId)

	// Submit multiple transactions
	req1 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx1,
	}
	req2 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx2,
	}
	req3 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx3,
	}

	_, err = sequencer.SubmitRollupTransaction(context.Background(), req1)
	assert.NoError(t, err)
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req2)
	assert.NoError(t, err)
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req3)
	assert.NoError(t, err)

	// Check that all transactions are added to the queue
	batch := sequencer.tq.GetNextBatch(math.MaxUint64)
	assert.Equal(t, 3, len(batch.Transactions))
	assert.Equal(t, tx1, batch.Transactions[0])
	assert.Equal(t, tx2, batch.Transactions[1])
	assert.Equal(t, tx3, batch.Transactions[2])
}

func TestDummySequencer_GetNextBatch(t *testing.T) {
	// Add a transaction to the queue
	rollupId := []byte("test_rollup_id")
	tx, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	sequencer := NewDummySequencer(rollupId)
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req)
	assert.NoError(t, err)

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

// Test retrieving a batch with no transactions
func TestDummySequencer_GetNextBatch_NoTransactions(t *testing.T) {
	rollupId := []byte("test_rollup_id")
	sequencer := NewDummySequencer(rollupId)
	// Attempt to retrieve a batch with no transactions in the queue
	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: nil,
	}
	batchResp, err := sequencer.GetNextBatch(context.Background(), getBatchReq)

	// Assert no error
	assert.NoError(t, err)
	// Assert that the batch is empty
	assert.NotNil(t, batchResp)
	assert.Equal(t, 0, len(batchResp.Batch.Transactions))
}

func TestDummySequencer_GetNextBatch_LastBatchHashMismatch(t *testing.T) {
	// Submit a transaction
	rollupId := []byte("test_rollup_id")
	sequencer := NewDummySequencer(rollupId)
	tx, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req)
	assert.NoError(t, err)

	// Retrieve the next batch
	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: []byte("invalid_hash"),
	}
	_, err = sequencer.GetNextBatch(context.Background(), getBatchReq)

	// Assert that the batch hash mismatch error is returned
	assert.Error(t, err)
	assert.ErrorContains(t, err, "batch hash mismatch", "unexpected error message")
}

// Test retrieving a batch with maxBytes limit
func TestDummySequencer_GetNextBatch_MaxBytesLimit(t *testing.T) {
	// Define a test rollup ID and multiple transactions
	rollupId := []byte("test_rollup_id")
	sequencer := NewDummySequencer(rollupId)
	tx1, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	tx2, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	tx3, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)

	// Submit multiple transactions
	req1 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx1,
	}
	req2 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx2,
	}
	req3 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx3,
	}

	_, err = sequencer.SubmitRollupTransaction(context.Background(), req1)
	assert.NoError(t, err)
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req2)
	assert.NoError(t, err)
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req3)
	assert.NoError(t, err)

	// Retrieve the next batch with maxBytes limit (assuming maxBytes limits the size of transactions returned)
	maxBytes := uint64(len(tx1) + len(tx2)) // Set maxBytes to the size of tx1 + tx2

	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: nil,
		MaxBytes:      maxBytes,
	}
	batchResp, err := sequencer.GetNextBatch(context.Background(), getBatchReq)

	// Assert no error
	assert.NoError(t, err)
	// Assert that only two transactions (tx1 and tx2) are retrieved due to the maxBytes limit
	assert.NotNil(t, batchResp)
	assert.Equal(t, 2, len(batchResp.Batch.Transactions))
	assert.Equal(t, tx1, batchResp.Batch.Transactions[0])
	assert.Equal(t, tx2, batchResp.Batch.Transactions[1])

	lastBatchHash, err := batchResp.Batch.Hash()
	assert.NoError(t, err)
	// Retrieve the remaining transaction (tx3) with another call
	getBatchReq2 := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: lastBatchHash,
		MaxBytes:      maxBytes, // This can be set higher to retrieve all remaining transactions
	}
	batchResp2, err2 := sequencer.GetNextBatch(context.Background(), getBatchReq2)

	// Assert no error and tx3 is retrieved
	assert.NoError(t, err2)
	assert.NotNil(t, batchResp2)
	assert.Equal(t, 1, len(batchResp2.Batch.Transactions))
	assert.Equal(t, tx3, batchResp2.Batch.Transactions[0])
}

func TestDummySequencer_VerifyBatch(t *testing.T) {
	// Add and retrieve a batch
	rollupId := []byte("test_rollup_id")
	sequencer := NewDummySequencer(rollupId)
	tx, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req)
	assert.NoError(t, err)

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
		RollupId:  rollupId,
		BatchHash: batchHash,
	}
	verifyResp, err := sequencer.VerifyBatch(context.Background(), verifyReq)

	// Assert no error and that the batch is verified successfully
	assert.NoError(t, err)
	assert.True(t, verifyResp.Status)
}

func TestDummySequencer_VerifyEmptyBatch(t *testing.T) {
	rollupId := []byte("test_rollup_id")
	sequencer := NewDummySequencer(rollupId)

	// Create a request with an empty batch hash
	emptyBatchHash := []byte{}
	verifyReq := sequencing.VerifyBatchRequest{
		RollupId:  rollupId,
		BatchHash: emptyBatchHash,
	}

	// Attempt to verify the empty batch
	verifyResp, err := sequencer.VerifyBatch(context.Background(), verifyReq)

	// Assert that the batch is not verified (as it's empty)
	assert.NoError(t, err)
	assert.False(t, verifyResp.Status)
}

func TestDummySequencer_VerifyBatchWithMultipleTransactions(t *testing.T) {
	// Define a test rollup ID and multiple transactions
	rollupId := []byte("test_rollup_id")
	sequencer := NewDummySequencer(rollupId)
	tx1, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	tx2, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)

	// Submit multiple transactions
	req1 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx1,
	}
	req2 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx2,
	}

	_, err = sequencer.SubmitRollupTransaction(context.Background(), req1)
	assert.NoError(t, err)
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req2)
	assert.NoError(t, err)

	// Get the next batch to generate the batch hash
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
		RollupId:  rollupId,
		BatchHash: batchHash,
	}
	verifyResp, err := sequencer.VerifyBatch(context.Background(), verifyReq)

	// Assert that the batch with multiple transactions is verified successfully
	assert.NoError(t, err)
	assert.True(t, verifyResp.Status)
}

func TestDummySequencer_VerifyBatch_NotFound(t *testing.T) {
	rollupId := []byte("test_rollup_id")
	sequencer := NewDummySequencer(rollupId)

	// Try verifying a batch with an invalid hash
	verifyReq := sequencing.VerifyBatchRequest{
		RollupId:  rollupId,
		BatchHash: []byte("invalid_hash"),
	}
	verifyResp, err := sequencer.VerifyBatch(context.Background(), verifyReq)

	// Assert no error and that the batch is not found
	assert.NoError(t, err)
	assert.False(t, verifyResp.Status)
}

// GenerateSecureRandomBytes generates cryptographically secure random bytes of the given length.
func GenerateSecureRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("invalid length: %d, must be greater than 0", length)
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return buf, nil
}
