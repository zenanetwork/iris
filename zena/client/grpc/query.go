package grpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	protoutil "github.com/zenanetwork/zenaproto/utils"
	proto "github.com/zenanetwork/zenaproto/zena"
)

func (h *ZenaGRPCClient) GetRootHash(ctx context.Context, startBlock uint64, endBlock uint64) (string, error) {

	req := &proto.GetRootHashRequest{
		StartBlockNumber: startBlock,
		EndBlockNumber:   endBlock,
	}

	log.Info("Fetching zena root hash")

	res, err := h.client.GetRootHash(ctx, req)
	if err != nil {
		return "", err
	}

	log.Info("Fetched zena root hash")

	return res.RootHash, nil
}

func (h *ZenaGRPCClient) GetVoteOnHash(ctx context.Context, startBlock uint64, endBlock uint64, rootHash string, milestoneId string) (bool, error) {

	req := &proto.GetVoteOnHashRequest{
		StartBlockNumber: startBlock,
		EndBlockNumber:   endBlock,
		Hash:             rootHash,
		MilestoneId:      milestoneId,
	}

	log.Info("Fetching vote on hash")

	res, err := h.client.GetVoteOnHash(ctx, req)
	if err != nil {
		return false, err
	}

	log.Info("Fetched vote on hash")

	return res.Response, nil
}

func (h *ZenaGRPCClient) HeaderByNumber(ctx context.Context, blockID int64) (*ethTypes.Header, error) {

	if blockID > math.MaxInt64 {
		return nil, fmt.Errorf("blockID too large: %d", blockID)
	}

	blockNumberAsString := ToBlockNumArg(big.NewInt(blockID))

	req := &proto.GetHeaderByNumberRequest{
		Number: blockNumberAsString,
	}

	log.Info("Fetching header by number")

	res, err := h.client.HeaderByNumber(ctx, req)
	if err != nil {
		return &ethTypes.Header{}, err
	}

	log.Info("Fetched header by number")

	resp := &ethTypes.Header{
		Number:     big.NewInt(0).SetUint64(res.Header.Number),
		ParentHash: protoutil.ConvertH256ToHash(res.Header.ParentHash),
		Time:       res.Header.Time,
	}

	return resp, nil
}

func (h *ZenaGRPCClient) BlockByNumber(ctx context.Context, blockID int64) (*ethTypes.Block, error) {

	if blockID > math.MaxInt64 {
		return nil, fmt.Errorf("blockID too large: %d", blockID)
	}

	blockNumberAsString := ToBlockNumArg(big.NewInt(blockID))

	req := &proto.GetBlockByNumberRequest{
		Number: blockNumberAsString,
	}

	log.Info("Fetching block by number")

	res, err := h.client.BlockByNumber(ctx, req)
	if err != nil {
		return &ethTypes.Block{}, err
	}

	log.Info("Fetched block by number")

	header := ethTypes.Header{
		Number:     big.NewInt(0).SetUint64(res.Block.Header.Number),
		ParentHash: protoutil.ConvertH256ToHash(res.Block.Header.ParentHash),
		Time:       res.Block.Header.Time,
	}
	return ethTypes.NewBlock(&header, nil, nil, nil), nil
}

func (h *ZenaGRPCClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethTypes.Receipt, error) {

	req := &proto.ReceiptRequest{
		Hash: protoutil.ConvertHashToH256(txHash),
	}

	log.Info("Fetching transaction receipt")

	res, err := h.client.TransactionReceipt(ctx, req)
	if err != nil {
		return &ethTypes.Receipt{}, err
	}

	log.Info("Fetched transaction receipt")

	return receiptResponseToTypesReceipt(res.Receipt), nil
}

func (h *ZenaGRPCClient) ZenaBlockReceipt(ctx context.Context, txHash common.Hash) (*ethTypes.Receipt, error) {

	req := &proto.ReceiptRequest{
		Hash: protoutil.ConvertHashToH256(txHash),
	}

	log.Info("Fetching zena block receipt")

	res, err := h.client.ZenaBlockReceipt(ctx, req)
	if err != nil {
		return &ethTypes.Receipt{}, err
	}

	log.Info("Fetched zena block receipt")

	return receiptResponseToTypesReceipt(res.Receipt), nil
}

func receiptResponseToTypesReceipt(receipt *proto.Receipt) *ethTypes.Receipt {
	var receiptType uint8
	if receipt.Type <= math.MaxUint8 {
		receiptType = uint8(receipt.Type)
	} else {
		log.Error("Invalid receipt type", "value", receipt.Type)
		return &ethTypes.Receipt{}
	}
	// Bloom and Logs have been intentionally left out as they are not used in the current implementation
	return &ethTypes.Receipt{
		Type:              receiptType,
		PostState:         receipt.PostState,
		Status:            receipt.Status,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		TxHash:            protoutil.ConvertH256ToHash(receipt.TxHash),
		ContractAddress:   protoutil.ConvertH160toAddress(receipt.ContractAddress),
		GasUsed:           receipt.GasUsed,
		EffectiveGasPrice: big.NewInt(receipt.EffectiveGasPrice),
		BlobGasUsed:       receipt.BlobGasUsed,
		BlobGasPrice:      big.NewInt(receipt.BlobGasPrice),
		BlockHash:         protoutil.ConvertH256ToHash(receipt.BlockHash),
		BlockNumber:       big.NewInt(receipt.BlockNumber),
		TransactionIndex:  uint(receipt.TransactionIndex),
	}
}

func ToBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}
	// It's negative.
	if number.IsInt64() {
		return rpc.BlockNumber(number.Int64()).String()
	}
	// It's negative and large, which is invalid.
	return fmt.Sprintf("<invalid %d>", number)
}
