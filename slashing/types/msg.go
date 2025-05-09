package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zenanetwork/go-zenanet/accounts/abi"
	hmCommon "github.com/zenanetwork/iris/common"
	"github.com/zenanetwork/iris/types"
	hmTypes "github.com/zenanetwork/iris/types"
)

// verify interface at compile time
var _ sdk.Msg = &MsgUnjail{}

// MsgUnjail - struct for unjailing jailed validator
type MsgUnjail struct {
	From        types.IrisAddress   `json:"from"`
	ID          hmTypes.ValidatorID `json:"id"`
	TxHash      types.IrisHash      `json:"tx_hash"`
	LogIndex    uint64              `json:"log_index"`
	BlockNumber uint64              `json:"block_number"`
}

func NewMsgUnjail(from types.IrisAddress, id uint64, txHash types.IrisHash, logIndex uint64, blockNumber uint64) MsgUnjail {
	return MsgUnjail{
		From:        from,
		ID:          hmTypes.NewValidatorID(id),
		TxHash:      txHash,
		LogIndex:    logIndex,
		BlockNumber: blockNumber,
	}
}

// nolint
func (msg MsgUnjail) Route() string { return RouterKey }
func (msg MsgUnjail) Type() string  { return "unjail" }
func (msg MsgUnjail) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{hmTypes.IrisAddressToAccAddress(msg.From)}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgUnjail) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgUnjail) ValidateBasic() sdk.Error {
	if msg.ID <= 0 {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid validator ID %v", msg.ID)
	}

	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid proposer %v", msg.From.String())
	}
	return nil
}

// Tick Msg

// TickMsg - struct for unjailing jailed validator
type MsgTick struct {
	ID                uint64            `json:"id"`
	Proposer          types.IrisAddress `json:"proposer"`
	SlashingInfoBytes types.HexBytes    `json:"slashinginfobytes"`
}

func NewMsgTick(id uint64, proposer types.IrisAddress, slashingInfoBytes types.HexBytes) MsgTick {
	return MsgTick{
		ID:                id,
		Proposer:          proposer,
		SlashingInfoBytes: slashingInfoBytes,
	}
}

// Type returns message type
func (msg MsgTick) Type() string {
	return "tick"
}

func (msg MsgTick) Route() string {
	return RouterKey
}

// GetSigners returns address of the signer
func (msg MsgTick) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.IrisAddressToAccAddress(msg.Proposer)}
}

func (msg MsgTick) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgTick) ValidateBasic() sdk.Error {

	if msg.Proposer.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid proposer %v", msg.Proposer.String())
	}
	return nil
}

// GetSideSignBytes returns side sign bytes
func (msg MsgTick) GetSideSignBytes() []byte {
	uintType, _ := abi.NewType("uint256", "", nil)
	addressType, _ := abi.NewType("address", "", nil)
	bytesType, _ := abi.NewType("bytes", "", nil)

	arguments := abi.Arguments{
		{
			Type: uintType,
		},
		{
			Type: addressType,
		},
		{
			Type: bytesType,
		},
	}

	bytes, _ := arguments.Pack(
		new(big.Int).SetUint64(msg.ID),
		msg.Proposer,
		msg.SlashingInfoBytes,
	)
	return bytes
}

//
// Msg Tick Ack
//

var _ sdk.Msg = &MsgTickAck{}

type MsgTickAck struct {
	From          types.IrisAddress `json:"from"`
	ID            uint64            `json:"tick_id"`
	SlashedAmount uint64            `json:"slashed_amount"`
	TxHash        types.IrisHash    `json:"tx_hash"`
	LogIndex      uint64            `json:"log_index"`
	BlockNumber   uint64            `json:"block_number"`
}

func NewMsgTickAck(from types.IrisAddress, id uint64, slashedAmount uint64, txHash types.IrisHash, logIndex uint64, blockNumber uint64) MsgTickAck {
	return MsgTickAck{
		From:          from,
		ID:            id,
		SlashedAmount: slashedAmount,
		TxHash:        txHash,
		BlockNumber:   blockNumber,
		LogIndex:      logIndex,
	}
}

// Type returns message type
func (msg MsgTickAck) Type() string {
	return "tick-ack"
}

func (msg MsgTickAck) Route() string {
	return RouterKey
}

// GetSigners returns address of the signer
func (msg MsgTickAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.IrisAddressToAccAddress(msg.From)}
}

func (msg MsgTickAck) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgTickAck) ValidateBasic() sdk.Error {
	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid from %v", msg.From.String())
	}
	return nil
}

// GetTxHash Returns tx hash
func (msg MsgTickAck) GetTxHash() types.IrisHash {
	return msg.TxHash
}

// GetLogIndex Returns log index
func (msg MsgTickAck) GetLogIndex() uint64 {
	return msg.LogIndex
}

// GetSideSignBytes returns side sign bytes
func (msg MsgTickAck) GetSideSignBytes() []byte {
	return nil
}
