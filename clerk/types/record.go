package types

import (
	"fmt"
	"time"

	"github.com/zenanetwork/iris/types"
)

// EventRecord represents state record
type EventRecord struct {
	ID         uint64            `json:"id" yaml:"id"`
	Contract   types.IrisAddress `json:"contract" yaml:"contract"`
	Data       types.HexBytes    `json:"data" yaml:"data"`
	TxHash     types.IrisHash    `json:"tx_hash" yaml:"tx_hash"`
	LogIndex   uint64            `json:"log_index" yaml:"log_index"`
	ChainID    string            `json:"zena_chain_id" yaml:"zena_chain_id"`
	RecordTime time.Time         `json:"record_time" yaml:"record_time"`
}

// NewEventRecord creates new record
func NewEventRecord(
	txHash types.IrisHash,
	logIndex uint64,
	id uint64,
	contract types.IrisAddress,
	data types.HexBytes,
	chainID string,
	recordTime time.Time,
) EventRecord {
	return EventRecord{
		ID:         id,
		Contract:   contract,
		Data:       data,
		TxHash:     txHash,
		LogIndex:   logIndex,
		ChainID:    chainID,
		RecordTime: recordTime,
	}
}

// String returns the string representation of span
func (s *EventRecord) String() string {
	return fmt.Sprintf(
		"EventRecord: id %v, contract %v, data: %v, txHash: %v, logIndex: %v, chainId: %v, recordTime: %v",
		s.ID,
		s.Contract.String(),
		s.Data.String(),
		s.TxHash.Hex(),
		s.LogIndex,
		s.ChainID,
		s.RecordTime,
	)
}
