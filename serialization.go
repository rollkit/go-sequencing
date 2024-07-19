package sequencing

import (
	pbseq "github.com/rollkit/go-sequencing/types/pb/sequencing"
)

func (batch *Batch) ToProto() *pbseq.Batch {
	return &pbseq.Batch{Transactions: txsToByteSlices(batch.Transactions)}
}

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

func (batch *Batch) Marshal() ([]byte, error) {
	pb := batch.ToProto()
	return pb.Marshal()
}

func (batch *Batch) Unmarshal(data []byte) error {
	var pb pbseq.Batch
	if err := pb.Unmarshal(data); err != nil {
		return err
	}
	batch.FromProto(&pb)
	return nil
}
