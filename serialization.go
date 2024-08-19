package sequencing

import (
	pbseq "github.com/rollkit/go-sequencing/types/pb/sequencing"
)

// ToProto serializes a batch to a protobuf message.
func (batch *Batch) ToProto() *pbseq.Batch {
	var txs [][]byte
	if batch != nil {
		txs = batch.Transactions
	}
	return &pbseq.Batch{Transactions: txsToByteSlices(txs)}
}

// FromProto deserializes a batch from a protobuf message.
func (batch *Batch) FromProto(pb *pbseq.Batch) {
	batch.Transactions = byteSlicesToTxs(pb.Transactions)
}

func txsToByteSlices(txs []Tx) [][]byte {
	if txs == nil {
		return nil
	}
	bytes := make([][]byte, len(txs))
	copy(bytes, txs)
	return bytes
}

func byteSlicesToTxs(bytes [][]byte) []Tx {
	if len(bytes) == 0 {
		return nil
	}
	txs := make([]Tx, len(bytes))
	copy(txs, bytes)
	return txs
}

// Marshal serializes a batch to a byte slice.
func (batch *Batch) Marshal() ([]byte, error) {
	var b []byte
	if batch == nil {
		return b, nil
	}
	pb := batch.ToProto()
	return pb.Marshal()
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
