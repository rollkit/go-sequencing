package test

import (
	"context"
	"testing"
	"time"

	"github.com/rollkit/go-sequencing"
	"github.com/stretchr/testify/assert"
)

func TestSubmitRollupTransaction(t *testing.T) {
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

	// Assert no error
	assert.NoError(t, err)
	// Assert the transaction was successfully added to the queue
	assert.Nil(t, resp)
	assert.Equal(t, rollupId, sequencer.RollupId)
}

func TestSubmitRollupTransaction_RollupIdMismatch(t *testing.T) {
	sequencer := NewDummySequencer()

	// Submit a transaction with one rollup ID
	rollupId1 := []byte("test_rollup_id1")
	tx1 := []byte("test_transaction_1")
	req1 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId1,
		Tx:       tx1,
	}
	sequencer.SubmitRollupTransaction(context.Background(), req1)

	// Submit a transaction with a different rollup ID (should cause an error)
	rollupId2 := []byte("test_rollup_id2")
	tx2 := []byte("test_transaction_2")
	req2 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId2,
		Tx:       tx2,
	}
	_, err := sequencer.SubmitRollupTransaction(context.Background(), req2)

	// Assert that the error is ErrorRollupIdMismatch
	assert.Error(t, err)
	assert.Equal(t, err, ErrorRollupIdMismatch)
}

func TestGetNextBatch(t *testing.T) {
	sequencer := NewDummySequencer()

	// Define a test rollup ID and transaction
	rollupId := []byte("test_rollup_id")
	tx := []byte("test_transaction")

	// Submit a transaction
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	sequencer.SubmitRollupTransaction(context.Background(), req)

	// Get the next batch
	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: nil,
		MaxBytes:      1024,
	}
	batchRes, err := sequencer.GetNextBatch(context.Background(), getBatchReq)

	// Assert no error
	assert.NoError(t, err)

	// Assert that the returned batch contains the correct transaction
	assert.NotNil(t, batchRes.Batch)
	assert.Equal(t, 1, len(batchRes.Batch.Transactions))
	assert.Equal(t, tx, batchRes.Batch.Transactions[0])

	// Assert timestamp is recent
	assert.WithinDuration(t, time.Now(), batchRes.Timestamp, time.Second)
}

func TestVerifyBatch(t *testing.T) {
	sequencer := NewDummySequencer()

	// Define a test rollup ID and transaction
	rollupId := []byte("test_rollup_id")
	tx := []byte("test_transaction")

	// Submit a transaction
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	sequencer.SubmitRollupTransaction(context.Background(), req)

	// Get the next batch to generate batch hash
	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: nil,
		MaxBytes:      1024,
	}
	batchRes, err := sequencer.GetNextBatch(context.Background(), getBatchReq)
	assert.NoError(t, err)

	batchHash, err := batchRes.Batch.Hash()
	assert.NoError(t, err)
	// Verify the batch
	verifyReq := sequencing.VerifyBatchRequest{
		RollupId:  rollupId,
		BatchHash: batchHash, // hash of the submitted transaction
	}
	verifyRes, err := sequencer.VerifyBatch(context.Background(), verifyReq)

	// Assert no error
	assert.NoError(t, err)

	// Assert that the batch was verified successfully
	assert.True(t, verifyRes.Status)
}
