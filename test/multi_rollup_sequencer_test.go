package test

import (
	"context"
	"math"
	"testing"

	"github.com/rollkit/go-sequencing"
	"github.com/stretchr/testify/assert"
)

func TestMultiRollupSequencer_SubmitRollupTransaction(t *testing.T) {
	sequencer := NewMultiRollupSequencer()

	rollupId := []byte("test-rollup")
	tx := []byte("transaction data")

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
	tx := []byte("transaction data")

	// Submit the transaction
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	_, err := sequencer.SubmitRollupTransaction(context.Background(), req)
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
	tx := []byte("transaction data")

	// Submit the transaction
	req := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId,
		Tx:       tx,
	}
	_, err := sequencer.SubmitRollupTransaction(context.Background(), req)
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
	tx1 := []byte("tx data 1")
	tx2 := []byte("tx data 2")

	// Submit transactions for two different rollups
	req1 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId1,
		Tx:       tx1,
	}
	req2 := sequencing.SubmitRollupTransactionRequest{
		RollupId: rollupId2,
		Tx:       tx2,
	}

	_, err := sequencer.SubmitRollupTransaction(context.Background(), req1)
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
