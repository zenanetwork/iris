package types

import (
	"bytes"
	"math/big"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	hmCommon "github.com/zenanetwork/iris/common"
	"github.com/zenanetwork/iris/helper"
	"github.com/zenanetwork/iris/types"
)

//
// Checkpoint Msg
//

var _ sdk.Msg = &MsgCheckpoint{}

// MsgCheckpointAdjust represents checkpoint adjust
type MsgCheckpointAdjust struct {
	HeaderIndex uint64            `json:"header_index"`
	Proposer    types.IrisAddress `json:"proposer"`
	From        types.IrisAddress `json:"from"`
	StartBlock  uint64            `json:"start_block"`
	EndBlock    uint64            `json:"end_block"`
	RootHash    types.IrisHash    `json:"root_hash"`
}

// NewMsgCheckpointAdjust adjust previous checkpoint
func NewMsgCheckpointAdjust(
	headerIndex uint64,
	startBlock uint64,
	endBlock uint64,
	proposer types.IrisAddress,
	from types.IrisAddress,
	rootHash types.IrisHash,
) MsgCheckpointAdjust {
	return MsgCheckpointAdjust{
		HeaderIndex: headerIndex,
		StartBlock:  startBlock,
		EndBlock:    endBlock,
		Proposer:    proposer,
		From:        from,
		RootHash:    rootHash,
	}
}

func (msg MsgCheckpointAdjust) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// GetSigners returns address of the signer
func (msg MsgCheckpointAdjust) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.IrisAddressToAccAddress(msg.From)}
}

func (msg MsgCheckpointAdjust) Route() string {
	return RouterKey
}

// Type returns message type
func (msg MsgCheckpointAdjust) Type() string {
	return "checkpoint-adjust"
}

func (msg MsgCheckpointAdjust) ValidateBasic() sdk.Error {
	if bytes.Equal(msg.RootHash.Bytes(), helper.ZeroHash.Bytes()) {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid rootHash %v", msg.RootHash.String())
	}

	if msg.Proposer.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid proposer %v", msg.Proposer.String())
	}

	if msg.StartBlock >= msg.EndBlock || msg.EndBlock == 0 {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid startBlock %v or/and endBlock %v", msg.StartBlock, msg.EndBlock)
	}

	return nil
}

// GetSideSignBytes returns side sign bytes
func (msg MsgCheckpointAdjust) GetSideSignBytes() []byte {
	return nil
}

// MsgCheckpoint represents checkpoint
type MsgCheckpoint struct {
	Proposer        types.IrisAddress `json:"proposer"`
	StartBlock      uint64            `json:"start_block"`
	EndBlock        uint64            `json:"end_block"`
	RootHash        types.IrisHash    `json:"root_hash"`
	AccountRootHash types.IrisHash    `json:"account_root_hash"`
	ZenaChainID     string            `json:"zena_chain_id"`
}

// NewMsgCheckpointBlock creates new checkpoint message using mentioned arguments
func NewMsgCheckpointBlock(
	proposer types.IrisAddress,
	startBlock uint64,
	endBlock uint64,
	roothash types.IrisHash,
	accountRootHash types.IrisHash,
	zenaChainID string,
) MsgCheckpoint {
	return MsgCheckpoint{
		Proposer:        proposer,
		StartBlock:      startBlock,
		EndBlock:        endBlock,
		RootHash:        roothash,
		AccountRootHash: accountRootHash,
		ZenaChainID:     zenaChainID,
	}
}

// Type returns message type
func (msg MsgCheckpoint) Type() string {
	return "checkpoint"
}

func (msg MsgCheckpoint) Route() string {
	return RouterKey
}

// GetSigners returns address of the signer
func (msg MsgCheckpoint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.IrisAddressToAccAddress(msg.Proposer)}
}

func (msg MsgCheckpoint) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

func (msg MsgCheckpoint) ValidateBasic() sdk.Error {
	if bytes.Equal(msg.RootHash.Bytes(), helper.ZeroHash.Bytes()) {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid rootHash %v", msg.RootHash.String())
	}

	if msg.Proposer.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid proposer %v", msg.Proposer.String())
	}

	if msg.StartBlock >= msg.EndBlock || msg.EndBlock == 0 {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid startBlock %v or/and endBlock %v", msg.StartBlock, msg.EndBlock)
	}

	return nil
}

