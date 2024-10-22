package test

import (
	"context"
	"crypto/rand"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rollkit/go-sequencing"
)

func TestMultiRollupSequencer_SubmitRollupTransaction(t *testing.T) {
	sequencer := NewMultiRollupSequencer()

	rollupId := []byte("test-rollup")
	tx, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)

	// Submit the transaction
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}

	res, err := sequencer.SubmitRollupTransaction(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Ensure that the transaction was added to the transaction queue for the rollup
	rollup, _ := sequencer.getOrCreateRollup(rollupId)
	nextBatch := rollup.tq.GetNextBatch(math.MaxInt32)

	assert.Equal(t, 1, len(nextBatch.Transactions))
	assert.Equal(t, tx, nextBatch.Transactions[0])
}

func TestMultiRollupSequencer_GetNextBatch(t *testing.T) {
	sequencer := NewMultiRollupSequencer()

	rollupId := []byte("test-rollup")
	tx, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)

	// Submit the transaction
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req)
	assert.NoError(t, err)

	// Get next batch
	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: nil,
		MaxBytes:      math.MaxInt32,
	}
	batchRes, err := sequencer.GetNextBatch(context.Background(), getBatchReq)
	assert.NoError(t, err)

	// Verify that the batch contains the transaction
	assert.NotNil(t, batchRes.Batch)
	assert.Equal(t, 1, len(batchRes.Batch.Transactions))
	assert.Equal(t, tx, batchRes.Batch.Transactions[0])
}

func TestMultiRollupSequencer_VerifyBatch(t *testing.T) {
	sequencer := NewMultiRollupSequencer()

	rollupId := []byte("test-rollup")
	tx, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)

	// Submit the transaction
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	_, err = sequencer.SubmitRollupTransaction(context.Background(), req)
	assert.NoError(t, err)

	// Get the next batch to update the last batch hash
	getBatchReq := sequencing.GetNextBatchRequest{
		RollupId:      rollupId,
		LastBatchHash: nil,
		MaxBytes:      math.MaxInt32,
	}
	batchRes, err := sequencer.GetNextBatch(context.Background(), getBatchReq)
	assert.NoError(t, err)

	bHash, err := batchRes.Batch.Hash()
	assert.NoError(t, err)

	// Verify the batch
	verifyReq := sequencing.VerifyBatchRequest{
		RollupId:  rollupId,
		BatchHash: bHash,
	}

	verifyRes, err := sequencer.VerifyBatch(context.Background(), verifyReq)
	assert.NoError(t, err)
	assert.True(t, verifyRes.Status)
}

func TestMultiRollupSequencer_MultipleRollups(t *testing.T) {
	sequencer := NewMultiRollupSequencer()

	rollupId1 := []byte("rollup-1")
	rollupId2 := []byte("rollup-2")
	tx1, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)
	tx2, err := GenerateSecureRandomBytes(32)
	assert.NoError(t, err)

	// Submit transactions for two different rollups
	req1 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId1,
		Tx:       tx1,
	}
	req2 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId2,
		Tx:       tx2,
	}

	_, err = sequencer.SubmitRollupTransaction(context.Background(), req1)
	assert.NoError(t, err)

	_, err = sequencer.SubmitRollupTransaction(context.Background(), req2)
	assert.NoError(t, err)

	// Get next batch for rollup 1
	getBatchReq1 := sequencing.GetNextBatchRequest{
		RollupId:      rollupId1,
		LastBatchHash: nil,
		MaxBytes:      math.MaxInt32,
	}
	batchRes1, err := sequencer.GetNextBatch(context.Background(), getBatchReq1)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(batchRes1.Batch.Transactions))
	assert.Equal(t, tx1, batchRes1.Batch.Transactions[0])

	// Get next batch for rollup 2
	getBatchReq2 := sequencing.GetNextBatchRequest{
		RollupId:      rollupId2,
		LastBatchHash: nil,
		MaxBytes:      math.MaxInt32,
	}
	batchRes2, err := sequencer.GetNextBatch(context.Background(), getBatchReq2)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(batchRes2.Batch.Transactions))
	assert.Equal(t, tx2, batchRes2.Batch.Transactions[0])
}

// TestMultiRollupSequencer tests the MultiRollupSequencer core functionality of
// submitting transactions, getting the next batch, and verifying a batch
func TestMultiRollupSequencer(t *testing.T) {
	// Test with 1 rollup
	testMultiRollupSequencer(t, 1)

	// Test with 2 to 5 rollups
	r, err := rand.Int(rand.Reader, big.NewInt(4)) // Generates a number between 0 and 3
	assert.NoError(t, err)

	numRollups := int(r.Int64() + 2) // Adjust range to be between 2 and 5, cast to int
	testMultiRollupSequencer(t, numRollups)
}

func testMultiRollupSequencer(t *testing.T, numRollups int) {
	sequencer := NewMultiRollupSequencer()

	for i := 0; i < numRollups; i++ {
		// Submit a transaction
		rollupId, err := GenerateSecureRandomBytes(10)
		assert.NoError(t, err)
		tx, err := GenerateSecureRandomBytes(30)
		assert.NoError(t, err)
		req := sequencing.SubmitRollupTransactionRequest{
			RollupId: rollupId,
			Tx:       tx,
		}
		_, err = sequencer.SubmitRollupTransaction(context.Background(), req)
		assert.NoError(t, err)

		// Get Next Batch
		getBatchReq := sequencing.GetNextBatchRequest{
			RollupId:      rollupId,
			LastBatchHash: nil,
			MaxBytes:      math.MaxInt32,
		}
		batchRes, err := sequencer.GetNextBatch(context.Background(), getBatchReq)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(batchRes.Batch.Transactions))
		assert.Equal(t, tx, batchRes.Batch.Transactions[0])

		// Verify the batch
		bHash, err := batchRes.Batch.Hash()
		assert.NoError(t, err)
		verifyReq := sequencing.VerifyBatchRequest{
			RollupId:  rollupId,
			BatchHash: bHash,
		}
		verifyRes, err := sequencer.VerifyBatch(context.Background(), verifyReq)
		assert.NoError(t, err)
		assert.True(t, verifyRes.Status)
	}
}
