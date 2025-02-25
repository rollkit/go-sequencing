package sequencing

import (
	"crypto/sha256"

	pbseq "github.com/rollkit/go-sequencing/types/pb/sequencing"
)

// ToProto serializes a batch to a protobuf message.
func (batch *Batch) ToProto() *pbseq.Batch {
	return &pbseq.Batch{Transactions: batch.Transactions}
}

// FromProto deserializes a batch from a protobuf message.
func (batch *Batch) FromProto(pb *pbseq.Batch) {
	batch.Transactions = pb.Transactions
}

// Marshal serializes a batch to a byte slice.
func (batch *Batch) Marshal() ([]byte, error) {
	return batch.ToProto().Marshal()
}

// Unmarshal deserializes a batch from a byte slice.
func (batch *Batch) Unmarshal(data []byte) error {
	var pb pbseq.Batch
	if err := pb.Unmarshal(data); err != nil {
		return err
	}
	batch.FromProto(&pb)
	return nil
}

// Hash returns the hash of a batch.
func (batch *Batch) Hash() ([]byte, error) {
	batchBytes, err := batch.Marshal()
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(batchBytes)
	return hash[:], nil
}
