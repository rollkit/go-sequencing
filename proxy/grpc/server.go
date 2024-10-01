package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/rollkit/go-sequencing"

	types "github.com/cosmos/gogoproto/types"

	pbseq "github.com/rollkit/go-sequencing/types/pb/sequencing"
)

// NewServer creates a new gRPC server for the Sequencer service.
func NewServer(
	input sequencing.SequencerInput,
	output sequencing.SequencerOutput,
	verifier sequencing.BatchVerifier,
	opts ...grpc.ServerOption,
) *grpc.Server {
	srv := grpc.NewServer(opts...)

	proxyInputSrv := &proxyInputSrv{SequencerInput: input}
	proxyOutputSrv := &proxyOutputSrv{SequencerOutput: output}
	proxyVerificationSrv := &proxyVerificationSrv{BatchVerifier: verifier}

	pbseq.RegisterSequencerInputServer(srv, proxyInputSrv)
	pbseq.RegisterSequencerOutputServer(srv, proxyOutputSrv)
	pbseq.RegisterBatchVerifierServer(srv, proxyVerificationSrv)

	return srv
}

type proxyInputSrv struct {
	sequencing.SequencerInput
}

type proxyOutputSrv struct {
	sequencing.SequencerOutput
}

type proxyVerificationSrv struct {
	sequencing.BatchVerifier
}

// SubmitRollupTransaction submits a transaction from rollup to sequencer.
func (s *proxyInputSrv) SubmitRollupTransaction(ctx context.Context, req *pbseq.SubmitRollupTransactionRequest) (*pbseq.SubmitRollupTransactionResponse, error) {
	_, err := s.SequencerInput.SubmitRollupTransaction(ctx, sequencing.SubmitRollupTransactionRequest{RollupId: req.RollupId, Tx: req.Data})
	if err != nil {
		return nil, err
	}
	return &pbseq.SubmitRollupTransactionResponse{}, nil
}

// GetNextBatch returns the next batch of transactions from sequencer to rollup.
func (s *proxyOutputSrv) GetNextBatch(ctx context.Context, req *pbseq.GetNextBatchRequest) (*pbseq.GetNextBatchResponse, error) {
	resp, err := s.SequencerOutput.GetNextBatch(ctx, sequencing.GetNextBatchRequest{RollupId: req.RollupId, LastBatchHash: req.LastBatchHash})
	if err != nil {
		return nil, err
	}
	ts, err := types.TimestampProto(resp.Timestamp)
	if err != nil {
		return nil, err
	}
	return &pbseq.GetNextBatchResponse{Batch: resp.Batch.ToProto(), Timestamp: ts}, nil
}

// VerifyBatch verifies a batch of transactions received from the sequencer.
func (s *proxyVerificationSrv) VerifyBatch(ctx context.Context, req *pbseq.VerifyBatchRequest) (*pbseq.VerifyBatchResponse, error) {
	resp, err := s.BatchVerifier.VerifyBatch(ctx, sequencing.VerifyBatchRequest{BatchHash: req.BatchHash})
	if err != nil {
		return nil, err
	}
	return &pbseq.VerifyBatchResponse{Status: resp.Status}, nil
}
