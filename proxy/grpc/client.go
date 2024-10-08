package grpc

import (
	"context"

	"google.golang.org/grpc"

	types "github.com/cosmos/gogoproto/types"

	"github.com/rollkit/go-sequencing"
	pbseq "github.com/rollkit/go-sequencing/types/pb/sequencing"
)

// Client is a gRPC proxy client for DA interface.
type Client struct {
	conn *grpc.ClientConn

	pbseq.SequencerInputClient
	pbseq.SequencerOutputClient
	pbseq.BatchVerifierClient
}

// NewClient returns new Client instance.
func NewClient() *Client {
	return &Client{}
}

// Start connects Client to target, with given options.
func (c *Client) Start(target string, opts ...grpc.DialOption) (err error) {
	c.conn, err = grpc.NewClient(target, opts...)
	if err != nil {
		return err
	}

	c.SequencerInputClient = pbseq.NewSequencerInputClient(c.conn)
	c.SequencerOutputClient = pbseq.NewSequencerOutputClient(c.conn)
	c.BatchVerifierClient = pbseq.NewBatchVerifierClient(c.conn)

	return nil
}

// Stop gently closes Client connection.
func (c *Client) Stop() error {
	return c.conn.Close()
}

// SubmitRollupTransaction submits a transaction from rollup to sequencer.
func (c *Client) SubmitRollupTransaction(ctx context.Context, req sequencing.SubmitRollupTransactionRequest) (*sequencing.SubmitRollupTransactionResponse, error) {
	_, err := c.SequencerInputClient.SubmitRollupTransaction(ctx, &pbseq.SubmitRollupTransactionRequest{
		RollupId: req.RollupId,
		Data:     req.Tx,
	})
	if err != nil {
		return nil, err
	}
	return &sequencing.SubmitRollupTransactionResponse{}, nil
}

// GetNextBatch returns the next batch of transactions from sequencer to rollup.
func (c *Client) GetNextBatch(ctx context.Context, req sequencing.GetNextBatchRequest) (*sequencing.GetNextBatchResponse, error) {
	resp, err := c.SequencerOutputClient.GetNextBatch(ctx, &pbseq.GetNextBatchRequest{RollupId: req.RollupId, LastBatchHash: req.LastBatchHash})
	if err != nil {
		return nil, err
	}
	b := &sequencing.Batch{}
	b.FromProto(resp.Batch)
	t, err := types.TimestampFromProto(resp.Timestamp)
	if err != nil {
		return nil, err
	}
	return &sequencing.GetNextBatchResponse{Batch: b, Timestamp: t}, nil
}

// VerifyBatch verifies a batch of transactions received from the sequencer.
func (c *Client) VerifyBatch(ctx context.Context, req sequencing.VerifyBatchRequest) (*sequencing.VerifyBatchResponse, error) {
	resp, err := c.BatchVerifierClient.VerifyBatch(ctx, &pbseq.VerifyBatchRequest{RollupId: req.RollupId, BatchHash: req.BatchHash})
	if err != nil {
		return nil, err
	}
	return &sequencing.VerifyBatchResponse{Status: resp.Status}, nil
}

var _ sequencing.Sequencer = &Client{}
