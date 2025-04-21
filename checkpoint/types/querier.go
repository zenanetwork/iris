package types

// query endpoints supported by the auth Querier
const (
	QueryParams           = "params"
	QueryAckCount         = "ack-count"
	QueryCheckpoint       = "checkpoint"
	QueryCheckpointBuffer = "checkpoint-buffer"
	QueryLastNoAck        = "last-no-ack"
	QueryCheckpointList   = "checkpoint-list"
	QueryNextCheckpoint   = "next-checkpoint"
	QueryProposer         = "is-proposer"
	QueryCurrentProposer  = "current-proposer"
	StakingQuerierRoute   = "staking"
)

// QueryCheckpointParams defines the params for querying accounts.
type QueryCheckpointParams struct {
	Number uint64
}

// NewQueryCheckpointParams creates a new instance of QueryCheckpointHeaderIndex.
func NewQueryCheckpointParams(number uint64) QueryCheckpointParams {
	return QueryCheckpointParams{Number: number}
}

// QueryZenaChainID defines the params for querying with zena chain id
type QueryZenaChainID struct {
	ZenaChainID string
}

// NewQueryZenaChainID creates a new instance of QueryZenaChainID with give chain id
func NewQueryZenaChainID(chainID string) QueryZenaChainID {
	return QueryZenaChainID{ZenaChainID: chainID}
}