// GetSideSignBytes returns side sign bytes
func (msg MsgCheckpoint) GetSideSignBytes() []byte {
	// keccak256(abi.encoded(proposer, startBlock, endBlock, rootHash, accountRootHash, zena chain id))
	zenaChainID, _ := strconv.ParseUint(msg.ZenaChainID, 10, 64)

	return appendBytes32(
		msg.Proposer.Bytes(),
		new(big.Int).SetUint64(msg.StartBlock).Bytes(),
		new(big.Int).SetUint64(msg.EndBlock).Bytes(),
		msg.RootHash.Bytes(),
		msg.AccountRootHash.Bytes(),
		new(big.Int).SetUint64(zenaChainID).Bytes(),
	)
}

//
// Msg Checkpoint Ack
//

var _ sdk.Msg = &MsgCheckpointAck{}

// MsgCheckpointAck Add mainchain commit transaction hash to MsgCheckpointAck
type MsgCheckpointAck struct {
	From       types.IrisAddress `json:"from"`
	Number     uint64            `json:"number"`
	Proposer   types.IrisAddress `json:"proposer"`
	StartBlock uint64            `json:"start_block"`
	EndBlock   uint64            `json:"end_block"`
	RootHash   types.IrisHash    `json:"root_hash"`
	TxHash     types.IrisHash    `json:"tx_hash"`
	LogIndex   uint64            `json:"log_index"`
}

func NewMsgCheckpointAck(
	from types.IrisAddress,
	number uint64,
	proposer types.IrisAddress,
	startBlock uint64,
	endBlock uint64,
	rootHash types.IrisHash,
	txHash types.IrisHash,
	logIndex uint64,
) MsgCheckpointAck {
	return MsgCheckpointAck{
		From:       from,
		Number:     number,
		Proposer:   proposer,
		StartBlock: startBlock,
		EndBlock:   endBlock,
		RootHash:   rootHash,
		TxHash:     txHash,
		LogIndex:   logIndex,
	}
}

func (msg MsgCheckpointAck) Type() string {
	return "checkpoint-ack"
}

func (msg MsgCheckpointAck) Route() string {
	return RouterKey
}

// GetSigners returns signers
func (msg MsgCheckpointAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.IrisAddressToAccAddress(msg.From)}
}

// GetSignBytes returns sign bytes
func (msg MsgCheckpointAck) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// ValidateBasic validate basic
func (msg MsgCheckpointAck) ValidateBasic() sdk.Error {
	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid from %v", msg.From.String())
	}

	if msg.Proposer.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid empty proposer")
	}

	if msg.RootHash.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid empty root hash")
	}

	return nil
}

// GetTxHash Returns tx hash
func (msg MsgCheckpointAck) GetTxHash() types.IrisHash {
	return msg.TxHash
}

// GetLogIndex Returns log index
func (msg MsgCheckpointAck) GetLogIndex() uint64 {
	return msg.LogIndex
}

// GetSideSignBytes returns side sign bytes
func (msg MsgCheckpointAck) GetSideSignBytes() []byte {
	return nil
}

//
// Msg Checkpoint No Ack
//

var _ sdk.Msg = &MsgCheckpointNoAck{}

type MsgCheckpointNoAck struct {
	From types.IrisAddress `json:"from"`
}

func NewMsgCheckpointNoAck(from types.IrisAddress) MsgCheckpointNoAck {
	return MsgCheckpointNoAck{
		From: from,
	}
}

func (msg MsgCheckpointNoAck) Type() string {
	return "checkpoint-no-ack"
}

func (msg MsgCheckpointNoAck) Route() string {
	return RouterKey
}

func (msg MsgCheckpointNoAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.IrisAddressToAccAddress(msg.From)}
}

func (msg MsgCheckpointNoAck) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

func (msg MsgCheckpointNoAck) ValidateBasic() sdk.Error {
	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid from %v", msg.From.String())
	}

	return nil
}
