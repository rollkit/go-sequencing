package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/rollkit/go-sequencing"

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
	err := s.SequencerInput.SubmitRollupTransaction(ctx, req.RollupId, req.Data)
	if err != nil {
		return nil, err
	}
	return &pbseq.SubmitRollupTransactionResponse{}, nil
}

// GetNextBatch returns the next batch of transactions from sequencer to rollup.
func (s *proxyOutputSrv) GetNextBatch(ctx context.Context, req *pbseq.Batch) (*pbseq.Batch, error) {
	lastBatch := sequencing.Batch{}
	lastBatch.FromProto(req)
	batch, err := s.SequencerOutput.GetNextBatch(ctx, lastBatch)
	if err != nil {
		return nil, err
	}
	return batch.ToProto(), nil
}

// VerifyBatch verifies a batch of transactions received from the sequencer.
func (s *proxyVerificationSrv) VerifyBatch(ctx context.Context, req *pbseq.Batch) (*pbseq.VerificationResponse, error) {
	batch := sequencing.Batch{}
	batch.FromProto(req)
	ok, err := s.BatchVerifier.VerifyBatch(ctx, batch)
	if err != nil {
		return nil, err
	}
	return &pbseq.VerificationResponse{Success: ok}, nil
}
